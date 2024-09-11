package main

import (
	"log"
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
