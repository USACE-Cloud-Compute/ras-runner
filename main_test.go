package main

import (
	"log"
	"os"
	"ras-runner/actions"
	"strings"
	"testing"

	"github.com/usace/cc-go-sdk"
)

func TestSaveOutput(t *testing.T) {
	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Print(err)
		t.Fail()
	}
	modelPrefix = "ElkRiver_at_Sutton"
	var sb strings.Builder
	sb.WriteString("hello this is a log")
	saveResults(pm, "01", &sb)

}
func TestHDF_ATTRS(t *testing.T) {
	err := MakeRasHdfTmp("/workspaces/cc-ras-runner/testData/Duwamish_17110013.p01.hdf", "output.p01.hdf")
	if err != nil {
		t.Fail()
	}
}

func Test_Read_HDF_Attributes(t *testing.T) {
	// set the manifest id environment variable.
	os.Setenv("CC_MANIFEST_ID", "99041f15-8274-4782-b67c-bf6216e9fd95")
	action := cc.Action{
		Name:        "simulation-attribute-metadata",
		Type:        "simulation-attribute-metadata", //ptr[string]("copy-inputs"),
		Description: "simulation-attribute-metadata",
		Parameters: map[string]any{
			"simulationDataSource":   "Duwamish_17110013.p01.hdf",
			"start_event_index":      1,
			"end_event_index":        2,
			"flow_areas":             "Perimeter 1",
			"output_file_dataSource": "simulation_metadata.csv",
			"bucket_prefix":          "FFRD",
		},
	}
	actions.ReadSimulationMetadata(action)
}
