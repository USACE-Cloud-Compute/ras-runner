package main

// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"ras-runner/actions"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/usace/cc-go-sdk"

	hdf5 "github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

//@TODO fix action description logging back to action.Name

func timePath(datapath string) string {
	tsroot := datapath[:strings.Index(datapath, "Unsteady Time Series")]
	return tsroot + actions.RASTIMEPATH
}

var modelPrefix string
var event int

// this is the tolerance we will use when comparing float64 values for comparison
// specifically it is used to compare RAS time values
var tolerance float64 = 0.000001

func main() {
	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Fatalf("unable to initialize the CC plugin manager: %s\n", err)
	}
	pm.RunActions()
	log.Println("Finished")
}

/*
*























 */

func actionSwitch(pm *cc.PluginManager) {
	var err error
	scriptPath := os.Getenv("RAS_SCRIPT_PATH")
	if scriptPath == "" {
		scriptPath = actions.MODEL_SCRIPT_PATH
	}

	for _, action := range pm.Payload.Actions {
		switch action.Type {

		case "update-breach-bfile":

			// runner:=actions.UpdateBfileAction2{
			// 	//PluginManager: pm,
			// 	Action: action,
			// 	ModelDir: MODEL_DIR,
			// }
			// err:=runner.Run()

			// Assumes bFile and fragility curve file  were copied local with the CopyLocal action.
			err := actions.UpdateBfileAction(action, actions.MODEL_DIR)
			if err != nil {
				pm.Logger.Fatal(err.Error())
			}
		case "update-outlet-ts-bfile":
			// Assumes bFile and pxx.tmp.hdf file  were copied local with the CopyLocal action.
			err := actions.UpdateOutletTSAction(action, actions.MODEL_DIR)
			if err != nil {
				log.Fatal(err)
			}
		case "create-ras-tmp":
			log.Printf("Ready to create temp for %s\n", action.Description)
			srcname := action.Attributes.GetStringOrFail("src")                               //.Parameters["src"].(map[string]any)["name"].(string)
			local_dest := action.Attributes.GetStringOrFail("local_dest")                     //.(map[string]any)["name"].(string)
			save_to_remote := action.Attributes.GetStringOrDefault("save_to_remote", "false") //].(map[string]any)["name"].(string)
			remote_dest_name := action.Attributes.GetStringOrDefault("remote_dest", "")
			/*src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}*/
			saveRemotely, err := strconv.ParseBool(save_to_remote)
			if err != nil {
				log.Fatal("could not parse save_to_remote to bool")
			}
			/*
				srcstore, err := pm.GetStore(src.StoreName)
				if err != nil {
					log.Fatalf("Error getting input store %s", src.StoreName)
				}*/
			src_local_path := fmt.Sprintf("%s/%s", actions.MODEL_DIR, srcname)
			err = MakeRasHdfTmp(src_local_path, local_dest)
			if err != nil {
				log.Println(err)
			}
			if saveRemotely {
				if remote_dest_name == "" {
					log.Fatal("user requested to save tmp file remotely but provided no output destination name")
				}
				//need to make sure the dest file is closed.
				//get the bytes of the dest file and push them to the dest datasource.
				destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, local_dest)

				pm.CopyFileToRemote(cc.CopyFileToRemoteInput{
					LocalPath:    destpath,
					RemoteDsName: remote_dest_name,
					DsPathKey:    "0",
				})
			}
			log.Printf("Finished creating temp for %s\n", action.Description)
		case "copy-hdf":
			log.Printf("Ready to copy %s\n", action.Description)
			srcname := action.Attributes["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Attributes["src"].(map[string]any)["datapath"].(string)
			dest := action.Attributes["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Attributes["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			log.Printf("%s::::%s", dest, srcstore)
			log.Printf("Finished creating temp for %s\n", action.Description)
			err = CopyHdf5Dataset(src.Paths["0"], srcdatapath, srcstore, dest, destdatapath)
		case "refline-to-boundary-condition":
			//@TODO need string length
			log.Printf("Updating boundary condition %s\n", action.Description)
			refline := action.Attributes["refline"].(string)
			srcname := action.Attributes["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Attributes["src"].(map[string]any)["datapath"].(string)
			dest := action.Attributes["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Attributes["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			err = MigrateRefLineData(src.Paths["0"], srcstore, srcdatapath, dest, destdatapath, refline)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("finished updating boundary condition %s\n", action.Description)
		case "update-boundary-condition":
			log.Printf("Updating boundary condition %s\n", action.Description)
			srcname := action.Attributes["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Attributes["src"].(map[string]any)["datapath"].(string)
			dest := action.Attributes["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Attributes["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			err = MigrateBoundaryConditionData(src.Paths["0"], srcstore, srcdatapath, dest, destdatapath)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("finished updating boundary condition %s\n", action.Description)
		case "column-to-boundary-condition":
			log.Printf("Updating boundary condition %s\n", action.Description)
			column_index := action.Attributes["column-index"].(string)
			readcol, err := strconv.Atoi(column_index)
			if err != nil {
				log.Fatalf("Invalid column index: %s\n", column_index)
			}
			srcname := action.Attributes["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Attributes["src"].(map[string]any)["datapath"].(string)
			dest := action.Attributes["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Attributes["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			err = MigrateColumnData(src.Paths["0"], srcstore, srcdatapath, dest, destdatapath, readcol)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("finished updating boundary condition %s\n", action.Description)
		// case "copy-inputs":
		// 	err = fetchInputSourceFiles(pm)
		// 	if err != nil {
		// 		pm.Logger.Fatal(err.Error())
		// 	}

		/*
			case "bcline-peak-outputs":
				err = actions.ReadBCLinePeak(action)
				if err != nil {
					log.Fatalln(err)
				}
			case "refline-peak-outputs":
				err = actions.ReadRefLinePeak(action)
				if err != nil {
					log.Fatalln(err)
				}
			case "refpoint-peak-outputs":
				err = actions.ReadRefPointPeak(action)
				if err != nil {
					log.Fatalln(err)
				}
			case "refpoint-min-outputs":
				err = actions.ReadRefPointMinimum(action)
				if err != nil {
					log.Fatalln(err)
				}
			case "simulation-attribute-metadata":
				err = actions.ReadSimulationMetadata(action)
				if err != nil {
					log.Fatalln(err)
				}
			case "structure-variables-peak-output":
				err = actions.ReadStructureVariablesPeak(action)
				if err != nil {
					log.Fatalln(err)
				}
		*/
		case "post-outputs":
			//this code is a short term fix to allow for more flexibility in this plugin to push things out to an output.
			//ultimately with the updated sdk's this would change to leverage action level inputs or outputs (which currently do not exist in this version of the sdk...)
			err := postOuptutFiles(pm)
			if err != nil {
				log.Fatalln(err)
			}
		//case "post-all-changed-files"
		//use the hashes from the pull from s3 on copy inputs, and check local files to see if there are any changes, and then push any changed files.
		case "unsteady-simulation":
			log.Printf("Running unsteady-simulation: %s", action.Description)
			modelPrefix = pm.Payload.Attributes["modelPrefix"].(string)

			plan := pm.Payload.Attributes["plan"].(string) //cfile
			geom := pm.Payload.Attributes["geom"].(string) //bfile

			out := strings.Builder{}

			if gproc, ok := pm.Payload.Attributes["geom_preproc"]; ok {
				runGeomPreproc := gproc.(string)
				if strings.ToLower(runGeomPreproc) == "true" {
					gppcmd := fmt.Sprintf("%s/%s", scriptPath, actions.GEOM_PREPROC)
					log.Printf("Running geometry preprocessor: %s %s %s %s\n", gppcmd, actions.MODEL_DIR, modelPrefix, geom)
					cmdout, err := exec.Command(gppcmd, actions.MODEL_DIR, modelPrefix, geom).Output()
					if err != nil {
						log.Fatalf("Error running geometry preprocessor:%s\n", err)
					}
					out.Write([]byte("---------- GEOMETRY PREPROCESSOR --------------"))
					_, err = out.Write(cmdout)
					out.Write([]byte("---------- END GEOMETRY PREPROCESSOR ----------"))
				}
			}

			log.Printf("Running model %s\n", action.Description)
			simcmd := fmt.Sprintf("%s/%s", scriptPath, actions.MODEL_SCRIPT)
			log.Printf("Running model script: %s %s %s %s %s\n", simcmd, actions.MODEL_DIR, modelPrefix, geom, plan)
			cmdout, err := exec.Command(simcmd, actions.MODEL_DIR, modelPrefix, geom, plan).CombinedOutput()
			// grab any log information and write to output location before dealing with any errors
			out.Write([]byte("---------- RAS Model Output --------------"))
			_, err = out.Write(cmdout)
			saveResults(pm, plan, &out)
			// handle the error now....
			if err != nil {
				log.Fatalf("Error running ras model:%s\n", err)
			}
		default:

			log.Fatalln(action.Description + " not found")

		}

	}
}

var RasTmpDatasets []string = []string{"Geometry", "Plan Data", "Event Conditions"}

const s3BucketTemplate = "https://%s.s3.amazonaws.com%s/%s"

func MakeRasHdfTmp(src string, local_dest string) error {
	srcfile, err := hdf5.OpenFile(src, hdf5.F_ACC_RDONLY)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, local_dest)
	_, err = os.Stat(destpath)

	var destfile *hdf5.File

	if os.IsNotExist(err) {
		destfile, err = hdf5.CreateFile(destpath, hdf5.F_ACC_EXCL)
		if err != nil {
			return err
		}
		defer destfile.Close()
	} else {
		destfile, err = hdf5.OpenFile(destpath, hdf5.F_ACC_RDWR)
		if err != nil {
			return err
		}
		defer destfile.Close()
	}
	for _, v := range RasTmpDatasets {
		err := srcfile.CopyTo(v, destfile, v)
		if err != nil {
			return err
		}
	}
	scalar, err := hdf5.CreateDataspace(hdf5.S_SCALAR)
	if err != nil {
		return err
	}
	defer scalar.Close()
	attrs := []string{"File Type", "File Version", "Projection", "Units System"}
	vals := make([]string, 4)
	srcrootgroup, err := srcfile.OpenGroup("/")
	if err != nil {
		return err
	}
	defer srcrootgroup.Close()
	destrootgroup, err := destfile.OpenGroup("/")
	if err != nil {
		return err
	}
	defer destrootgroup.Close()

	for i, v := range attrs {
		err := func() error {
			if srcrootgroup.AttributeExists(v) {
				attr, err := srcrootgroup.OpenAttribute(v)
				if err != nil {
					return err
				}
				defer attr.Close()

				var attrdata string
				/*dt, err := hdf5.T_C_S1.Copy()
				if err != nil {
					return err
				}*/
				attr.Read(&attrdata, hdf5.T_GO_STRING) //strong limiting restriction acceptable given our use case.
				vals[i] = string(attrdata)
				sdt, err := hdf5.T_C_S1.Copy()
				if err != nil {
					return err
				}

				err = sdt.SetSize(len(attrdata))
				//fmt.Println(sdt.Size())
				//fmt.Println(attrdata)
				if err != nil {
					return err
				}

				destattribute, err := destrootgroup.CreateAttribute(v, sdt, scalar)
				if err != nil {
					return err
				}
				defer destattribute.Close()
				cstring := C.CString(attrdata)
				defer C.free(unsafe.Pointer(cstring))
				err = destattribute.Write(cstring, sdt)
				if err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func CopyHdf5Dataset(src string, srcdataset string, srcstore *cc.DataStore, dest string, destdataset string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], url.QueryEscape(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, dest)
	_, err = os.Stat(destpath)

	var destfile *hdf5.File

	if os.IsNotExist(err) {
		destfile, err = hdf5.CreateFile(destpath, hdf5.F_ACC_EXCL)
		if err != nil {
			return err
		}
		defer destfile.Close()
	} else {
		destfile, err = hdf5.OpenFile(destpath, hdf5.F_ACC_RDWR)
		if err != nil {
			return err
		}
		defer destfile.Close()
	}

	err = srcfile.CopyTo(srcdataset, destfile, destdataset)
	if err != nil {
		return err
	}
	return nil
}

func getRowVal2(srcVals *hdf5utils.HdfDataset, srcTimes *hdf5utils.HdfDataset, timeval float32, readcol int) (float32, error) {
	numcols := srcVals.Dims()[1]
	//
	srcdata := make([]float32, numcols)
	srctime := make([]float64, 1)

	if timeval <= 0.0 {
		err := srcVals.ReadRow(0, &srcdata)
		if err != nil {
			return 0, err
		}
		return srcdata[0], nil
	}

	for i := 0; i < srcVals.Rows(); i++ {
		err := srcVals.ReadRow(i, &srcdata)
		if err != nil {
			return 0, err
		}
		err = srcTimes.ReadRow(i, &srctime)
		if err != nil {
			return 0, err
		}

		if math.Abs(float64(timeval)-srctime[0]) < tolerance {
			return srcdata[readcol-1], nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Unable to find corresponding input source record for time %f", timeval))
}

func getRowVal(srcVals *hdf5utils.HdfDataset, srcTimes *hdf5utils.HdfDataset, timeval float32) (float32, error) {
	srcdata := make([]float32, 5)
	srctime := make([]float64, 1)

	if timeval <= 0.0 {
		err := srcVals.ReadRow(0, &srcdata)
		if err != nil {
			return 0, err
		}
		return srcdata[0], nil
	}

	for i := 0; i < srcVals.Rows(); i++ {
		err := srcVals.ReadRow(i, &srcdata)
		if err != nil {
			return 0, err
		}
		err = srcTimes.ReadRow(i, &srctime)
		if err != nil {
			return 0, err
		}

		if math.Abs(float64(timeval)-srctime[0]) < tolerance {
			return srcdata[0], nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Unable to find corresponding input source record for time %f", timeval))
}

func encodeUrlPath(src string) string {
	srcvals := strings.Split(src, "/")
	srcencoded := strings.Builder{}
	for i, sv := range srcvals {
		if i == 0 {
			srcencoded.WriteString(url.PathEscape(sv))
		} else {
			srcencoded.WriteString("/" + url.PathEscape(sv))
		}
	}
	return srcencoded.String()
}

func MigrateRefLineData(src string, srcstore *cc.DataStore, src_datapath string, dest string, dest_datapath string, refline string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], encodeUrlPath(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	srcTime, err := hdf5utils.NewHdfDataset(timePath(src_datapath), hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float64,
		File:         srcfile,
		ReadOnCreate: true,
	})

	if err != nil {
		return err
	}
	defer srcTime.Close()

	//get the reference line flow dataset
	refLineVals, err := hdf5utils.NewHdfDataset(src_datapath+"/Flow", hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float32,
		File:         srcfile,
		ReadOnCreate: true,
	})
	if err != nil {
		return err
	}
	defer refLineVals.Close()

	//get the reference line positions
	refLineNames, err := hdf5utils.NewHdfDataset(src_datapath+"/Name", hdf5utils.HdfReadOptions{
		Dtype:        reflect.String,
		Strsizes:     hdf5utils.NewHdfStrSet(43),
		File:         srcfile,
		ReadOnCreate: true,
	})
	if err != nil {
		return err
	}
	defer refLineNames.Close()

	refLineColumnIndex := -1
	for i := 0; i < refLineNames.Rows(); i++ {
		name := []string{}
		err := refLineNames.ReadRow(i, &name)
		if err != nil || len(name) == 0 {
			return errors.New("Error reading Reference Line Names")
		}

		if refline == name[0] {
			refLineColumnIndex = i
		}
	}
	if refLineColumnIndex < 0 {
		return errors.New(fmt.Sprintf("Invalid Reference Line: %s\n", refline))
	}

	destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, dest)
	_, err = os.Stat(destpath)

	var destfile *hdf5.File

	destfile, err = hdf5.OpenFile(destpath, hdf5.F_ACC_RDWR)
	if err != nil {
		return err
	}
	defer destfile.Close()

	//get a copy of the destination data
	var destVals *hdf5utils.HdfDataset

	err = func() error {
		destoptions := hdf5utils.HdfReadOptions{
			Dtype:        reflect.Float32,
			File:         destfile,
			ReadOnCreate: true,
		}
		destVals, err = hdf5utils.NewHdfDataset(dest_datapath, destoptions)
		if err != nil {
			return err
		}
		defer destVals.Close()
		return nil
	}()
	if err != nil {
		return err
	}

	//create a new dataset
	boundaryConditionData := make([]float32, destVals.Rows()*2)

	for i := 0; i < destVals.Rows(); i++ {
		refLineRow := []float32{}
		err = refLineVals.ReadRow(i, &refLineRow)
		if err != nil {
			return err
		}
		destRow := []float32{}
		err = destVals.ReadRow(i, &destRow)
		if err != nil {
			return err
		}

		boundaryConditionData[i*2] = destRow[0]
		boundaryConditionData[i*2+1] = refLineRow[refLineColumnIndex]
	}
	//write the new boundary condition buffer back to the destiation dataset
	destWriter, err := destfile.OpenDataset(dest_datapath)
	if err != nil {
		return err
	}
	defer destWriter.Close()
	return destWriter.Write(&boundaryConditionData)
}

func MigrateBoundaryConditionData(src string, srcstore *cc.DataStore, src_datapath string, dest string, dest_datapath string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], encodeUrlPath(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, dest)
	_, err = os.Stat(destpath)

	var destfile *hdf5.File

	destfile, err = hdf5.OpenFile(destpath, hdf5.F_ACC_RDWR)
	if err != nil {
		return err
	}
	defer destfile.Close()

	//Get the data values from the source file
	//this is the RAS model output
	options := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float32,
		File:         srcfile,
		ReadOnCreate: true,
	}

	srcVals, err := hdf5utils.NewHdfDataset(src_datapath, options)
	if err != nil {
		return err
	}
	defer srcVals.Close()

	//Get the times corresponding to the source file values

	tsoptions := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float64,
		File:         srcfile,
		ReadOnCreate: true,
	}

	srcTime, err := hdf5utils.NewHdfDataset(timePath(src_datapath), tsoptions)
	if err != nil {
		return err
	}
	defer srcTime.Close()

	//Get a copy of the destination dataset
	var destVals *hdf5utils.HdfDataset

	err = func() error {
		destoptions := hdf5utils.HdfReadOptions{
			Dtype:        reflect.Float32,
			File:         destfile,
			ReadOnCreate: true,
		}
		destVals, err = hdf5utils.NewHdfDataset(dest_datapath, destoptions)
		if err != nil {
			return err
		}
		defer destVals.Close()
		return nil
	}()
	if err != nil {
		return err
	}

	//create a new buffer with mutated boundary conditions
	boundaryConditionData := make([]float32, destVals.Rows()*2)

	for i := 0; i < destVals.Rows(); i++ {

		destRow := make([]float32, 2)
		err := destVals.ReadRow(i, &destRow)
		if err != nil {
			return err
		}

		val, err := getRowVal(srcVals, srcTime, destRow[0])
		if err != nil {
			return err
		}

		boundaryConditionData[i*2] = destRow[0]
		boundaryConditionData[i*2+1] = val
	}

	//write the new boundary condition buffer back to the destiation dataset
	destWriter, err := destfile.OpenDataset(dest_datapath)
	if err != nil {
		return err
	}
	defer destWriter.Close()
	err = destWriter.Write(&boundaryConditionData)
	if err != nil {
		return err
	}
	return nil
}
func MigrateColumnData(src string, srcstore *cc.DataStore, src_datapath string, dest string, dest_datapath string, readcol int) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], encodeUrlPath(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, dest)
	_, err = os.Stat(destpath)

	var destfile *hdf5.File

	destfile, err = hdf5.OpenFile(destpath, hdf5.F_ACC_RDWR)
	if err != nil {
		return err
	}
	defer destfile.Close()

	//Get the data values from the source file
	//this is the RAS model output
	options := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float32,
		File:         srcfile,
		ReadOnCreate: true,
	}

	srcVals, err := hdf5utils.NewHdfDataset(src_datapath, options)
	if err != nil {
		return err
	}
	defer srcVals.Close()

	//Get the times corresponding to the source file values

	tsoptions := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float64,
		File:         srcfile,
		ReadOnCreate: true,
	}

	srcTime, err := hdf5utils.NewHdfDataset(timePath(src_datapath), tsoptions)
	if err != nil {
		return err
	}
	defer srcTime.Close()

	//Get a copy of the destination dataset
	var destVals *hdf5utils.HdfDataset

	err = func() error {
		destoptions := hdf5utils.HdfReadOptions{
			Dtype:        reflect.Float32,
			File:         destfile,
			ReadOnCreate: true,
		}
		destVals, err = hdf5utils.NewHdfDataset(dest_datapath, destoptions)
		if err != nil {
			return err
		}
		defer destVals.Close()
		return nil
	}()
	if err != nil {
		return err
	}

	//create a new buffer with mutated boundary conditions
	boundaryConditionData := make([]float32, destVals.Rows()*2)

	for i := 0; i < destVals.Rows(); i++ {

		destRow := make([]float32, 2)
		err := destVals.ReadRow(i, &destRow)
		if err != nil {
			return err
		}

		val, err := getRowVal2(srcVals, srcTime, destRow[0], readcol)
		if err != nil {
			return err
		}

		boundaryConditionData[i*2] = destRow[0]
		boundaryConditionData[i*2+1] = val
	}

	//write the new boundary condition buffer back to the destiation dataset
	destWriter, err := destfile.OpenDataset(dest_datapath)
	if err != nil {
		return err
	}
	defer destWriter.Close()
	err = destWriter.Write(&boundaryConditionData)
	if err != nil {
		return err
	}
	return nil
}

// //////////////

func init() {
	//cc.ActionRegistry.RegisterAction(&CopyInputsAction{ActionRunnerBase: cc.ActionRunnerBase{ActionName: "copy-inputs"}})
	action := &CopyInputsAction{}
	action.SetName("copy-inputs")
	cc.ActionRegistry.RegisterAction(action)
}

type CopyInputsAction struct {
	cc.ActionRunnerBase
}

func (ca *CopyInputsAction) Run() error {
	for _, ds := range ca.PluginManager.Inputs {
		err := func() error {
			source, err := ca.PluginManager.GetReader(cc.DataSourceOpInput{
				DataSourceName: ds.Name,
				PathKey:        "0",
			})
			if err != nil {
				return err
			}
			defer source.Close()
			destfile := fmt.Sprintf("%s/%s", actions.MODEL_DIR, filepath.Base(ds.Paths["0"]))
			log.Printf("Copying %s to %s\n", ds.Paths["0"], destfile)
			destination, err := os.Create(destfile)
			if err != nil {
				return err
			}
			defer destination.Close()
			_, err = io.Copy(destination, source)
			return err
		}()
		if err != nil {
			log.Printf("Error fetching %s", ds.Paths["0"])
			return err
		}
	}
	return nil
}

////////////////

func fetchInputSourceFiles(pm *cc.PluginManager) error {
	for _, ds := range pm.Inputs {
		err := func() error {
			source, err := pm.GetReader(cc.DataSourceOpInput{
				DataSourceName: ds.Name,
				PathKey:        "0",
			})
			if err != nil {
				return err
			}
			defer source.Close()
			destfile := fmt.Sprintf("%s/%s", actions.MODEL_DIR, filepath.Base(ds.Paths["0"]))
			log.Printf("Copying %s to %s\n", ds.Paths["0"], destfile)
			destination, err := os.Create(destfile)
			if err != nil {
				return err
			}
			defer destination.Close()
			_, err = io.Copy(destination, source)
			return err
		}()
		if err != nil {
			log.Printf("Error fetching %s", ds.Paths["0"])
			return err
		}
	}
	return nil
}
func postOuptutFiles(pm *cc.PluginManager) error {
	//this code is intended to be updated in the future to be more clean, for now it is structured to work without changing any previous actions, and is written with an out of date sdk to support multiple project.s
	modelPrefix = pm.Payload.Attributes["modelPrefix"].(string)
	plan := pm.Payload.Attributes["plan"].(string)
	reservedfilename := fmt.Sprintf("%s.p%s.tmp.hdf", modelPrefix, plan)
	for _, ds := range pm.Outputs {
		err := func() error {
			//check if the datasource name is reserved// rasoutput or pxx.tmp.hdf -> ignore these two.
			if ds.Name == "rasoutput" {
				return nil
			}

			if ds.Name == reservedfilename {
				return nil
			}
			//get the local file from the datasource name.
			return pm.CopyFileToRemote(cc.CopyFileToRemoteInput{
				LocalPath:       fmt.Sprintf("%s/%s", actions.MODEL_DIR, ds.Name),
				RemoteStoreName: ds.Name,
				RemotePath:      "0",
			})
		}()
		if err != nil {
			log.Printf("Error fetching %s", ds.Paths["0"])
			return err
		}
	}
	return nil
}
func saveResults(pm *cc.PluginManager, rasplan string, raslog *strings.Builder) error {
	//write plan results
	file := fmt.Sprintf("%s.p%s.tmp.hdf", modelPrefix, rasplan)
	ds, err := pm.GetOutputDataSource(file)
	filepath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, file)
	reader, err := os.Open(filepath)
	if err != nil {
		raslog.WriteString(fmt.Sprintf("Unable to open %s for copying: %s\n", file, err))
	} else {
		defer reader.Close()
		//err = pm.FileWriter(reader, ds, 0)
		_, err = pm.Put(cc.PutOpInput{
			SrcReader: reader,
			DataSourceOpInput: cc.DataSourceOpInput{
				DataSourceName: ds.Name,
				PathKey:        "0",
			},
		})
		if err != nil {
			raslog.WriteString(fmt.Sprintf("Unable to copy %s: %s\n", file, err))
		}
	}
	//write log
	ds, err = pm.GetOutputDataSource("rasoutput")
	if err != nil {
		return err
	}
	logReader := strings.NewReader(raslog.String())
	log.Printf("Output log:%s", ds.Paths["0"])
	//err = pm.FileWriter(logReader, ds, 0)
	_, err = pm.Put(cc.PutOpInput{
		SrcReader: logReader,
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: ds.Name,
			PathKey:        "0",
		},
	})
	return err
}

func writeFile(fExt string, obj io.ReadCloser) error {
	filepath := fmt.Sprintf("%s/%s.%s", actions.MODEL_DIR, modelPrefix, fExt)
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, obj)
	return err
}
