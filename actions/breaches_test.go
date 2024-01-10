package actions

import (
	"testing"
)

const oneBreachBFile string = "/workspaces/cc-ras-runner/actions/testResources/DamBreachOverlapDem.b01"

func TestBreaching(t *testing.T) {
	getBreachRows(oneBreachBFile)

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
