package actions

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"
	"unicode"
)

const rowLengthCellSizeError string = "the row was not able to be divided evenly by the cell size without remainder. Ensure the b-file has not been modified outside of RAS"
const breachDataHeader string = "Breach Data"

func getBreachRows(bfilePath string) [][]string {
	var breachDataRows [][]string

	file, err := os.Open(bfilePath)
	if err != nil {
		log.Fatal(err)
	}
	//close the file when we're done
	defer file.Close()

	//read the file line by line
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), breachDataHeader) {
			scanner.Scan()                                 //next line
			var firstLetter rune = rune(scanner.Text()[0]) //first letter as a rune / kinda like a char.
			for !unicode.IsLetter(firstLetter) {           // until we hit another header, keep going
				row, err := splitRowsIntoCells(scanner.Text(), 8)
				if err != nil {
					log.Fatal(err)
					return nil
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
	return breachDataRows
}

// The b file is formated into columns 8 characters wide.
// This function returns a row as a string array of "cells" 8 char wide.
func splitRowsIntoCells(row string, cellSize int) ([]string, error) {
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
