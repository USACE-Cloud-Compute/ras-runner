package utils

import (
	"fmt"
	"log"
	"ras-runner/actions"

	"github.com/usace/cc-go-sdk"
)

type PostOutputsAction struct {
	cc.ActionRunnerBase
}

func (a *PostOutputsAction) Run() error {
	err := postOuptutFiles(a.PluginManager)
	if err != nil {
		return fmt.Errorf("failed to post outputs: %s", err)
	}
	return nil
}

func postOuptutFiles(pm *cc.PluginManager) error {
	//this code is intended to be updated in the future to be more clean, for now it is structured to work without changing any previous actions, and is written with an out of date sdk to support multiple project.s
	modelPrefix := pm.Payload.Attributes["modelPrefix"].(string)
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
