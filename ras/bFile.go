package ras

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
	Filename string
	//Rows                []string
	//StructureBreachData []BreachData
	BfileBlocks        []BfileBlock
	SNETidToStructName map[string]int // This should be initialized with a geometry hdf using InitSNETidToStructName("*.g**.hdf")
}
type BfileBlock interface {
	UpdateFloat(value float64) error
	UpdateFloatArray(values []float64) error
	ToBytes() ([]byte, error)
}
type DefaultBlock struct {
	Rows []string
}

func (db *DefaultBlock) UpdateFloat(value float64) error {
	return errors.New("cannot update float on default blocks")
}
func (db *DefaultBlock) UpdateFloatArray(values []float64) error {
	return errors.New("cannot update float array on default blocks")
}
func (db *DefaultBlock) ToBytes() ([]byte, error) {
	bytedata := make([]byte, 0)
	for _, row := range db.Rows {
		bytedata = append(bytedata, row...)
	}
	return bytedata, nil
}
func InitBFile(bfilePath string) (*Bfile, error) {
	bf := Bfile{
		Filename: bfilePath,
	}

	//read bfile and gather all blocks.
	err := bf.readBFile()
	if err != nil {
		return &bf, err
	}
	return &bf, nil

}

// read the file into memory as a slice of string, where each line/row is a string.
func (bf *Bfile) readBFile() error {
	file, err := os.Open(bf.Filename)
	if err != nil {
		return err
	}
	//close the file when we're done
	defer file.Close()

	//read the file line by line
	blocks := make([][]string, 0)
	scanner := bufio.NewScanner(file)
	blockRows := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if rowIsNotAHeader(line) {
			blockRows = append(blockRows, line)
		} else {
			blocks = append(blocks, blockRows)
			//new block
			blockRows := make([]string, 0)
			blockRows = append(blockRows, line)
		}
	}
	bFileBlocks := make([]BfileBlock, 0)
	for _, block := range blocks {
		if strings.Contains(block[0], BREACH_DATA_HEADER) {
			breachData, err := InitBreachData(block)
			if err != nil {
				return err
			}
			bFileBlocks = append(bFileBlocks, breachData...)
		} else if strings.Contains(block[0], TS_OUTFLOW_HEADER) {
			tsOutflowData, err := InitOutletTS(block)
			if err != nil {
				return err
			}
			bFileBlocks = append(bFileBlocks, tsOutflowData)
		} else {
			db := DefaultBlock{Rows: block}
			bFileBlocks = append(bFileBlocks, &db)
		}
	}
	bf.BfileBlocks = bFileBlocks
	return nil
}

func (bf *Bfile) WriteBreachRows(bds []BreachData, bfilePath string) ([]byte, error) {
	return nil, nil
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

func getIntFromCellValue(cell string) (int, error) {
	trimmedCell := strings.TrimSpace(cell)
	return strconv.Atoi(trimmedCell)
}

// AmmendBreachElevations finds the structure breach data which matches the structureName and updates it's elevation in the breach data rows.
func (bf *Bfile) AmmendBreachElevations(structureName string, newFailureElevation float64) error {
	//searching for the right breach data with a loop seems inefficient. I could bring in the geom in the Init and build a dictionary
	if bf.SNETidToStructName == nil {
		return errors.New("use SetSNetIDToNameFromGeoHDF() to set SNETidToStructName property in BFile before ammending elevations")
	}
	targetSNetID := bf.SNETidToStructName[structureName]
	for _, v := range bf.BfileBlocks {
		Breach, ok := v.(*BreachData)
		if ok {
			strucID := Breach.SNetID
			if strucID == targetSNetID {
				err := Breach.UpdateFloat(newFailureElevation)
				if err != nil {
					return err
				}
				return nil
			}
		}

	}
	//if we made it through the loop without finding a structure, it's not there.
	return fmt.Errorf("structure name, %v, did not exist in bFile", structureName)
}

// Write writes bFile to byte array
func (bf Bfile) Write() ([]byte, error) {
	//for each row in th file
	b := make([]byte, 0)
	for _, block := range bf.BfileBlocks {
		blockBytes, err := block.ToBytes()
		if err != nil {
			return b, err
		}
		b = append(b, blockBytes...)
	}
	return b, nil
}

func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
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
