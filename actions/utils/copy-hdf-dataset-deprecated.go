package utils

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"ras-runner/actions"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const (
	nameField         = "name"
	dataPathField     = "datapath"
	srcDataSourcePath = "hdf"
)

func init() {
	cc.ActionRegistry.RegisterAction("copy-hdf", &CopyHdfDatasetAction{})
}

/*
	copies content of one hdf5 data file into another one that is local
*/

type CopyHdfDatasetAction struct {
	cc.ActionRunnerBase
}

func (a *CopyHdfDatasetAction) Run() error {
	log.Printf("Ready to copy %s\n", a.Action.Description)

	srcconfig, err := a.Action.Attributes.GetMap("src")
	if err != nil {
		return fmt.Errorf("missing src attribute data")
	}

	destconfig, err := a.Action.Attributes.GetMap("dest")
	if err != nil {
		return fmt.Errorf("missing dest attribute data")
	}

	//this type assertion is ugly but since we are stopping on error, a panic is ok
	srcname := srcconfig[nameField].(string)
	srcdatapath := srcconfig[dataPathField].(string)
	dest := destconfig[nameField].(string)
	destdatapath := destconfig[dataPathField].(string)

	src, err := a.PluginManager.GetInputDataSource(srcname)
	if err != nil {
		return fmt.Errorf("error getting input source %s: %s", srcname, err)
	}

	srcstore, err := a.PluginManager.GetStore(src.StoreName)
	if err != nil {
		return fmt.Errorf("error getting input store %s: %s", src.StoreName, err)
	}

	log.Printf("%s::::%s", dest, srcstore)
	log.Printf("finished creating temp for %s\n", a.Action.Description)

	return CopyHdf5Dataset(src.Paths[srcDataSourcePath], srcdatapath, srcstore, dest, destdatapath)
}

func CopyHdf5Dataset(src string, srcdataset string, srcstore *cc.DataStore, dest string, destdataset string) error {
	if srcstore.StoreType == "S3" {
		profile := srcstore.DsProfile
		bucket := os.Getenv(fmt.Sprintf("%s_%s", profile, actions.AWSBUCKET))
		src = fmt.Sprintf(actions.S3BucketTemplate, bucket, srcstore.Parameters["root"], url.QueryEscape(src))
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
