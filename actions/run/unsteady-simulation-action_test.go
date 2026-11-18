package run

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"ras-runner/actions"
	"testing"

	"github.com/usace-cloud-compute/cc-go-sdk"
)

// func TestMain(m *testing.M) {
// 	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
// 	slog.SetDefault(logger)
// }

func setupTest() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func TestRunUnsteadyMuncieAction(t *testing.T) {
	setupTest()
	filesrc := "../../testData/actions/run/unsteady-simulation"
	filedest := actions.MODEL_DIR

	err := os.CopyFS(filedest, os.DirFS(filesrc))
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(filedest)

	f, err := os.Open(fmt.Sprintf("%s/%s", filesrc, "payload"))
	if err != nil {
		t.Error(err)
	}
	payload := cc.Payload{}
	err = json.NewDecoder(f).Decode(&payload)
	runner := UnsteadySimulationAction{
		ActionRunnerBase: cc.ActionRunnerBase{
			ActionName: "unsteady-simulation",
			Action:     payload.Actions[0],
			PluginManager: &cc.PluginManager{
				Payload: payload,
			},
		},
	}
	err = runner.Run()
	if err != nil {
		t.Error()
	}

}
