package hdf

import (
	"fmt"
	"log"
	"reflect"
	"regexp"

	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const (
	ConsoleWriter RasExtractWriterType = "console"
	JsonWriter    RasExtractWriterType = "json"
	CsvWriter     RasExtractWriterType = "csv"
	EventDbWriter RasExtractWriterType = "eventdb"
)

type RasExtractWriterType string

type RasExtractDataTypes interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string
}

type RasExtractInput struct {
	DataPath        string
	GroupPath       string //for group reads....
	GroupSuffix     string
	MatchPattern    string
	ExcludePattern  string
	Postprocess     []string
	Colnames        []string
	ColNamesDataset string
	ColData         string
	StringSizes     []int
	DataType        reflect.Kind
	WriteData       bool
	WriteSummary    bool
	WriterType      RasExtractWriterType
}

type WriteRasDataInput[T RasExtractDataTypes] struct {
	Data         *RasExtractData[T]
	OutputName   string
	Colnames     []string
	WriteData    bool
	WriteSummary bool
}

func DataExtract(input RasExtractInput, filepath string) error {
	extractor, err := NewRasExtractor(filepath)
	if err != nil {
		return err
	}

	var datasets []string
	if input.GroupPath != "" {
		datasetNames, err := extractor.GroupMembers(input.GroupPath)
		if err != nil {
			return fmt.Errorf("unable to read hdf5 group objects: %s", err)
		}
		datasets = []string{}
		for _, dsname := range datasetNames {
			var include bool

			if input.MatchPattern != "" {
				re := regexp.MustCompile(input.MatchPattern)
				include = re.MatchString(dsname)
			} else if input.ExcludePattern != "" {
				re := regexp.MustCompile(input.ExcludePattern)
				include = !re.MatchString(dsname)
			} else {
				include = true
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

	for _, dataset := range datasets {
		input.DataPath = dataset
		err := extractor.RunExtract(input)
		if err != nil {
			return fmt.Errorf("failed to extract dataset: %s due to error %s", dataset, err)
		}
	}
	return nil
}

func NewRasExtractor(filepath string) (*RasExtractor, error) {
	extractor := RasExtractor{}
	err := extractor.open(filepath)
	return &extractor, err
}

type RasExtractor struct {
	f *hdf5.File
}

func (rer *RasExtractor) open(filepath string) error {
	f, err := hdf5utils.OpenFile(filepath)
	if err != nil {
		return err
	}
	rer.f = f
	return nil
}

func (rer *RasExtractor) Close() error {
	return rer.f.Close()
}

func (rer *RasExtractor) GroupMembers(groupPath string) ([]string, error) {
	group, err := hdf5utils.NewHdfGroup(rer.f, groupPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the hdf group '%s': %s", groupPath, err)
	}
	defer group.Close()
	return group.ObjectNames()
}

func flattenArray[T any](input [][]T, index int) []T {
	out := make([]T, len(input))
	for i, v := range input {
		out[i] = v[index]
	}
	return out
}

func (rer *RasExtractor) RunExtract(input RasExtractInput) error {

	err := rer.columnNamesPreprocessor(&input)
	if err != nil {
		return err
	}

	switch input.DataType {
	case reflect.Float32:
		reader := RasExtractorReader[float32]{rer.f}
		out, err := reader.Read(input)
		if err != nil {
			return err
		}
		writer := ConsoleRasExtractWriter[float32]{}
		writer.Write(WriteRasDataInput[float32]{
			Data:         out,
			WriteData:    input.WriteData,
			WriteSummary: input.WriteSummary,
			Colnames:     input.Colnames,
			OutputName:   input.DataPath,
		})

	}
	return nil
}

func (rer *RasExtractor) columnNamesPreprocessor(input *RasExtractInput) error {
	if input.ColNamesDataset != "" {
		colreader := RasExtractorReader[string]{rer.f}
		cols, err := colreader.ReadArray(RasExtractInput{
			DataPath:    input.ColNamesDataset,
			StringSizes: []int{36},
		})
		if err != nil {
			return err
		}
		input.Colnames = flattenArray(cols, 0)
		input.ColNamesDataset = ""
	}
	return nil
}

type RasExtractorReader[T RasExtractDataTypes] struct {
	f *hdf5.File
}

type RasExtractData[T RasExtractDataTypes] struct {
	data      [][]T
	summaries map[string][]T
}

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

func columnMax[T RasExtractDataTypes](data [][]T, col int) T {
	maxVal := data[0][col]
	for i := range data {
		if data[i][col] > maxVal {
			maxVal = data[i][col]
		}
	}
	return maxVal
}

func columnMin[T RasExtractDataTypes](data [][]T, col int) T {
	minVal := data[0][col]
	for i := range data {
		if data[i][col] < minVal {
			minVal = data[i][col]
		}
	}
	return minVal
}

// ////////////////
// WRITERS !!!
// ////////////////
type RasDataExtractWriter[T RasExtractDataTypes] interface {
	Write(RasExtractData[T]) error
}

type ConsoleRasExtractWriter[T RasExtractDataTypes] struct{}

func (rw *ConsoleRasExtractWriter[T]) Write(input WriteRasDataInput[T]) error {

	//print data
	fmt.Println(input.OutputName)
	if input.WriteData {
		for _, colname := range input.Colnames {
			fmt.Printf("%-20s", colname)
		}
		fmt.Println()
		for _, vals := range input.Data.data {
			for _, val := range vals {
				fmt.Printf("%-20v", val)
			}
			fmt.Println()
		}
	}

	//print summaries
	if input.WriteSummary {
		fmt.Printf("%-10s", "summary")
		for _, colname := range input.Colnames {
			fmt.Printf("%-20s", colname)
		}
		fmt.Println()

		for summaryName, summaryValues := range input.Data.summaries {
			fmt.Printf("%-10s", summaryName)
			for _, val := range summaryValues {
				fmt.Printf("%-20v", val)
			}
			fmt.Println()
		}
	}

	return nil
}

// //////////////////
// /////////////////

type AttributeExtractWriter interface {
	Write(vals map[string]any) error
}

type ConsoleAttributeExtractWriter struct{}

func (cw *ConsoleAttributeExtractWriter) Write(vals map[string]any) error {
	for k, v := range vals {
		fmt.Printf("%s::%v\n", k, v)
	}
	return nil
}

type AttributeExtractInput struct {
	AttributePath  string
	AttributeNames []string
	AttributeTypes []string
	WriterType     RasExtractWriterType
}

func AttributeExtract(input AttributeExtractInput, filepath string) error {
	extractor, err := NewRasExtractor(filepath)
	if err != nil {
		return err
	}
	vals, err := extractor.Attributes(input)
	if err != nil {
		return err
	}
	writer := ConsoleAttributeExtractWriter{}
	writer.Write(vals)
	return nil
}

func (rer *RasExtractor) Attributes(input AttributeExtractInput) (map[string]any, error) {
	vals := make(map[string]any)
	root, err := rer.f.OpenGroup(input.AttributePath)
	if err != nil {
		log.Fatal(err)
	}
	defer root.Close()

	for i, v := range input.AttributeNames {
		err := func() error {
			if root.AttributeExists(v) {
				attr, err := root.OpenAttribute(v)
				if err != nil {
					return err
				}
				defer attr.Close()

				hdf5type, ok := extractorTypeToHdf5Type[input.AttributeTypes[i]]
				if !ok {
					return fmt.Errorf("invalid attribute type: %s", input.AttributeTypes[i])
				}

				switch hdf5type {
				case hdf5.T_GO_STRING:
					var attrdata string
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				case hdf5.T_NATIVE_FLOAT:
					var attrdata float32
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				case hdf5.T_NATIVE_DOUBLE:
					var attrdata float64
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				case hdf5.T_NATIVE_INT32:
					var attrdata int32
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				case hdf5.T_NATIVE_INT64:
					var attrdata int64
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				case hdf5.T_NATIVE_INT8:
					var attrdata int8
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				case hdf5.T_NATIVE_UINT8:
					var attrdata uint8
					attr.Read(&attrdata, hdf5type)
					vals[v] = attrdata
				}
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return vals, nil
}

var extractorTypeToHdf5Type map[string]*hdf5.Datatype = map[string]*hdf5.Datatype{
	"string":  hdf5.T_GO_STRING,
	"float32": hdf5.T_NATIVE_FLOAT,
	"float64": hdf5.T_NATIVE_DOUBLE,
	"int32":   hdf5.T_NATIVE_INT32,
	"int64":   hdf5.T_NATIVE_INT64,
	"int8":    hdf5.T_NATIVE_INT8,
	"uint8":   hdf5.T_NATIVE_UINT8,
}
