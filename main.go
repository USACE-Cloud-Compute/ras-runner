package main

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
	"strings"

	"github.com/usace/cc-go-sdk"
	hdf5 "github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const (
	MODEL_DIR    = "/sim/model"
	MODEL_SCRIPT = "/ras/run-model.sh"
	GEOM_PREPROC = "/ras/run-geom-preproc.sh"
	RASTIMEPATH  = "Unsteady Time Series/Time"
	AWSBUCKET    = "AWS_S3_BUCKET"
)

func timePath(datapath string) string {
	tsroot := datapath[:strings.Index(datapath, "Unsteady Time Series")]
	return tsroot + RASTIMEPATH
}

var modelPrefix string
var event int

var tolerance float64 = 0.000001

func main() {
	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Fatalf("Unable to initialize the CC plugin manager: %s\n", err)
	}
	for _, action := range pm.GetPayload().Actions {
		fmt.Println(action.Name)
		switch action.Type {
		case "update-breach-bfile":
			// Assumes bFile and fragility curve file  were copied local with the CopyLocal action.
			err := actions.UpdateBfileAction(action, MODEL_DIR)
			if err != nil {
				log.Fatal(err)
			}
		case "create-ras-tmp":
			log.Printf("Ready to create temp for %s\n", action.Name)
			srcname := action.Parameters["src"].(map[string]any)["name"].(string)
			dest := action.Parameters["dest"].(map[string]any)["name"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}

			err = MakeRasHdfTmp(src.Paths[0], srcstore, dest)
			if err != nil {
				log.Println(err)
			}
			log.Printf("Finished creating temp for %s\n", action.Name)
		case "copy-hdf":
			log.Printf("Ready to copy %s\n", action.Name)
			srcname := action.Parameters["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Parameters["src"].(map[string]any)["datapath"].(string)
			dest := action.Parameters["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Parameters["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			log.Printf("%s::::%s", dest, srcstore)
			log.Printf("Finished creating temp for %s\n", action.Name)
			err = CopyHdf5Dataset(src.Paths[0], srcdatapath, srcstore, dest, destdatapath)
		case "refline-to-boundary-condition":
			log.Printf("Updating boundary condition %s\n", action.Name)
			refline := action.Parameters["refline"].(string)
			srcname := action.Parameters["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Parameters["src"].(map[string]any)["datapath"].(string)
			dest := action.Parameters["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Parameters["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			err = MigrateRefLineData(src.Paths[0], srcstore, srcdatapath, dest, destdatapath, refline)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("finished updating boundary condition %s\n", action.Name)
		case "update-boundary-condition":
			log.Printf("Updating boundary condition %s\n", action.Name)
			srcname := action.Parameters["src"].(map[string]any)["name"].(string)
			srcdatapath := action.Parameters["src"].(map[string]any)["datapath"].(string)
			dest := action.Parameters["dest"].(map[string]any)["name"].(string)
			destdatapath := action.Parameters["dest"].(map[string]any)["datapath"].(string)
			src, err := pm.GetInputDataSource(srcname)
			if err != nil {
				log.Fatalf("Error getting input source %s", srcname)
			}
			srcstore, err := pm.GetStore(src.StoreName)
			if err != nil {
				log.Fatalf("Error getting input store %s", src.StoreName)
			}
			err = MigrateBoundaryConditionData(src.Paths[0], srcstore, srcdatapath, dest, destdatapath)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("finished updating boundary condition %s\n", action.Name)
		case "copy-inputs":
			err = fetchInputSourceFiles(pm)
			if err != nil {
				log.Fatalln(err)
			}
		case "unsteady-simulation":
			log.Printf("Running unsteady-simulation: %s", action.Description)
			modelPrefix = pm.GetPayload().Attributes["modelPrefix"].(string)

			plan := pm.GetPayload().Attributes["plan"].(string) //cfile
			geom := pm.GetPayload().Attributes["geom"].(string) //bfile

			out := strings.Builder{}

			if gproc, ok := pm.GetPayload().Attributes["geom_preproc"]; ok {
				runGeomPreproc := gproc.(string)
				if strings.ToLower(runGeomPreproc) == "true" {
					log.Println("Running geometry preprocessor")
					cmdout, err := exec.Command(GEOM_PREPROC, MODEL_DIR, modelPrefix, geom).Output()
					if err != nil {
						log.Fatalf("Error running geometry preprocessor:%s\n", err)
					}
					out.Write([]byte("---------- GEOMETRY PREPROCESSOR --------------"))
					_, err = out.Write(cmdout)
					out.Write([]byte("---------- END GEOMETRY PREPROCESSOR ----------"))
				}
			}

			log.Printf("Running model %s\n", action.Name)
			cmdout, err := exec.Command(MODEL_SCRIPT, MODEL_DIR, modelPrefix, geom, plan).CombinedOutput()
			// grab any log information and write to output location before dealing with any errors
			out.Write([]byte("---------- RAS Model Output --------------"))
			_, err = out.Write(cmdout)
			saveResults(pm, plan, &out)
			// handle the error now....
			if err != nil {
				log.Fatalf("Error running ras model:%s\n", err)
			}

		}
	}
	log.Println("Finished")
}

var RasTmpDatasets []string = []string{"Geometry", "Plan Data", "Event Conditions"}

const s3BucketTemplate = "https://%s.s3.amazonaws.com%s/%s"

func MakeRasHdfTmp(src string, srcstore *cc.DataStore, dest string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], url.QueryEscape(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", MODEL_DIR, dest)
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
	return nil
}

func CopyHdf5Dataset(src string, srcdataset string, srcstore *cc.DataStore, dest string, destdataset string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], url.QueryEscape(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", MODEL_DIR, dest)
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
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, AWSBUCKET))
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

	destpath := fmt.Sprintf("%s/%s", MODEL_DIR, dest)
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
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, AWSBUCKET))
		src = fmt.Sprintf(s3BucketTemplate, bucket, srcstore.Parameters["root"], encodeUrlPath(src))
	}
	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destpath := fmt.Sprintf("%s/%s", MODEL_DIR, dest)
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

func fetchInputSourceFiles(pm *cc.PluginManager) error {
	for _, ds := range pm.GetInputDataSources() {
		err := func() error {
			source, err := pm.FileReader(ds, 0)
			if err != nil {
				return err
			}
			defer source.Close()
			destfile := fmt.Sprintf("%s/%s", MODEL_DIR, filepath.Base(ds.Paths[0]))
			log.Printf("Copying %s to %s\n", ds.Paths[0], destfile)
			destination, err := os.Create(destfile)
			if err != nil {
				return err
			}
			defer destination.Close()
			_, err = io.Copy(destination, source)
			return err
		}()
		if err != nil {
			log.Printf("Error fetching %s", ds.Paths[0])
			return err
		}
	}
	return nil
}

func saveResults(pm *cc.PluginManager, rasplan string, raslog *strings.Builder) error {
	//write plan results
	file := fmt.Sprintf("%s.p%s.tmp.hdf", modelPrefix, rasplan)
	ds, err := pm.GetOutputDataSource(file)
	filepath := fmt.Sprintf("%s/%s", MODEL_DIR, file)
	reader, err := os.Open(filepath)
	if err != nil {
		raslog.WriteString(fmt.Sprintf("Unable to open %s for copying: %s\n", file, err))
	} else {
		defer reader.Close()
		err = pm.FileWriter(reader, ds, 0)
		if err != nil {
			raslog.WriteString(fmt.Sprintf("Unable to copy %s: %s\n", file, err))
		}
	}
	//write log
	ds, err = pm.GetOutputDataSource("rasoutput")
	logReader := strings.NewReader(raslog.String())
	log.Printf("Output log:%s", ds.Paths[0])
	err = pm.FileWriter(logReader, ds, 0)
	return err
}

func writeFile(fExt string, obj io.ReadCloser) error {
	filepath := fmt.Sprintf("%s/%s.%s", MODEL_DIR, modelPrefix, fExt)
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, obj)
	return err
}
