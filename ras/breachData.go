package ras

import (
	"errors"
	"fmt"
	"strings"
)

type BreachData struct {
	Name                   string
	SNetID                 int
	NumRows                int
	FailureElevationRowNum int
	BreachDataRows         [][]string
}

func (bd *BreachData) Header() string {
	return "Breach Data" //@TODO: not sure how to handle this
}

func (bd *BreachData) UpdateFloat(value float64) error {
	bd.BreachDataRows[bd.FailureElevationRowNum][0] = convertFloatToBfileCellValue(value)
	return nil
}
func (bd *BreachData) UpdateFloatArray(values []float32) error {
	return errors.New("updating float arrays not currently supported for breach data, could be used to update breach progression")
}
func (bd *BreachData) ToBytes() ([]byte, error) {
	bytes := make([]byte, 0)
	for i := 0; i < bd.NumRows-1; i++ {
		row := fmt.Sprintf("%v\n", strings.Join(bd.BreachDataRows[i], ""))
		bytes = append(bytes, row...)
	}
	return bytes, nil
}
func InitBreachData(rows []string) ([]BfileBlock, error) {
	headerDefaultBlock := DefaultBlock{
		Rows: rows[0:2],
	}
	blocks := make([]BfileBlock, 0)
	blocks = append(blocks, &headerDefaultBlock)
	breachCount, err := getIntFromCellValue(rows[1])
	if err != nil {
		return blocks, err
	}
	if breachCount == 0 {
		return blocks, nil
	}
	//take the remaining rows and create breach data instances for each.
	breachdataRows, err := getBreachRows(rows)
	if err != nil {
		return blocks, err
	}
	breachDatas, err := SplitIntoBreachDataArray(breachdataRows)
	blocks = append(blocks, breachDatas...)
	return blocks, err
}

// The b file is formated into columns 8 characters wide.
// This function returns a row as a string array of "cells" 8 char wide.
func splitRowsIntoCells(row string) ([]string, error) {
	var cellSize int = 8 // RAS B file format.
	var lengthOfRow int = len(row)
	var result []string

	//does the row divide evenly into complete cells?
	if lengthOfRow%cellSize != 0 {
		return nil, errors.New(CELL_SIZE_ERROR)
	}

	//divide the row into cells, add those cells to the array
	numCells := lengthOfRow / cellSize

	for i := 0; i < numCells; i++ {
		var startIndex int = 0 + i*cellSize
		var endIndex int = startIndex + cellSize          // start index is inclusive, end is not. So no -1.
		result = append(result, row[startIndex:endIndex]) // we can treat strings as an array of chars.
	}

	return result, nil
}

// get a slice of rows (which are slices of string cells) that represents all the breach data in the b-file
func getBreachRows(bfileRows []string) ([][]string, error) {
	var breachDataRows [][]string

	for i := 0; i < len(bfileRows); i++ {
		if strings.Contains(bfileRows[i], BREACH_DATA_HEADER) {
			i++ // next line
			isBreachData := true
			var rowText = bfileRows[i]
			for isBreachData { // until we hit another header or empty line, keep going
				row, err := splitRowsIntoCells(rowText)
				if err != nil {
					return breachDataRows, err
				}
				breachDataRows = append(breachDataRows, row)
				i++ //next line
				if i >= len(bfileRows) {
					break
				} else {
					isBreachData = rowIsBreachData(rowText)
				}

				rowText = bfileRows[i]
			}
			if breachDataRows != nil {
				break
			}
		}
	}
	return breachDataRows, nil
}
func SplitIntoBreachDataArray(rows [][]string) ([]BfileBlock, error) {
	var breachdatas []BfileBlock

	numBreachingStructures, err := getIntFromCellValue(rows[0][0])
	if err != nil {
		return breachdatas, err
	}

	structureFirstRowIndex := 1 //0 //1

	for i := 0; i < numBreachingStructures; i++ {

		//create a BreachData Object
		numRowsInStructureBreachData, err := numRowsForStructureInBreachData(rows, structureFirstRowIndex)
		if err != nil {
			return breachdatas, err
		}
		startingElevationRowIndex, err := getStartingElevationRowIndex(rows, structureFirstRowIndex)
		if err != nil {
			return breachdatas, err
		}
		specificRows := rows[structureFirstRowIndex:(structureFirstRowIndex + numRowsInStructureBreachData)]
		cellValue := specificRows[0][0] //Always the first cell for a set of structure breach data.
		sNetID, err := getIntFromCellValue(cellValue)
		if err != nil {
			return nil, err
		}
		bd := BreachData{
			Name:                   "",
			SNetID:                 sNetID,
			NumRows:                numRowsInStructureBreachData,
			FailureElevationRowNum: startingElevationRowIndex,
			BreachDataRows:         specificRows,
		}

		//add it to the list
		breachdatas = append(breachdatas, &bd)

		//update first row index for the next guy.
		structureFirstRowIndex = structureFirstRowIndex + numRowsInStructureBreachData
	}
	return breachdatas, nil
}
func getRow7and8Exist(rows [][]string, firstRowIndex int) (bool, error) {
	columnIndexBreachMethod := 9
	cellValueBreachMethod := rows[firstRowIndex][columnIndexBreachMethod]
	breachMethodIndex, err := getIntFromCellValue(cellValueBreachMethod)
	if err != nil {
		return false, err
	}
	return (breachMethodIndex == 1), nil
}

// row 2 only exists if we're using mass wasting, which as indicated by a 1 in column index 13. if not, it's a 0
func getRow2Exists(rows [][]string, firstRowIndex int) (bool, error) {
	columnIndexMassWasting := 13
	cellValueMassWasting := rows[firstRowIndex][columnIndexMassWasting]
	MassWastingIndex, err := getIntFromCellValue(cellValueMassWasting)
	if err != nil {
		return false, err
	}
	return (MassWastingIndex == 1), nil
}
func getStartingElevationRowIndex(rows [][]string, firstRowIndex int) (int, error) {
	row2exists, err := getRow2Exists(rows, firstRowIndex)
	if err != nil {
		return 0, err
	}
	if row2exists {
		return 3, nil
	}
	return 2, nil
}
func additionalRowsFromStoredOrdinates(rows [][]string, ProgOrdNumIndex int) (int, error) {
	ProgOrdNum, err := getIntFromCellValue(rows[ProgOrdNumIndex][0])
	if err != nil {
		return 0, err
	}
	partialRow := 0
	if ProgOrdNum%5 != 0 {
		partialRow = 1
	}
	fullRow := ProgOrdNum / 5
	return (partialRow + fullRow), err
}
func numRowsForStructureInBreachData(rows [][]string, firstRowIndex int) (int, error) {
	rowCount := 3 // doesn't include any coordinate rows.
	var ProgOrdNumIndex int
	row2Exists, err := getRow2Exists(rows, firstRowIndex)
	if err != nil {
		return 0, err
	}
	row7and8Exist, err := getRow7and8Exist(rows, firstRowIndex)
	if err != nil {
		return 0, err
	}

	//row 5 tells us how many progression or downcutting ordinates we have, which can add extra rows. The existance of row 2 tells us what row that ordinate number is on.
	if row2Exists {
		rowCount += 1
		ProgOrdNumIndex = firstRowIndex + 4
	} else {
		ProgOrdNumIndex = firstRowIndex + 3
	}

	//additional rows from progression/owncutting
	//rowCount += 1 //for the count of coordinates
	additionalRows, err := additionalRowsFromStoredOrdinates(rows, ProgOrdNumIndex)
	if err != nil {
		return 0, err
	}
	rowCount += additionalRows

	//rows 7 and 8 only exist for the simplified physical breaching method. They are the number of oridnates, and a list of ordinates respectively.
	if row7and8Exist {
		rowCount += 1 //for the count of coordinates
		DowncuttingOrdNumIndex := firstRowIndex + rowCount
		additionalRows, err = additionalRowsFromStoredOrdinates(rows, DowncuttingOrdNumIndex)
		if err != nil {
			return 0, err
		}
		rowCount += additionalRows
	}

	return rowCount + 1, nil //plus 1 for the first row.
}

// Checks that We're not a header and not white space
func rowIsBreachData(row string) bool {
	if rowIsNotAHeader(row) && rowIsNotWhiteSpace(row) {
		return true
	}
	return false
}
