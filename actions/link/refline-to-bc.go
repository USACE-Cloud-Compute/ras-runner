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

func init() {
	cc.ActionRegistry.RegisterAction("refline-to-boundary-condition", &ReflineToBc{})
}

// refline to boundary condition
type ReflineToBc struct {
	cc.ActionRunnerBase
}

func (a *ReflineToBc) Run() error {
	//@TODO need string length
	log.Printf("Updating refline to boundary condition %s\n", a.Action.Description)
	refline := a.Action.Attributes["refline"].(string)
	srcname := a.Action.Attributes["src"].(map[string]any)["name"].(string)
	srcdatapath := a.Action.Attributes["src"].(map[string]any)["datapath"].(string)
	dest := a.Action.Attributes["dest"].(map[string]any)["name"].(string)
	destdatapath := a.Action.Attributes["dest"].(map[string]any)["datapath"].(string)

	src, err := a.PluginManager.GetInputDataSource(srcname)
	if err != nil {
		return fmt.Errorf("error getting input source %s: %s", srcname, err)
	}

	srcstore, err := a.PluginManager.GetStore(src.StoreName)
	if err != nil {
		return fmt.Errorf("error getting input store %s: %s", src.StoreName, err)
	}

	err = MigrateRefLineData(src.Paths["0"], srcstore, srcdatapath, dest, destdatapath, refline)
	if err != nil {
		return fmt.Errorf("failed to migrate refline data: %s", err)
	}

	log.Printf("finished updating boundary condition %s\n", a.Action.Description)

	return nil
}

func MigrateRefLineData(src string, srcstore *cc.DataStore, src_datapath string, dest string, dest_datapath string, refline string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		src = fmt.Sprintf(actions.S3BucketTemplate, bucket, srcstore.Parameters["root"], actions.EncodeUrlPath(src))
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
