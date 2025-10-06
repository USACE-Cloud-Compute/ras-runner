package hdf

import (
	"fmt"
	"log"
	"math"
	"path"
	"reflect"
	"regexp"

	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

/*
ras extractor provides functionality for extracting data from HEC-RAS HDF5 files.

It implements extraction capabilities for RAS model outputs stored in HDF5 format and supports three main extraction methods:
1. Dataset Extraction: Extracts a specific dataset from the HDF5 file
2. Group Extraction: Enumerates datasets within an HDF5 group and extracts them
3. Attribute Extraction: Extracts attributes from datasets or groups

The extractor provides generic functions for data extraction that work with various data types,
including numeric and string data. It supports post-processing operations such as calculating
maximum and minimum values for extracted data columns.

Key components:
- RasExtractInput: Configuration structure for extraction parameters
- RasExtractData: Container for extracted data and summaries
- RasExtractor: Main extractor struct that handles HDF5 file operations
- Various writer implementations for different output formats

Supported data types include:
- int, int8, int16, int32, int64
- float32, float64
- string

The extractor integrates with the cc-go-sdk for action execution and supports
accumulation of results for later writing to data sources.
*/

type RasExtractWriterType string

const (
	// ConsoleWriter writes output to console (STDOUT)
	ConsoleWriter RasExtractWriterType = "console"
	// JsonWriter writes output to JSON format
	JsonWriter RasExtractWriterType = "json"
	// CsvWriter writes output to CSV format
	CsvWriter RasExtractWriterType = "csv"
	// EventDbWriter writes output to event database
	EventDbWriter RasExtractWriterType = "eventdb"
	// ByteBuffer writes output to byte buffer
	ByteBuffer RasExtractWriterType = "bytebuffer"
)

// RasExtractDataTypes is a type constraint for supported data types in extraction
type RasExtractDataTypes interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string
}

// RasExtractInput defines the configuration for data extraction operations
type RasExtractInput struct {
	// DataPath is the path to the dataset within the HDF5 file
	DataPath string
	// Attributes indicates whether to extract attributes instead of data
	Attributes bool
	// GroupPath is the path to an HDF5 group whose datasets will be extracted
	GroupPath string //for group reads....
	// GroupSuffix is optional suffix appended to each group object path during reading
	GroupSuffix string
	// MatchPattern is a regular expression to select specific dataset objects within the group
	MatchPattern string
	// ExcludePattern is a regular expression to filter out unwanted dataset objects from selection
	ExcludePattern string
	// Postprocess defines post-processing operations (e.g., "max", "min") for summary calculations
	Postprocess []string
	// Colnames defines column names for the extracted data
	Colnames []string
	// ColNamesDataset is path to a dataset containing column names
	ColNamesDataset string
	// StringSizes defines sizes for string datasets when reading
	StringSizes []int
	// DataType is the reflect.Kind of the data type to extract
	DataType reflect.Kind
	// WriteData indicates whether to write the raw extracted data
	WriteData bool
	// WriteSummary indicates whether to calculate and write summary statistics
	WriteSummary bool
	// WriterType specifies the output format (console, json, etc.)
	WriterType RasExtractWriterType
	// WriteBlockName is the name used for identifying the block in output
	WriteBlockName string
	// Accumulate indicates whether to accumulate results rather than write immediately
	Accumulate bool
	// datasetNames is private field to hold group names during extraction
	datasetNames []string
}

// WriteRasDataInput defines the configuration for writing extracted data
type WriteRasDataInput[T RasExtractDataTypes] struct {
	// Data contains the extracted data and summaries
	Data *RasExtractData[T]
	// OutputName is the name/path of the dataset being written
	OutputName string
	// Colnames are the column names for the data
	Colnames []string
	// WriteData indicates whether to write raw data
	WriteData bool
	// WriteSummary indicates whether to write summary statistics
	WriteSummary bool
	// datasetName is internal field storing the dataset name
	datasetName string
}

// getWriter returns a writer instance based on the specified writer type
func getWriter[T RasExtractDataTypes](writertype RasExtractWriterType, writeBlockName string, datasetnum int) (RasDataExtractWriter[T], error) {
	switch writertype {
	case ConsoleWriter:
		return &ConsoleRasExtractWriter[T]{}, nil
	case JsonWriter:
		return NewJsonRasExtractWriter[T](writeBlockName, datasetnum)
	default:
		return nil, fmt.Errorf("invalid writer type: %s", writertype)
	}
}

// =============================================================================
// Data Extract
// =============================================================================
// DataExtract performs data extraction from an HDF5 file based on the provided input configuration
func DataExtract[T RasExtractDataTypes](input RasExtractInput, filepath string) error {
	extractor, err := NewRasExtractor[T](filepath)
	if err != nil {
		return err
	}

	var datasets []string
	if input.GroupPath != "" {
		datasetNames, err := extractor.GroupMembers(input.GroupPath)
		input.datasetNames = datasetNames
		if err != nil {
			return fmt.Errorf("unable to read hdf5 group objects: %s", err)
		}
		datasets = []string{}
		for _, dsname := range datasetNames {
			var include bool

			if input.MatchPattern == "" && input.ExcludePattern == "" {
				include = true
			} else {
				if input.MatchPattern != "" {
					re := regexp.MustCompile(input.MatchPattern)
					include = re.MatchString(dsname)
				}

				if input.ExcludePattern != "" {
					re := regexp.MustCompile(input.ExcludePattern)
					include = !re.MatchString(dsname)
				}
			}

			if include {
				dsname := fmt.Sprintf("%s/%s", input.GroupPath, dsname)
				if input.GroupSuffix != "" {
					dsname = fmt.Sprintf("%s/%s", dsname, input.GroupSuffix)
				}
				datasets = append(datasets, dsname)
			}
		}
	} else {
		datasets = []string{input.DataPath}
	}

	for i, dataset := range datasets {
		input.DataPath = dataset
		err := extractor.RunExtract(i, input)
		if err != nil {
			log.Printf("failed to extract dataset: %s due to error %s\n", dataset, err)
			//return fmt.Errorf("failed to extract dataset: %s due to error %s", dataset, err)
		}
	}

	return nil
}

// NewRasExtractor creates a new RasExtractor instance for the specified file path
func NewRasExtractor[T RasExtractDataTypes](filepath string) (*RasExtractor[T], error) {
	extractor := RasExtractor[T]{}
	err := extractor.open(filepath)
	return &extractor, err
}

// RasExtractor handles HDF5 file operations for data extraction
type RasExtractor[T RasExtractDataTypes] struct {
	f *hdf5.File
}

// open opens an HDF5 file for reading
func (rer *RasExtractor[T]) open(filepath string) error {
	f, err := hdf5utils.OpenFile(filepath)
	if err != nil {
		return err
	}
	rer.f = f
	return nil
}

// Close closes the underlying HDF5 file
func (rer *RasExtractor[T]) Close() error {
	return rer.f.Close()
}

// GroupMembers returns the names of objects within a specified group path
func (rer *RasExtractor[T]) GroupMembers(groupPath string) ([]string, error) {
	group, err := hdf5utils.NewHdfGroup(rer.f, groupPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the hdf group '%s': %s", groupPath, err)
	}
	defer group.Close()
	return group.ObjectNames()
}

// flattenArray converts a 2D array by transposing rows to columns
func flattenArray[T any](input [][]T, index int) []T {
	out := make([]T, len(input))
	for i, v := range input {
		out[i] = v[index]
	}
	return out
}

// RunExtract executes the extraction process for a single dataset or group
func (rer *RasExtractor[T]) RunExtract(datasetnum int, input RasExtractInput) error {

	err := rer.columnNamesPreprocessor(&input)
	if err != nil {
		return err
	}

	reader := RasExtractorReader[T]{rer.f}
	out, err := reader.Read(input)
	if err != nil {
		return err
	}
	writer, err := getWriter[T](input.WriterType, input.WriteBlockName, datasetnum)
	if err != nil {
		return err
	}

	//set the output name to the dataset path.  later will we extract the path base for naming the dataset block
	outputName := input.DataPath
	datasetName := ""
	if len(input.datasetNames) > 0 {
		datasetName = input.datasetNames[datasetnum]
	} else {
		datasetName = path.Base(input.DataPath)
	}

	writer.Write(WriteRasDataInput[T]{
		Data:         out,
		WriteData:    input.WriteData,
		WriteSummary: input.WriteSummary,
		Colnames:     input.Colnames,
		OutputName:   outputName,
		datasetName:  datasetName,
	})

	return nil
}

// columnNamesPreprocessor handles processing of column names from datasets or direct values
func (rer *RasExtractor[T]) columnNamesPreprocessor(input *RasExtractInput) error {
	if input.ColNamesDataset != "" {
		//get string size for the column
		attr, err := hdf5utils.GetAttrMetadata(rer.f, hdf5utils.DatasetMetadata, input.ColNamesDataset, "")
		if err != nil {
			return fmt.Errorf("unable to read metadata for %s: %s", input.ColNamesDataset, err)
		}
		colreader := RasExtractorReader[string]{rer.f}
		cols, err := colreader.ReadArray(RasExtractInput{
			DataPath:    input.ColNamesDataset,
			StringSizes: []int{int(attr.AttrSize)},
		})
		if err != nil {
			return err
		}
		input.Colnames = flattenArray(cols, 0)
		input.ColNamesDataset = ""
	}
	return nil
}

// RasExtractorReader handles reading data from HDF5 datasets
type RasExtractorReader[T RasExtractDataTypes] struct {
	f *hdf5.File
}

// RasExtractData holds the extracted data and summary statistics
type RasExtractData[T RasExtractDataTypes] struct {
	// data contains the raw extracted data
	data [][]T
	// summaries contains calculated summary statistics for each column
	summaries map[string][]T
}

// Read performs data extraction with optional post-processing
func (rr *RasExtractorReader[T]) Read(input RasExtractInput) (*RasExtractData[T], error) {

	data, err := rr.ReadArray(input)

	output := RasExtractData[T]{}
	output.summaries = make(map[string][]T)
	if err != nil {
		return &output, err
	}

	if len(data) > 0 {
		output.data = data
		cols := len(data[0])
		for _, pp := range input.Postprocess {
			colSummary := make([]T, cols)
			for k := 0; k < cols; k++ {
				switch pp {
				case "max":
					colSummary[k] = columnMax(data, k)
				case "min":
					colSummary[k] = columnMin(data, k)
				}

			}
			output.summaries[pp] = colSummary
		}

	}
	return &output, nil
}

// ReadArray reads data from an HDF5 dataset into a 2D slice
func (rr *RasExtractorReader[T]) ReadArray(input RasExtractInput) ([][]T, error) {
	var strSet hdf5utils.HdfStrSet

	if len(input.StringSizes) > 0 {
		strSet = hdf5utils.NewHdfStrSet(input.StringSizes...)
	}

	dataSet, err := hdf5utils.NewHdfDataset(string(input.DataPath), hdf5utils.HdfReadOptions{
		Dtype:        reflect.TypeFor[T]().Kind(),
		Strsizes:     strSet,
		File:         rr.f,
		ReadOnCreate: true,
	})

	if err != nil {
		return nil, err
	}

	return toSlice[T](dataSet)
}

// =============================================================================
// Attribute Extract
// =============================================================================

// AttributeExtractInput defines the configuration for attribute extraction
type AttributeExtractInput struct {
	// AttributePath is the path to the object whose attributes are extracted
	AttributePath string
	// AttributeNames is a list of attribute names to extract
	AttributeNames []string
	// WriterType specifies the output format for attributes
	WriterType RasExtractWriterType
	// WriteBlockName is the name used for identifying the block in output
	WriteBlockName string

	// AttributeFailureConditionField string
	// AttributeFailureConditionValue any
}

// AttributeExtract extracts attributes from an HDF5 file based on the provided input configuration
func AttributeExtract(input AttributeExtractInput, filepath string) error {
	//@TODO type is irrelevant here.  Probably should make this a completely separate type of extractor
	//. and not reuse the RasExtractor[T]
	extractor, err := NewRasExtractor[int](filepath)
	if err != nil {
		return err
	}
	vals, err := extractor.Attributes(input)
	if err != nil {
		return err
	}

	//@TODO use the proper writer!!!!!!!!!!!!!!!!!!!!!!!!
	//writer := ConsoleAttributeExtractWriter{}
	writer, err := NewJsonAttributeExtractor(input.WriteBlockName, input.AttributePath)
	if err != nil {
		return err
	}
	writer.Write(vals)
	return nil
}

// Attributes extracts attributes from a specified path in the HDF5 file
func (rer *RasExtractor[T]) Attributes(input AttributeExtractInput) (map[string]any, error) {
	vals := make(map[string]any)
	group, err := rer.f.OpenGroup(input.AttributePath)
	if err != nil {
		return nil, err
	}
	defer group.Close()

	for _, v := range input.AttributeNames {
		err := func() error {
			if group.AttributeExists(v) {
				attr, err := group.OpenAttribute(v)
				if err != nil {
					return err
				}
				defer attr.Close()
				attrtype := attr.GetType()
				attrDatatype := &hdf5.Datatype{Identifier: attrtype}
				attrGoDatatype := attrDatatype.GoType()
				val := reflect.New(attrGoDatatype)
				err = attr.Read(val.Interface(), attrDatatype)
				if err != nil {
					return fmt.Errorf("unable to read attribute '%s': %s", v, err)
				}
				if isValueNaN(val.Elem()) { //val is a pointer, so get the elem for a NaN check
					vals[v] = nil
				} else {
					vals[v] = val.Elem().Interface() //get the value from the pointer
				}

				//perform fail on check here
				// if input.AttributeFailureConditionField == v && areEqual(input.AttributeFailureConditionValue, val) {
				// 	log.Printf("failure check triggered: %s was %v\n", "myfield", val)
				// 	os.Exit(1) //exit with error condition
				// }
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return vals, nil
}

// =============================================================================
// Utility functions
// =============================================================================

func areEqual(a, b any) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Get reflect.Value of both values
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// If types are different, they can't be equal
	if va.Type() != vb.Type() {
		return false
	}

	// Use reflect.DeepEqual for comparison
	return reflect.DeepEqual(a, b)
}

// isValueNaN checks if a reflect.Value represents NaN (for floating point types)
func isValueNaN(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Float32, reflect.Float64:
		return math.IsNaN(value.Float())
	}
	return false
}

// toSlice converts an HDF5 dataset to a 2D slice of type T
func toSlice[T any](dataset *hdf5utils.HdfDataset) ([][]T, error) {
	rows := make([][]T, dataset.Rows())
	for i := range rows {
		dest := []T{}
		err := dataset.ReadRow(i, &dest)
		if err != nil {
			return nil, fmt.Errorf("failed to read row %d: %s", i, err)
		}
		rows[i] = dest
	}
	return rows, nil
}

// columnMax returns the maximum value in a specified column of a 2D slice
func columnMax[T RasExtractDataTypes](data [][]T, col int) T {
	maxVal := data[0][col]
	for i := range data {
		if data[i][col] > maxVal {
			maxVal = data[i][col]
		}
	}
	return maxVal
}

// columnMin returns the minimum value in a specified column of a 2D slice
func columnMin[T RasExtractDataTypes](data [][]T, col int) T {
	minVal := data[0][col]
	for i := range data {
		if data[i][col] < minVal {
			minVal = data[i][col]
		}
	}
	return minVal
}
