package run

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"ras-runner/actions"
	"ras-runner/actions/extract/hdf"
	"strings"

	"github.com/usace-cloud-compute/cc-go-sdk"
)

const (
	rasModelSummaryPath                    string = "/Results/Unsteady/Summary"
	rasModelSummarySolutionAttrName        string = "Solution"
	rasModelSummarySolutionSuccessCriteria string = "Finished Successfully"
	outputDataSourcePathKey                string = "output"
	outputLogDataSourcePathKey             string = "log"
)

var rasModelSummaryExtractFields []string = []string{"Solution", "Time Stamp Solution Went Unstable"}

func init() {
	cc.ActionRegistry.RegisterAction("unsteady-simulation", &UnsteadySimulationAction{})
}

type UnsteadySimulationAction struct {
	cc.ActionRunnerBase
}

func (a UnsteadySimulationAction) Run() error {
	slog.Info("Running unsteady-simulation", "description", a.Action.Description)

	scriptPath := os.Getenv(actions.RAS_SCRIPT_PATH_ENV)
	if scriptPath == "" {
		scriptPath = actions.MODEL_SCRIPT_PATH
	}

	modelPrefix := a.PluginManager.Attributes.GetStringOrFail("modelPrefix")
	plan := a.PluginManager.Attributes.GetStringOrFail("plan") //cfile
	geom := a.PluginManager.Attributes.GetStringOrFail("geom") //bfile

	out := strings.Builder{}

	if gproc, ok := a.PluginManager.Attributes["geom_preproc"]; ok {
		runGeomPreproc, ok := gproc.(string)
		if ok {
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
		} else {
			return fmt.Errorf("invalid geometry preprocessor directive")
		}
	}

	log.Printf("Running model %s\n", a.Action.Description)
	simcmd := fmt.Sprintf("%s/%s", scriptPath, actions.MODEL_SCRIPT)
	log.Printf("Running model script: %s %s %s %s %s\n", simcmd, actions.MODEL_DIR, modelPrefix, plan, geom)
	cmdout, runerr := exec.Command(simcmd, actions.MODEL_DIR, modelPrefix, plan, geom).CombinedOutput()

	// grab any log information and write to output location before dealing with any errors
	out.Write([]byte("---------- RAS Model Output --------------"))
	_, err := out.Write(cmdout)

	err = saveResults(a.PluginManager, modelPrefix, plan, &out)
	if err != nil {
		return fmt.Errorf("failed to save the results: %w", err)
	}

	if runerr != nil {
		return fmt.Errorf("model run failed: %w", runerr)
	}
	//check for failure condition here
	if stable, err := isModelStable(modelPrefix, plan); !stable || err != nil {
		if err != nil {
			return fmt.Errorf("failed to evaluate model stability: %w", err)
		}
		return fmt.Errorf("model was not stable")
	}

	return nil
}

func isModelStable(modelPrefix string, plan string) (bool, error) {
	modelResultsPath := fmt.Sprintf("%s/%s.p%s.tmp.hdf", actions.MODEL_DIR, modelPrefix, plan)

	extractor, err := hdf.NewRasExtractor[int](modelResultsPath)
	if err != nil {
		return false, err
	}
	defer extractor.Close()

	summaryVals, err := extractor.Attributes(hdf.AttributeExtractInput{
		AttributePath:  rasModelSummaryPath,
		AttributeNames: rasModelSummaryExtractFields,
	})
	if err != nil {
		slog.Error("unable to read summary attributes", "error", err)
		return false, err
	}

	if solutionVal, ok := summaryVals[rasModelSummarySolutionAttrName]; ok {
		if solutionString, ok := solutionVal.(string); ok {
			slog.Info("ras solution value", "value", solutionString)
			if strings.Contains(solutionString, rasModelSummarySolutionSuccessCriteria) {
				return true, nil //model results are valid
			}
		}
	}

	return false, nil //no error but model results are not valid
}

func saveResults(pm *cc.PluginManager, modelPrefix string, rasplan string, raslog *strings.Builder) error {
	var errs []error

	// Save model results (HDF)
	file := fmt.Sprintf("%s.p%s.tmp.hdf", modelPrefix, rasplan)
	if err := putModelResult(pm, file); err != nil {
		errs = append(errs, err)
	}

	// Save log (always attempted, diagnostic)
	if err := putLog(pm, raslog); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func putModelResult(pm *cc.PluginManager, file string) error {
	ds, err := pm.GetOutputDataSource(file)
	if err != nil {
		return fmt.Errorf("get output data source for %s: %w", file, err)
	}
	reader, err := os.Open(fmt.Sprintf("%s/%s", actions.MODEL_DIR, file))
	if err != nil {
		return fmt.Errorf("open %s: %w", file, err)
	}
	defer reader.Close()
	_, err = pm.Put(cc.PutOpInput{
		SrcReader: reader,
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: ds.Name,
			PathKey:        outputDataSourcePathKey,
		},
	})
	if err != nil {
		return fmt.Errorf("put %s: %w", file, err)
	}
	return nil
}

func putLog(pm *cc.PluginManager, raslog *strings.Builder) error {
	ds, err := pm.GetOutputDataSource("rasoutput")
	if err != nil {
		return fmt.Errorf("get log data source: %w", err)
	}
	_, err = pm.Put(cc.PutOpInput{
		SrcReader: strings.NewReader(raslog.String()),
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: ds.Name,
			PathKey:        outputLogDataSourcePathKey,
		},
	})
	if err != nil {
		return fmt.Errorf("put log: %w", err)
	}
	return nil
}
