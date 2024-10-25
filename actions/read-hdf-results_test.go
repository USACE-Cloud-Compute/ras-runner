package actions

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/usace/cc-go-sdk"
)

func Testread_hdf_results(t *testing.T) {
	parameters := make(map[string]any)
	parameters["bFile"] = filepath.Base(ONE_BREACH_FILE)
	parameters["fcFile"] = filepath.Base(FRAG_CURVE_PATH)
	parameters["geoHdfFile"] = filepath.Base(Geometry_HDF_PATH)
	modelDir := filepath.Dir(ONE_BREACH_FILE)
	action := cc.Action{
		Name:        "update-bfile",
		Type:        "update-bfile",
		Description: "update bfile",
		Parameters:  parameters,
	}
	err := UpdateBfileAction(action, modelDir)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}
