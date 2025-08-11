package hdf

import (
	"testing"
)

func TestReadBclinePeakAction(t *testing.T) {
	hdffile := "/mnt/testdata/%d/testdata.hdf"
	input := ReadBcLinePeakInput{
		StartEventIndex: 1,
		EndEventIndex:   1,
		Hdf5Path:        hdffile,
		BucketPath:      "",
		BcLines:         []string{"HHRes_Outlet_DS - Flow per Cell", "S_BIGSOOS10 - Flow per Cell"},
		VariableType:    "flow",
	}

	_, err := readBcLinePeak(input)
	if err != nil {
		t.Fatal(err)
	}

}
