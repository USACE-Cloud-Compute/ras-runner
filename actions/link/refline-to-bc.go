package actions

import (
	"errors"
	"fmt"
	"log"
	"os"
	"ras-runner/actions"
	"reflect"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const (
	reflineAttrName = "refline"
)

func init() {
	cc.ActionRegistry.RegisterAction("refline-to-boundary-condition", &ReflineToBc{})
}

// ReflineToBc reads reference line data from HDF5 RAS output files and writes it to boundary condition datasets in HDF5 RAS input files.
//
// This action facilitates the transfer of reference line flow data from RAS output results to boundary conditions in RAS input models.
// Unlike the column-to-boundary-condition action, this action assumes that the time arrays in source and destination datasets are identical
// and does not perform time matching between datasets.
type ReflineToBc struct {
	cc.ActionRunnerBase
}

// Run executes the refline-to-boundary-condition action.
//
// It performs the following steps:
// 1. Retrieves the reference line name from action attributes
// 2. Gets source and destination data sources from IOManager
// 3. Calls MigrateRefLineData to process the data transfer
// 4. Returns any errors encountered during processing
//
// The action requires:
// - "refline" attribute specifying which reference line to extract
// - "source" configuration with name and datapath for input data
// - "destination" configuration with name and datapath for output data
func (a *ReflineToBc) Run() error {

	//@TODO need string length
	log.Printf("Updating refline to boundary condition %s\n", a.Action.Description)
	refline := a.Action.Attributes.GetStringOrFail(reflineAttrName)

	src, err := a.Action.IOManager.GetInputDataSource("source")
	if err != nil {
		return fmt.Errorf("error getting input source %s: %s", "source", err)
	}

	srcstore, err := a.Action.IOManager.GetStore(src.StoreName)
	if err != nil {
		return fmt.Errorf("error getting input store %s: %s", src.StoreName, err)
	}

	dest, err := a.Action.IOManager.GetOutputDataSource("destination")
	if err != nil {
		return fmt.Errorf("error getting input source %s: %s", "source", err)
	}

	err = MigrateRefLineData(src.Paths["hdf"], srcstore, src.DataPaths["refline"], dest.Paths["hdf"], dest.DataPaths["bcline"], refline)
	if err != nil {
		return fmt.Errorf("failed to migrate refline data: %s", err)
	}

	log.Printf("finished updating boundary condition %s\n", a.Action.Description)

	return nil
}

// MigrateRefLineData transfers reference line flow data from source HDF5 file to destination boundary condition dataset.
//
// This function performs the core data migration logic:
// 1. Handles S3 store type by constructing proper AWS S3 URL template
// 2. Opens source HDF5 file and reads time, flow, and name datasets
// 3. Locates the specified reference line column index
// 4. Opens destination HDF5 file
// 5. Reads existing boundary condition data
// 6. Combines destination time values with reference line flow data
// 7. Writes updated boundary condition data back to destination
//
// Parameters:
//   - src: path to source HDF5 file
//   - srcstore: data store configuration for source
//   - src_datapath: path to reference line dataset in source file
//   - dest: path to destination HDF5 file
//   - dest_datapath: path to boundary condition dataset in destination file
//   - refline: name of reference line to extract data from
//
// Returns error if any step fails during data processing or file operations.
func MigrateRefLineData(src string, srcstore *cc.DataStore, src_datapath string, dest string, dest_datapath string, refline string) error {

	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		template := os.Getenv("HDF_AWS_S3_TEMPLATE")
		src = fmt.Sprintf(template, bucket, srcstore.Parameters["root"], actions.EncodeUrlPath(src))
	}

	srcfile, err := hdf5utils.OpenFile(src, srcstore.DsProfile)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	srcTime, err := hdf5utils.NewHdfDataset(actions.TimePath(src_datapath), hdf5utils.HdfReadOptions{
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
	mt := hdf5utils.DatasetMetadata
	attr, err := hdf5utils.GetAttrMetadata(srcfile, mt, src_datapath+"/Name", "")
	if err != nil {
		return err
	}

	refLineNames, err := hdf5utils.NewHdfDataset(src_datapath+"/Name", hdf5utils.HdfReadOptions{
		Dtype:        reflect.String,
		Strsizes:     hdf5utils.NewHdfStrSet(int(attr.AttrSize)),
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
			return errors.New("error reading reference line Names")
		}

		if refline == name[0] {
			refLineColumnIndex = i
		}
	}
	if refLineColumnIndex < 0 {
		return fmt.Errorf("invalid reference line: %s", refline)
	}

	destpath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, dest)
	_, err = os.Stat(destpath)
	if err != nil {
		return err
	}

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
