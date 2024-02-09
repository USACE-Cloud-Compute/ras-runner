package ras

import (
	"errors"
	"strings"
)

type BreachData struct {
	FailureElevationRowNum int
	BreachDataRows         [][]string
}

func InitBreachData(rowNumber int, breachDataRows [][]string) BreachData {
	return BreachData{
		FailureElevationRowNum: rowNumber,
		BreachDataRows:         breachDataRows,
	}
}

func (bd BreachData) updateFailureElevation(newFailureElevation float64) error {
	bd.BreachDataRows[bd.FailureElevationRowNum][0] = bd.convertFloatToBfileCellValue(newFailureElevation)
	return nil
}

func (bd BreachData) getUnetID() (int, error) {
	if bd.BreachDataRows == nil {
		return 0, errors.New("breach data rows were not set. make sure to initialize through InitBreachData()")
	}
	cellValue := bd.BreachDataRows[0][0] //Always the first cell for a set of structure breach data.
	sNetID, err := getIntFromCellValue(cellValue)
	if err != nil {
		return 0, err
	}
	return sNetID, nil
}

func (bd BreachData) getRowsAsString() []string {
	var rows []string
	for i := 0; i < len(bd.BreachDataRows); i++ {
		row := bd.BreachDataRows[i]
		mergedRow := strings.Join(row, "")
		rows = append(rows, mergedRow)
	}
	return rows
}
