package actions

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
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
			rowIsBreachData, _ := rowIsNotAHeader(scanner.Text())
			for rowIsBreachData { // until we hit another header, keep going
				row, err := splitRowsIntoCells(scanner.Text())
				if err != nil {
					return breachDataRows, err
				}
				breachDataRows = append(breachDataRows, row)
				scanner.Scan()
			}
			if breachDataRows != nil {
				break
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return breachDataRows, nil
}

func rowIsNotAHeader(row string) (bool, error) {
	var firstLetter rune = rune(row[0]) //first letter as a rune / kinda like a char.
	isAHeader := unicode.IsLetter(firstLetter)
	return !isAHeader, nil
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

// This assumes no mass wasting. Only concerned with finding trigger elevation.
func readBreachData(breachRows [][]string) {
	//numBreachingStructures, _ := strconv.Atoi(breachRows[0][0])//row 0 column 0s

}
