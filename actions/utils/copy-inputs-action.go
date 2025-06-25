package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"ras-runner/actions"

	"github.com/usace/cc-go-sdk"
)

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
