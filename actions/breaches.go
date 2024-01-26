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

// Parsing of these files is guided by the investigation here: https://www.hec.usace.army.mil/confluence/display/FFRD/Deciphering+Breach+Data+in+Intermediate+Files
// nomenclature used in comments, as well as method and variable names is done to reflect the language on the above page.
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

type Bfile struct {
	Filename string
}

func InitBFile(bfilePath string) Bfile {
	return Bfile{
		Filename: bfilePath,
	}
}

// get a slice of rows (which are slices of string cells) that represents all the breach data in the b-file
func (bf Bfile) getBreachRows() ([][]string, error) {
	var breachDataRows [][]string

	file, err := os.Open(bf.Filename)
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
				row, err := bf.splitRowsIntoCells(rowText)
				if err != nil {
					return breachDataRows, err
				}
				breachDataRows = append(breachDataRows, row)
				scanner.Scan()
				rowText = scanner.Text()
				isBreachData = bf.rowIsBreachData(rowText)
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

func (bf Bfile) WriteBreachRows(bds []BreachData, bfilePath string) ([]byte, error) {
	return nil, nil
}

// Checks that We're not a header and not white space
func (bf Bfile) rowIsBreachData(row string) bool {
	if bf.rowIsNotAHeader(row) && bf.rowIsNotWhiteSpace(row) {
		return true
	}
	return false
}

// /Headers always start with a letter, Checks if the row starts with a letter, if it doesn't, returns false.
func (Bfile) rowIsNotAHeader(row string) bool {
	var firstLetter rune = rune(row[0]) //first letter as a rune / kinda like a char.
	isAHeader := unicode.IsLetter(firstLetter)
	return !isAHeader
}

// /checks that a row isn't completely empty. if it is, return true, if not, false.
func (Bfile) rowIsNotWhiteSpace(row string) bool {
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
func (Bfile) splitRowsIntoCells(row string) ([]string, error) {
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

func (BreachData) convertFloatToBfileCellValue(fl float64) string {
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

func (bf Bfile) numRowsForStructureInBreachData(rows [][]string, firstRowIndex int) (int, error) {
	rowCount := 3 // doesn't include any coordinate rows.
	var ProgOrdNumIndex int
	row2Exists, err := bf.getRow2Exists(rows, firstRowIndex)
	if err != nil {
		return 0, err
	}
	row7and8Exist, err := bf.getRow7and8Exist(rows, firstRowIndex)
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
	rowCount += 1 //for the count of coordinates
	additionalRows, err := bf.additionalRowsFromStoredOrdinates(rows, ProgOrdNumIndex)
	if err != nil {
		return 0, err
	}
	rowCount += additionalRows

	//rows 7 and 8 only exist for the simplified physical breaching method. They are the number of oridnates, and a list of ordinates respectively.
	if row7and8Exist {
		rowCount += 1 //for the count of coordinates
		DowncuttingOrdNumIndex := firstRowIndex + rowCount
		additionalRows, err = bf.additionalRowsFromStoredOrdinates(rows, DowncuttingOrdNumIndex)
		if err != nil {
			return 0, err
		}
		rowCount += additionalRows
	}

	return rowCount, nil
}

func (Bfile) additionalRowsFromStoredOrdinates(rows [][]string, ProgOrdNumIndex int) (int, error) {
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

func (bf Bfile) getStartingElevationRowIndex(rows [][]string, firstRowIndex int) (int, error) {
	row2exists, err := bf.getRow2Exists(rows, firstRowIndex)
	if err != nil {
		return 0, err
	}
	if row2exists {
		return 3, nil
	}
	return 2, nil
}

// row 2 only exists if we're using mass wasting, which as indicated by a 1 in column index 13. if not, it's a 0
func (Bfile) getRow2Exists(rows [][]string, firstRowIndex int) (bool, error) {
	columnIndexMassWasting := 13
	cellValueMassWasting := rows[firstRowIndex][columnIndexMassWasting]
	MassWastingIndex, err := getIntFromCellValue(cellValueMassWasting)
	if err != nil {
		return false, err
	}
	return (MassWastingIndex == 1), nil
}

func getIntFromCellValue(cell string) (int, error) {
	trimmedCell := strings.TrimSpace(cell)
	return strconv.Atoi(trimmedCell)
}

func (Bfile) getRow7and8Exist(rows [][]string, firstRowIndex int) (bool, error) {
	columnIndexBreachMethod := 9
	cellValueBreachMethod := rows[firstRowIndex][columnIndexBreachMethod]
	breachMethodIndex, err := getIntFromCellValue(cellValueBreachMethod)
	if err != nil {
		return false, err
	}
	return (breachMethodIndex == 1), nil
}

func (bf Bfile) GetBreachData(rows [][]string) ([]BreachData, error) {
	var breachdatas []BreachData

	numBreachingStructures, err := getIntFromCellValue(rows[0][0])
	if err != nil {
		panic(err)
	}

	structureFirstRowIndex := 1

	for i := 0; i < numBreachingStructures; i++ {

		//create a BreachData Object
		numRowsInStructureBreachData, err := bf.numRowsForStructureInBreachData(rows, structureFirstRowIndex)
		if err != nil {
			return nil, err
		}
		startingElevationRowIndex, err := bf.getStartingElevationRowIndex(rows, structureFirstRowIndex)
		if err != nil {
			return nil, err
		}
		specificRows := rows[structureFirstRowIndex:(structureFirstRowIndex + numRowsInStructureBreachData)]
		bd := InitBreachData(startingElevationRowIndex, specificRows)

		//add it to the list
		breachdatas = append(breachdatas, bd)

		//update first row index for the next guy.
		structureFirstRowIndex = structureFirstRowIndex + numRowsInStructureBreachData
	}
	return breachdatas, nil
}

func AmmendBreachElevations(newFailureElevationsByIndex map[int]float64, structureBreachData []BreachData) error {
	countNewFailureElevs := len(newFailureElevationsByIndex)
	countStructuresBreaching := len(structureBreachData)
	if countNewFailureElevs != countStructuresBreaching {
		return errors.New("the number of new elevations, and available structures did not match")
	}
	for i := 0; i < countNewFailureElevs; i++ {
		var structure *BreachData = &structureBreachData[i]
		strucID, err := structure.getUnetID()
		if err != nil {
			return err
		}
		err = structure.updateFailureElevation(newFailureElevationsByIndex[strucID])
		if err != nil {
			return err
		}
	}
	return nil

}

//TODO: Write ammended data back to b01

//TODO: Get the SNET-ID from the Geometry HDF.
