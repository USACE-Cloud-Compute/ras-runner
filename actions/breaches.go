package actions

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const rowLengthCellSizeError string = "the row was not able to be divided evenly by the cell size without remainder. Ensure the b-file has not been modified outside of RAS"
const breachDataHeader string = "Breach Data"

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
	bd.BreachDataRows[bd.FailureElevationRowNum][0] = convertFloatToBfileCellValue(newFailureElevation)
	return nil
}

// get a slice of rows (which are slices of string cells) that represents all the breach data in the b-file
func getBreachRows(bfilePath string) ([][]string, error) {
	var breachDataRows [][]string

	file, err := os.Open(bfilePath)
	if err != nil {
		return nil, err
	}
	//close the file when we're done
	defer file.Close()

	//read the file line by line
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), breachDataHeader) {
			scanner.Scan() // next line
			isBreachData := true
			var rowText = scanner.Text()
			for isBreachData { // until we hit another header or empty line, keep going
				row, err := splitRowsIntoCells(rowText)
				if err != nil {
					return breachDataRows, err
				}
				breachDataRows = append(breachDataRows, row)
				scanner.Scan()
				rowText = scanner.Text()
				isBreachData = rowIsBreachData(rowText)
			}
			if breachDataRows != nil {
				break
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return breachDataRows, nil
}

// Checks that We're not a header and not white space
func rowIsBreachData(row string) bool {
	if rowIsNotAHeader(row) && rowIsNotWhiteSpace(row) {
		return true
	}
	return false
}

// /Headers always start with a letter, Checks if the row starts with a letter, if it doesn't, returns false.
func rowIsNotAHeader(row string) bool {
	var firstLetter rune = rune(row[0]) //first letter as a rune / kinda like a char.
	isAHeader := unicode.IsLetter(firstLetter)
	return !isAHeader
}

// /checks that a row isn't completely empty. if it is, return true, if not, false.
func rowIsNotWhiteSpace(row string) bool {
	for i := 0; i < len(row); i++ {
		charecter := row[i]
		if !unicode.IsSpace(rune(charecter)) {
			return true
		}
	}
	return false
}

// The b file is formated into columns 8 characters wide.
// This function returns a row as a string array of "cells" 8 char wide.
func splitRowsIntoCells(row string) ([]string, error) {
	var cellSize int = 8 // RAS B file format.
	var lengthOfRow int = len(row)
	var result []string

	//does the row divide evenly into complete cells?
	if lengthOfRow%cellSize != 0 {
		return nil, errors.New(rowLengthCellSizeError)
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

func convertFloatToBfileCellValue(fl float64) string {
	// Round the float to 8 digits
	rounded := math.Round(fl*1e8) / 1e8

	// Convert the float to a string with 8 characters
	result := fmt.Sprintf("%8.8f", rounded)

	// Trim the excess characters
	if len(result) > 8 {
		result = result[:8]
	}

	return result
}

func numRowsForStructureInBreachData(rows [][]string, firstRowIndex int) int {
	rowCount := 5
	var ProgOrdNumIndex int
	row2Exists := getRow2Exists(rows, firstRowIndex)
	row7and8Exist := getRow7and8Exist(rows, firstRowIndex)

	//row 5 tells us how many progression or downcutting ordinates we have, which can add extra rows. The existance of row 2 tells us what row that ordinate number is on.
	if row2Exists {
		rowCount += 1
		ProgOrdNumIndex = firstRowIndex + 4
	} else {
		ProgOrdNumIndex = firstRowIndex + 3
	}

	//additional rows from progression/owncutting
	additionalRows := additionalRowsFromStoredOrdinates(rows, ProgOrdNumIndex)
	rowCount += additionalRows

	//rows 7 and 8 only exist for the simplified physical breaching method. They are the number of oridnates, and a list of ordinates respectively.
	if row7and8Exist {
		rowCount += 2
		DowncuttingOrdNumIndex := firstRowIndex + rowCount - 1
		additionalRows = additionalRowsFromStoredOrdinates(rows, DowncuttingOrdNumIndex)
		rowCount += additionalRows
	}

	return rowCount
}

func additionalRowsFromStoredOrdinates(rows [][]string, ProgOrdNumIndex int) int {
	ProgOrdNum, _ := getIntFromCellValue(rows[ProgOrdNumIndex][0])
	partialRow := 0
	if ProgOrdNum%5 != 0 {
		partialRow = 1
	}
	fullRow := ProgOrdNum / 5
	return partialRow + fullRow
}

func getStartingElevationRowIndex(rows [][]string, firstRowIndex int) int {
	if getRow2Exists(rows, firstRowIndex) {
		return 2
	}
	return 1
}

func getRow2Exists(rows [][]string, firstRowIndex int) bool {
	columnIndexMassWasting := 13
	cellValueMassWasting := rows[firstRowIndex][columnIndexMassWasting]
	MassWastingIndex, _ := getIntFromCellValue(cellValueMassWasting)
	if MassWastingIndex == 0 || MassWastingIndex == 1 {
		return true
	}
	return false
}

func getIntFromCellValue(cell string) (int, error) {
	trimmedCell := strings.TrimSpace(cell)
	return strconv.Atoi(trimmedCell)
}

func getRow7and8Exist(rows [][]string, firstRowIndex int) bool {
	columnIndexBreachMethod := 10
	cellValueBreachMethod := rows[firstRowIndex][columnIndexBreachMethod]
	breachMethodIndex, _ := getIntFromCellValue(cellValueBreachMethod)
	return breachMethodIndex == 1
}

func BreakBreachDataOutForSeparateStructures(rows [][]string) []BreachData {
	var breachdatas []BreachData

	numBreachingStructures, err := getIntFromCellValue(rows[0][0])
	if err != nil {
		panic(err)
	}

	structureFirstRowIndex := 1

	for i := 0; i < numBreachingStructures; i++ {

		//create a BreachData Object
		numRowsInStructureBreachData := numRowsForStructureInBreachData(rows, structureFirstRowIndex)
		startingElevationRowIndex := getStartingElevationRowIndex(rows, structureFirstRowIndex)
		specificRows := rows[structureFirstRowIndex:numRowsInStructureBreachData]
		bd := InitBreachData(startingElevationRowIndex, specificRows)

		//add it to the list
		breachdatas = append(breachdatas, bd)

		//update first row index for the next guy.
		structureFirstRowIndex = structureFirstRowIndex + numRowsInStructureBreachData
	}
	return breachdatas

}
