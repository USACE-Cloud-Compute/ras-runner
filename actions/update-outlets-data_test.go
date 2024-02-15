package actions

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/usace/cc-go-sdk"
)

const ELKATSUTTON_HDF string = "/ElkRiver_at_Sutton.p01.tmp.hdf"
const BLUESTONELOCAL_HDF string = "/BluestoneLocal.p01.tmp.hdf"
const UPPERNEW_HDF string = "/UpperNew.p01.tmp.hdf"
const ELKATSUTTON_BFILE = "/workspaces/cc-ras-runner/testData/ElkRiver_at_Sutton.b01"

func TestOutletTSAction(t *testing.T) {
	parameters := make(map[string]any)
	parameters["bFile"] = filepath.Base(ELKATSUTTON_BFILE)
	parameters["outletTS"] = "SA Conn: SuttonDam (Outlet TS: SuttonDam_OUT)"
	parameters["hdfFile"] = filepath.Base(ELKATSUTTON_HDF)
	parameters["hdfDataPath"] = "/Event Conditions/Unsteady/Boundary Conditions/Flow Hydrographs/SA Conn: SuttonDam (Outlet TS: SuttonDam_OUT)"
	modelDir := filepath.Dir(ELKATSUTTON_BFILE)
	action := cc.Action{
		Name:        "update-outletTS",
		Type:        "update-outletTS",
		Description: "update-outletTS",
		Parameters:  parameters,
	}
	err := UpdateOutletTSAction(action, modelDir)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}
