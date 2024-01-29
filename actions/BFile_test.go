package actions

import (
	"fmt"
	"strconv"
	"testing"
)

const oneBreachBFile string = "/workspaces/cc-ras-runner/TestData/DamBreachOverlapDem.b01"
const multiBreachBFile string = "/workspaces/cc-ras-runner/TestData/multiDamBreach.b01"

func TestWrite(t *testing.T) {
	bf, err := InitBFile(oneBreachBFile) // hold the original for comparison (expected)
	if err != nil {
		t.Fail()
	}
	bfAmmended, err := InitBFile(oneBreachBFile) //ammend this one (actual)
	if err != nil {
		t.Fail()
	}
	mapdata := make(map[int]float64, 1)
	mapdata[2] = 999
	err = bfAmmended.AmmendBreachElevations(mapdata)
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
	_, err := InitBFile(oneBreachBFile)
	if err != nil {
		t.Fail()
	}
}

func TestRowsIntoCells(t *testing.T) {
	var bf Bfile = Bfile{Filename: ""}
	row, err := bf.splitRowsIntoCells("000000010000004500000005")
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
	var bd BreachData = BreachData{}
	cellValue := bd.convertFloatToBfileCellValue(450.456)
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
	row[0] = bd.convertFloatToBfileCellValue(123.4)
	rows := [][]string{}
	rows = append(rows, row)
	rows = append(rows, row)
	bd = BreachData{
		FailureElevationRowNum: 1,
		BreachDataRows:         rows,
	}
	bd.updateFailureElevation(newFailElev)
	actualFailElev, err := strconv.ParseFloat(bd.BreachDataRows[bd.FailureElevationRowNum][0], 64)
	if err != nil {
		t.Fail()
	}
	if actualFailElev != newFailElev {
		t.Fail()
	}

}
