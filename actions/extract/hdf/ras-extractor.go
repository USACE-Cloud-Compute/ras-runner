package hdf

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

type RasExtractWriterType string

const (
	ConsoleWriter RasExtractWriterType = "console"
	JsonWriter    RasExtractWriterType = "json"
	CsvWriter     RasExtractWriterType = "csv"
	EventDbWriter RasExtractWriterType = "eventdb"
	ByteBuffer    RasExtractWriterType = "bytebuffer"
)

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
	StringSizes     []int
	DataType        reflect.Kind
	WriteData       bool
	WriteSummary    bool
	WriterType      RasExtractWriterType
	WriteBlockName  string
	Accumulate      bool
}

type WriteRasDataInput[T RasExtractDataTypes] struct {
	Data         *RasExtractData[T]
	OutputName   string
	Colnames     []string
	WriteData    bool
	WriteSummary bool
}

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

	for i, dataset := range datasets {
		input.DataPath = dataset
		err := extractor.RunExtract(i, input)
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

func (rer *RasExtractor) RunExtract(datasetnum int, input RasExtractInput) error {

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
		writer, err := getWriter[float32](input.WriterType, input.WriteBlockName, datasetnum)
		if err != nil {
			return err
		}
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
