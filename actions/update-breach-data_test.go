package actions

import (
	"path/filepath"
	"testing"

	"github.com/usace/cc-go-sdk"
)

const ONE_BREACH_FILE string = "/workspaces/cc-ras-runner/testData/DamBreachOverlapDem.b01"
const MULTI_BREACH_FILE string = "/workspaces/cc-ras-runner/testData/multiDamBreach.b01"
const FRAG_CURVE_PATH string = "/workspaces/cc-ras-runner/testData/testFragilityCurveOutput.json"
const BALD_EAGLE_HDF_PATH string = "/BaldEagleDamBrk.g03.hdf"

func TestBfileAction(t *testing.T) {
	parameters := make(map[string]any)
	parameters["bFile"] = filepath.Base(ONE_BREACH_FILE) //these may eventually need to be map[string]any instead of strings. Look at Kanawah-runner manifests as examples.
	parameters["fcFile"] = filepath.Base(FRAG_CURVE_PATH)
	modelDir := filepath.Dir(ONE_BREACH_FILE)
	action := cc.Action{
		Name:        "update-bfile",
		Type:        "update-bfile",
		Description: "update bfile",
		Parameters:  parameters,
	}
	err := UpdateBfileAction(action, modelDir)
	if err != nil {
		t.Fail()
	}
}
