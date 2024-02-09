package actions

import (
	"path/filepath"
	"testing"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/cc-ras-runner/ras"
)

func TestBfileAction(t *testing.T) {
	parameters := make(map[string]any)
	parameters["bFile"] = filepath.Base(ras.ONE_BREACH_FILE) //these may eventually need to be map[string]any instead of strings. Look at Kanawah-runner manifests as examples.
	parameters["fcFile"] = filepath.Base(ras.FRAG_CURVE_PATH)
	modelDir := filepath.Dir(ras.ONE_BREACH_FILE)
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
