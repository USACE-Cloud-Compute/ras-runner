package hdf

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const (
	ConsoleWriter RasExtractWriterType = "console"
	JsonWriter    RasExtractWriterType = "json"
	CsvWriter     RasExtractWriterType = "csv"
	EventDbWriter RasExtractWriterType = "eventdb"
	ByteBuffer    RasExtractWriterType = "bytebuffer"
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
	//ColData         string
	StringSizes      []int
	DataType         reflect.Kind
	WriteData        bool
	WriteSummary     bool
	WriterType       RasExtractWriterType
	WriteAccumulator WriteAccumulator
	WriteBlockName   string
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

	writer, err := getWriter[float32](input.WriterType, input.WriteBlockName, 0)
	if err != nil {
		return err
	}

	//if we are using an accumulator for multiple datasets, write the accoumulator starting tags
	if input.WriteAccumulator != nil {
		input.WriteAccumulator.Write(writer.AccumulatorStart(input.WriteBlockName))
	}

	for i, dataset := range datasets {
		input.DataPath = dataset
		err := extractor.RunExtract(i, input)
		if err != nil {
			return fmt.Errorf("failed to extract dataset: %s due to error %s", dataset, err)
		}
	}

	//if we are using an accumulator, write the closing tags
	if input.WriteAccumulator != nil {
		input.WriteAccumulator.Write(writer.AccumulatorEnd())
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
		//writer := ConsoleRasExtractWriter[float32]{}
		writer.Write(WriteRasDataInput[float32]{
			Data:         out,
			WriteData:    input.WriteData,
			WriteSummary: input.WriteSummary,
			Colnames:     input.Colnames,
			OutputName:   input.DataPath,
		})
		if input.WriteAccumulator != nil {
			data, err := writer.Flush()
			if err != nil {
				return err
			}
			input.WriteAccumulator.Write(data)
		}

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
	AccumulatorStart(startTag string) []byte
	AccumulatorEnd() []byte
	Write(WriteRasDataInput[T]) error
	Flush() ([]byte, error)
}

func NewJsonRasExtractWriter[T RasExtractDataTypes](blockName string, datasetnum int) (RasDataExtractWriter[T], error) {
	writer := JsonRasExtractWriter[T]{blockName: blockName}
	if datasetnum > 0 {
		writer.body.WriteString(",")
	}
	//writer.body.WriteString(fmt.Sprintf("{\"%s\":", blockName))
	return &writer, nil
}

// JSON Writer
type JsonRasExtractWriter[T RasExtractDataTypes] struct {
	blockName string
	body      strings.Builder
}

func (rw *JsonRasExtractWriter[T]) AccumulatorStart(startTag string) []byte {
	return []byte(fmt.Sprintf("{\"%s\":[", startTag))
}

func (rw *JsonRasExtractWriter[T]) AccumulatorEnd() []byte {
	return []byte("]}")
}

func (rw *JsonRasExtractWriter[T]) Write(input WriteRasDataInput[T]) error {
	builder := strings.Builder{}
	builder.WriteString("{")
	builder.WriteString(fmt.Sprintf("\"dataset\":\"%s\",", input.OutputName))
	if input.WriteData {
		builder.WriteString("\"columns\":[")
		for i, colname := range input.Colnames {
			if i > 0 {
				builder.WriteString(",")
			}
			builder.WriteString(fmt.Sprintf("\"%s\"", colname))
		}
		builder.WriteString("],")
		builder.WriteString("\"data\":[")
		for j, vals := range input.Data.data {
			if j > 0 {
				builder.WriteString(",")
			}
			builder.WriteString("[")
			for k, val := range vals {
				if k > 0 {
					builder.WriteString(",")
				}
				builder.WriteString(fmt.Sprintf("%v", val))
			}
			builder.WriteString("]")
		}
		builder.WriteString("]")
	}

	///////////////////////////////////////
	//print summaries

	if input.WriteSummary {
		if input.WriteData {
			builder.WriteString(",")
		}
		builder.WriteString("\"summaries\": {")
		mapindex := 0
		for summaryName, summaryValues := range input.Data.summaries {
			if mapindex > 0 {
				builder.WriteString(",")
			}
			mapindex++
			builder.WriteString(fmt.Sprintf("\"%s\":{", summaryName))
			for i, val := range summaryValues {
				if i > 0 {
					builder.WriteString(",")
				}
				builder.WriteString(fmt.Sprintf("\"%s\":%v", input.Colnames[i], val))
			}
			builder.WriteString("}")
		}
		builder.WriteString("}")
	}
	builder.WriteString("}")
	rw.body.WriteString(builder.String())

	return nil
}

func (rw *JsonRasExtractWriter[T]) Flush() ([]byte, error) {
	return []byte(rw.body.String()), nil
}

// Console Writer
type ConsoleRasExtractWriter[T RasExtractDataTypes] struct{}

func (rw *ConsoleRasExtractWriter[T]) AccumulatorStart(startTag string) []byte {
	return []byte(fmt.Sprintf("-----------starting %s-----------", startTag))
}

func (rw *ConsoleRasExtractWriter[T]) AccumulatorEnd() []byte {
	return []byte("-----------finished-----------")
}

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

func (rw *ConsoleRasExtractWriter[T]) Flush() ([]byte, error) {
	return nil, nil
}

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
	WriteBuffer    []byte
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

	for _, v := range input.AttributeNames {
		err := func() error {
			if root.AttributeExists(v) {
				attr, err := root.OpenAttribute(v)
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
				vals[v] = val.Elem().Interface() //get the value from the pointer
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return vals, nil
}

// /////data accumulator///////////
type WriteAccumulator interface {
	Write(data []byte) error
	Flush() []byte
}

type ByteBufferWriteAccumulator struct {
	data bytes.Buffer
}

func (wa *ByteBufferWriteAccumulator) Write(data []byte) error {
	_, err := wa.data.Write(data)
	return err
}

func (wa *ByteBufferWriteAccumulator) Flush() []byte {
	return wa.data.Bytes()
}
