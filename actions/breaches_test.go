package actions

import (
	"fmt"
	"testing"
)

const oneBreachBFile string = "/workspaces/cc-ras-runner/TestData/DamBreachOverlapDem.b01"
const multiBreachBFile string = "/workspaces/cc-ras-runner/TestData/multiDamBreach.b01"

func TestGetBreachRows(t *testing.T) {
	bf := InitBFile(oneBreachBFile)
	rows, err := bf.getBreachRows()
	if err != nil || rows == nil {
		t.Fail()
	}

}

func TestGetBreachData(t *testing.T) {
	bf := InitBFile(oneBreachBFile)
	rows, err := bf.getBreachRows()
	if err != nil || rows == nil {
		t.Fail()
	}
	bd, err := bf.rowToBreachData(rows)
	if bd == nil || err != nil {
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
	row := make([]string, 1)
	var bd BreachData = BreachData{}
	row[0] = bd.convertFloatToBfileCellValue(123.4)

	rows := [][]string{}

	rows = append(rows, row)
	rows = append(rows, row)
	rows = append(rows, row)
	rows = append(rows, row)
	rows = append(rows, row)

	bd = BreachData{
		FailureElevationRowNum: 1,
		BreachDataRows:         rows,
	}
	fmt.Print(bd)
	func(data *BreachData) {
		err := data.updateFailureElevation(65432)
		if err != nil {
			t.Fail()
		}
	}(&bd)
	fmt.Print(bd)

	bd.updateFailureElevation(432.1)
	fmt.Print(bd)

}
