package ras

import (
	"fmt"
	"strconv"
	"testing"
)

const ONE_BREACH_FILE string = "/workspaces/cc-ras-runner/testData/DamBreachOverlapDem.b01"
const MULTI_BREACH_FILE string = "/workspaces/cc-ras-runner/testData/multiDamBreach.b01"
const FRAG_CURVE_PATH string = "/workspaces/cc-ras-runner/testData/testFragilityCurveOutput.json"
const BALD_EAGLE_HDF_PATH string = "/BaldEagleDamBrk.g03.hdf"
const Elk_At_Sutton_BFile string = "/workspaces/cc-ras-runner/testData/ElkRiver_at_Sutton.b01"

func TestRead_ElkAtSutton(t *testing.T) {
	bfile, err := InitBFile(Elk_At_Sutton_BFile)
	if err != nil {
		t.Fail()
	}
	for _, bfileblock := range bfile.BfileBlocks {
		outts, ok := bfileblock.(*OutletTS)
		if ok {
			fmt.Println("Outlet TS found")
			fmt.Println(outts.Name)
			fmt.Println(outts.RowCount)
			return
		}
	}
	t.Fail()
}
func TestWrite(t *testing.T) {
	bf, err := InitBFile(ONE_BREACH_FILE) // hold the original for comparison (expected)
	if err != nil {
		t.Fail()
	}
	bfAmmended, err := InitBFile(ONE_BREACH_FILE) //ammend this one (actual)
	if err != nil {
		t.Fail()
	}
	bfAmmended.SNETidToStructName = make(map[string]int) //ideally I'd create this map from  a geometry HDF associated with this geom.
	bfAmmended.SNETidToStructName["2"] = 2
	err = bfAmmended.AmmendBreachElevations("2", 999)
	if err != nil {
		t.Fail()
	}
	bammend, err := bfAmmended.Write()
	if err != nil {
		t.Fail()
	}
	b, err := bf.Write()
	if err != nil {
		t.Fail()
	}
	var stringbammed string = string(bammend)
	var stringb string = string(b)

	//ammend should change the string that's written.
	if string(bammend) == string(b) {
		t.Fail()
	}
	fmt.Println(stringbammed)
	fmt.Println(stringb)
}
func TestGetBreachRows(t *testing.T) {
	_, err := InitBFile(ONE_BREACH_FILE)
	if err != nil {
		t.Fail()
	}
}
func TestRowsIntoCells(t *testing.T) {
	row, err := splitRowsIntoCells("000000010000004500000005")
	expected := []string{"00000001", "00000045", "00000005"}
	if err == nil {
		for i := 0; i < len(row); i++ {
			if expected[i] != row[i] {
				t.Fail()
			}
		}
	}
}
func TestConvertFloatToBcellValue(t *testing.T) {
	cellValue := convertFloatToBfileCellValue(450.456)
	expected := "450.4560"
	if cellValue != expected {
		t.Fail()
	}
}
func TestEditFailureElevationData(t *testing.T) {
	newFailElev := 432.1
	//set up some fake breach data
	row := make([]string, 1)
	var bd BreachData = BreachData{}
	row[0] = convertFloatToBfileCellValue(123.4)
	rows := [][]string{}
	rows = append(rows, row)
	rows = append(rows, row)
	bd = BreachData{
		FailureElevationRowNum: 1,
		BreachDataRows:         rows,
	}
	bd.UpdateFloat(newFailElev)
	actualFailElev, err := strconv.ParseFloat(bd.BreachDataRows[bd.FailureElevationRowNum][0], 64)
	if err != nil {
		t.Fail()
	}
	if actualFailElev != newFailElev {
		t.Fail()
	}

}

func TestHDFread(t *testing.T) {
	bfile, err := InitBFile(ONE_BREACH_FILE)
	if err != nil {
		t.Fail()
	}
	err = bfile.SetSNetIDToNameFromGeoHDF(BALD_EAGLE_HDF_PATH)
	if err != nil {
		t.Fail()
	}
	if len(bfile.SNETidToStructName) < 10 {
		t.Fail()
	}
	fmt.Print(bfile.SNETidToStructName)
}
