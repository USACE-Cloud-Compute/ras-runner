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

func init() {
	cc.ActionRegistry.RegisterAction(&CopyHdfDataset{
		ActionRunnerBase: cc.ActionRunnerBase{ActionName: "copy-hdf"},
	})
}

type CopyHdfDataset struct {
	cc.ActionRunnerBase
}

func (a *CopyHdfDataset) Run() error {
	log.Printf("Ready to copy %s\n", a.Action.Description)

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

	log.Printf("%s::::%s", dest, srcstore)
	log.Printf("finished creating temp for %s\n", a.Action.Description)

	return CopyHdf5Dataset(src.Paths["0"], srcdatapath, srcstore, dest, destdatapath)
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
