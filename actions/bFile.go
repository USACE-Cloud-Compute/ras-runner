package actions

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const CELL_SIZE_ERROR string = "the row was not able to be divided evenly by the cell size without remainder. Ensure the b-file has not been modified outside of RAS"
const BREACH_DATA_HEADER string = "Breach Data"
const STRUCTURE_DATA_PATH string = "Geometry/Structures/Attributes/"
const TS_OUTFLOW_HEADER string = "Outlet TS - "

// Parsing of these files is guided by the investigation here: https://www.hec.usace.army.mil/confluence/display/FFRD/Deciphering+Breach+Data+in+Intermediate+Files
// nomenclature used in comments, as well as method and variable names is done to reflect the language on the above page.

type Bfile struct {
	Filename            string
	Rows                []string
	StructureBreachData []BreachData
	SNETidToStructName  map[string]int // This should be initialized with a geometry hdf using InitSNETidToStructName("*.g**.hdf")
}

type FragilityCurveLocationResult struct {
	Name             string  `json:"location"`
	FailureElevation float64 `json:"failure_elevation"`
}
type ModelResult struct {
	Results []FragilityCurveLocationResult `json:"results"`
}

func InitBFile(bfilePath string) (*Bfile, error) {
	bf := Bfile{
		Filename: bfilePath,
	}

	//sets Rows
	err := bf.readBFile()
	if err != nil {
		return &bf, err
	}

	//set Breach Data
	breachRows, err := bf.getBreachRows(bf.Rows)
	if err != nil {
		return &bf, err
	}
	err = bf.setBreachData(breachRows)
	if err != nil {
		return &bf, err
	}

	//done
	return &bf, nil

}

// read the file into memory as a slice of string, where each line/row is a string.
func (bf *Bfile) readBFile() error {
	var lines []string

	file, err := os.Open(bf.Filename)
	if err != nil {
		return err
	}
	//close the file when we're done
	defer file.Close()

	//read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	bf.Rows = lines
	return nil
}

// get a slice of rows (which are slices of string cells) that represents all the breach data in the b-file
func (bf *Bfile) getBreachRows(bfileRows []string) ([][]string, error) {
	var breachDataRows [][]string

	for i := 0; i < len(bfileRows); i++ {
		if strings.Contains(bfileRows[i], BREACH_DATA_HEADER) {
			i++ // next line
			isBreachData := true
			var rowText = bfileRows[i]
			for isBreachData { // until we hit another header or empty line, keep going
				row, err := bf.splitRowsIntoCells(rowText)
				if err != nil {
					return breachDataRows, err
				}
				breachDataRows = append(breachDataRows, row)
				i++ //next line
				rowText = bfileRows[i]
				isBreachData = bf.rowIsBreachData(rowText)
			}
			if breachDataRows != nil {
				break
			}
		}
	}
	return breachDataRows, nil
}

func (bf *Bfile) WriteBreachRows(bds []BreachData, bfilePath string) ([]byte, error) {
	return nil, nil
}

// Checks that We're not a header and not white space
func (bf *Bfile) rowIsBreachData(row string) bool {
	if bf.rowIsNotAHeader(row) && bf.rowIsNotWhiteSpace(row) {
		return true
	}
	return false
}

// /Headers always start with a letter, Checks if the row starts with a letter, if it doesn't, returns false.
func (*Bfile) rowIsNotAHeader(row string) bool {
	var firstLetter rune = rune(row[0]) //first letter as a rune / kinda like a char.
	isAHeader := unicode.IsLetter(firstLetter)
	return !isAHeader
}

// /checks that a row isn't completely empty. if it is, return true, if not, false.
func (*Bfile) rowIsNotWhiteSpace(row string) bool {
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
func (*Bfile) splitRowsIntoCells(row string) ([]string, error) {
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

func (bf *Bfile) numRowsForStructureInBreachData(rows [][]string, firstRowIndex int) (int, error) {
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

func (*Bfile) additionalRowsFromStoredOrdinates(rows [][]string, ProgOrdNumIndex int) (int, error) {
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

func (bf *Bfile) getStartingElevationRowIndex(rows [][]string, firstRowIndex int) (int, error) {
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
func (*Bfile) getRow2Exists(rows [][]string, firstRowIndex int) (bool, error) {
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

func (*Bfile) getRow7and8Exist(rows [][]string, firstRowIndex int) (bool, error) {
	columnIndexBreachMethod := 9
	cellValueBreachMethod := rows[firstRowIndex][columnIndexBreachMethod]
	breachMethodIndex, err := getIntFromCellValue(cellValueBreachMethod)
	if err != nil {
		return false, err
	}
	return (breachMethodIndex == 1), nil
}

// SetBreachData sets the StructureBreachData property on the Bfile.
func (bf *Bfile) setBreachData(rows [][]string) error {
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
			return err
		}
		startingElevationRowIndex, err := bf.getStartingElevationRowIndex(rows, structureFirstRowIndex)
		if err != nil {
			return err
		}
		specificRows := rows[structureFirstRowIndex:(structureFirstRowIndex + numRowsInStructureBreachData)]
		bd := InitBreachData(startingElevationRowIndex, specificRows)

		//add it to the list
		breachdatas = append(breachdatas, bd)

		//update first row index for the next guy.
		structureFirstRowIndex = structureFirstRowIndex + numRowsInStructureBreachData
	}
	bf.StructureBreachData = breachdatas
	return nil
}

// AmmendBreachElevations finds the structure breach data which matches the structureName and updates it's elevation in the breach data rows.
func (bf *Bfile) AmmendBreachElevations(structureName string, newFailureElevation float64) error {
	//searching for the right breach data with a loop seems inefficient. I could bring in the geom in the Init and build a dictionary
	if bf.SNETidToStructName == nil {
		return errors.New("use SetSNetIDToNameFromGeoHDF() to set SNETidToStructName property in BFile before ammending elevations")
	}
	targetSNetID := bf.SNETidToStructName[structureName]
	for _, v := range bf.StructureBreachData {
		strucID, err := v.getUnetID()
		if err != nil {
			return err
		}
		if strucID == targetSNetID {
			err = v.updateFailureElevation(newFailureElevation)
			if err != nil {
				return err
			}
			return nil
		}
	}
	//if we made it through the loop without finding a structure, it's not there.
	return fmt.Errorf("structure name, %v, did not exist in bFile", structureName)
}

// Write writes bFile to byte array
func (bf Bfile) Write() ([]byte, error) {
	//for each row in th file
	b := make([]byte, 0)
	for i := 0; i < len(bf.Rows); i++ {
		if strings.Contains(bf.Rows[i], BREACH_DATA_HEADER) {
			b = append(b, bf.Rows[i]...)
			b = append(b, "\n"...)
			i++ // next line (structure count)
			b = append(b, bf.Rows[i]...)
			b = append(b, "\n"...)
			i++ // first line with structure data
			//for each structure we've got breach data for
			for j := 0; j < len(bf.StructureBreachData); j++ {
				strucRows := bf.StructureBreachData[j].getRowsAsString()
				//for each row of data for that structure
				for k := 0; k < len(strucRows); k++ {
					bf.Rows[i] = strucRows[k]
					b = append(b, bf.Rows[i]...)
					b = append(b, "\n"...)
					i++
				}
			}
		}
		b = append(b, bf.Rows[i]...)
		b = append(b, "\n"...)
	}
	return b, nil
}

func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

// UpdateBfileAction reads a fragility curve output file and uses it to read and write a bfile with updated elevations.
func UpdateBfileAction(action cc.Action, modelDir string) error {
	// Assumes bFile and fragility curve file  were copied local with the CopyLocal action.
	log.Printf("Ready to update bFile.")
	bFileName := action.Parameters["bFile"].(string) //these may eventually need to be map[string]any instead of strings. Look at Kanawah-runner manifests as examples.
	bfilePath := fmt.Sprintf("%v/%v", modelDir, bFileName)
	if !fileExists(bfilePath) {
		log.Fatalf("Input source %s, was not found in local directory. Run copy-local first", bfilePath)
	}
	bf, err := InitBFile(bfilePath)
	if err != nil {
		log.Fatal(err)
	}
	fcFileName := action.Parameters["fcFile"].(string)
	fcFilePath := fmt.Sprintf("%v/%v", modelDir, fcFileName)
	fcFileBytes, err := os.ReadFile(fcFilePath)
	if err != nil {
		log.Fatalf("Error getting input source %s", fcFileName) //why don't we use err?
		return err
	}
	var fcResult ModelResult
	err = json.Unmarshal(fcFileBytes, &fcResult)
	if err != nil {
		log.Fatalf("Error getting input source %s", fcFileName)
		return err
	}
	for _, fclr := range fcResult.Results {
		bf.AmmendBreachElevations(fclr.Name, fclr.FailureElevation)
	}
	resultBytes, err := bf.Write()
	if err != nil {
		log.Fatalf("Error getting input source %s", fcFileName)
		return err
	}
	return os.WriteFile(bfilePath, resultBytes, 0600)
}

// Create a dictionary of SNET-ID to structure name from the Geometry HDF.
func (bf *Bfile) SetSNetIDToNameFromGeoHDF(filePath string) error {
	//need to get a handle on the table located at STRUCTURE_DATA_PATH
	if !fileExists(filePath) {
		return errors.New("file doesn't exist")
	}
	hdfReadOptions := hdf5utils.HdfReadOptions{
		Dtype:              0,
		Strsizes:           hdf5utils.HdfStrSet{},
		IncrementalRead:    false,
		IncrementalReadDir: 0,
		IncrementSize:      0,
		ReadOnCreate:       false,
		Filepath:           filePath,
		File:               &hdf5.File{},
	}
	file, err := hdf5utils.NewHdfDataset(STRUCTURE_DATA_PATH, hdfReadOptions)
	if err != nil {
		log.Fatal(err)
	}

	//Read the column. Should get some slice of structure names from "Connection" at column index 5
	var structureNames []string
	err = file.ReadColumn(5, structureNames) //this probably assigns the data to param 2?
	if err != nil {
		log.Fatal(err)
	}

	//the SNET ID is the index of the structure in the table at STRUCTURE_DATA_PATH +2
	sNetIDDict := make(map[string]int, len(structureNames))
	for index, name := range structureNames {
		sNetIDDict[name] = index + 2
	}

	bf.SNETidToStructName = sNetIDDict
	return nil
}
