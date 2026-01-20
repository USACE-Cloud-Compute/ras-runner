package actions

import (
	"fmt"
	"log"
	"os"
	"ras-runner/actions"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
)

const (
	srcDataSourcePath = "hdf"
)

func init() {
	cc.ActionRegistry.RegisterAction("hdf-to-hdf", &HdftoHdfDatasetAction{})
}

/*
	copies content of one hdf5 data file into another one that is local
*/

type HdftoHdfDatasetAction struct {
	cc.ActionRunnerBase
}

func (a *HdftoHdfDatasetAction) Run() error {
	log.Printf("Ready to copy %s\n", a.Action.Description)

	src, err := a.Action.GetInputDataSource("src")
	if err != nil {
		return fmt.Errorf("missing input datasource named src")
	}

	dest, err := a.Action.GetOutputDataSource("dest")
	if err != nil {
		return fmt.Errorf("missing output source named dest")
	}

	if len(src.DataPaths) != len(dest.DataPaths) {
		return fmt.Errorf("src and dest datapath lengths do not match")
	}

	for srckey, srcdatapath := range src.DataPaths {
		err = CopyHdf5Dataset(src.Paths["hdf"], srcdatapath, dest.Paths["hdf"], dest.DataPaths[srckey])
		if err != nil {
			return fmt.Errorf("error copying from src to dest %s: %s", srcdatapath, dest.DataPaths[srckey])
		}
	}
	return nil
}

func CopyHdf5Dataset(src string, srcdataset string, dest string, destdataset string) error {

	srcfile, err := hdf5.OpenFile(src, hdf5.F_ACC_RDWR)
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
