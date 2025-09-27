package run

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"ras-runner/actions"
	"strings"

	"github.com/usace/cc-go-sdk"
)

func init() {
	cc.ActionRegistry.RegisterAction("unsteady-simulation", &UnsteadySimulationAction{})
}

type UnsteadySimulationAction struct {
	cc.ActionRunnerBase
}

func (a UnsteadySimulationAction) Run() error {
	log.Printf("Running unsteady-simulation: %s", a.Action.Description)

	scriptPath := os.Getenv(actions.RAS_SCRIPT_PATH_ENV)
	if scriptPath == "" {
		scriptPath = actions.MODEL_SCRIPT_PATH
	}

	modelPrefix := a.PluginManager.Attributes.GetStringOrFail("modelPrefix")
	plan := a.PluginManager.Attributes.GetStringOrFail("plan") //cfile
	geom := a.PluginManager.Attributes.GetStringOrFail("geom") //bfile

	out := strings.Builder{}

	if gproc, ok := a.PluginManager.Attributes["geom_preproc"]; ok {
		runGeomPreproc := gproc.(string)
		if strings.ToLower(runGeomPreproc) == "true" {
			gppcmd := fmt.Sprintf("%s/%s", scriptPath, actions.GEOM_PREPROC)
			log.Printf("Running geometry preprocessor: %s %s %s %s\n", gppcmd, actions.MODEL_DIR, modelPrefix, geom)
			cmdout, err := exec.Command(gppcmd, actions.MODEL_DIR, modelPrefix, geom).Output()
			if err != nil {
				return fmt.Errorf("error running geometry preprocessor:%s", err)
			}
			out.Write([]byte("---------- GEOMETRY PREPROCESSOR --------------"))
			_, err = out.Write(cmdout)
			if err != nil {
				return err
			}
			out.Write([]byte("---------- END GEOMETRY PREPROCESSOR ----------"))
		}
	}

	log.Printf("Running model %s\n", a.Action.Description)
	simcmd := fmt.Sprintf("%s/%s", scriptPath, actions.MODEL_SCRIPT)
	log.Printf("Running model script: %s %s %s %s %s\n", simcmd, actions.MODEL_DIR, modelPrefix, plan, geom)
	cmdout, err := exec.Command(simcmd, actions.MODEL_DIR, modelPrefix, plan, geom).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run: %s", err)
	}
	// grab any log information and write to output location before dealing with any errors
	out.Write([]byte("---------- RAS Model Output --------------"))
	_, err = out.Write(cmdout)

	saveResults(a.PluginManager, modelPrefix, plan, &out)
	if err != nil {
		return fmt.Errorf("failed to save the results: %s", err)
	}
	return nil
}

func saveResults(pm *cc.PluginManager, modelPrefix string, rasplan string, raslog *strings.Builder) error {
	//write plan results
	file := fmt.Sprintf("%s.p%s.tmp.hdf", modelPrefix, rasplan)
	ds, err := pm.GetOutputDataSource(file)
	if err != nil {
		return err
	}
	filepath := fmt.Sprintf("%s/%s", actions.MODEL_DIR, file)
	reader, err := os.Open(filepath)
	if err != nil {
		raslog.WriteString(fmt.Sprintf("Unable to open %s for copying: %s\n", file, err))
	} else {
		defer reader.Close()
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
	_, err = pm.Put(cc.PutOpInput{
		SrcReader: logReader,
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: ds.Name,
			PathKey:        "0",
		},
	})
	return err
}
