package actions

import "strings"

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
	cellValue := bd.BreachDataRows[0][0] //Always the first cell for a set of structure breach data.
	id, err := getIntFromCellValue(cellValue)
	if err != nil {
		return 0, err
	}
	return id, nil
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
