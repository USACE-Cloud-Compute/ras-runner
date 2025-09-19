package hdf

import (
	"encoding/json"
	"fmt"
	"math"
)

type RasExtractWriterAction string

const (
	RasExtractWrite      RasExtractWriterAction = "write"
	RasExtractAccumulate RasExtractWriterAction = "accumulate"
)

type RasDataExtractWriter[T RasExtractDataTypes] interface {
	Write(WriteRasDataInput[T]) error
}

// var writerAccumulator map[string][]map[string]any = make(map[string][]map[string]any)
var writerAccumulator map[string][]map[string]any = make(map[string][]map[string]any)

// =============================================================================
// JSON writer
// =============================================================================
type RasExtractorOutputBlock[T RasExtractDataTypes] struct {
	Dataset string   `json:"dataset"`
	Columns []string `json:"columns"`
	//Data           [][]T             `json:"data,omitempty"`
	Data           NaNSafeMatrix[T]  `json:"data,omitempty"`
	Summaries      map[string][]any  `json:"summaries,omitempty"`
	RasExtractData RasExtractData[T] `json:"-"`
}

func NewJsonRasExtractWriter[T RasExtractDataTypes](blockName string, datasetnum int) (RasDataExtractWriter[T], error) {
	writer := JsonRasExtractWriter[T]{blockName: blockName}
	return &writer, nil
}

type JsonRasExtractWriter[T RasExtractDataTypes] struct {
	blockName string
}

func (rw *JsonRasExtractWriter[T]) Write(input WriteRasDataInput[T]) error {
	block := RasExtractorOutputBlock[T]{
		Dataset:        input.OutputName,
		Columns:        input.Colnames,
		RasExtractData: *input.Data,
	}

	if input.WriteData {
		block.Data = input.Data.data
	}

	if input.WriteSummary {
		summaries := make(map[string][]any)
		for summaryname, vals := range input.Data.summaries {
			nanSafeVals := make([]any, len(vals))
			for i, val := range vals {
				switch vval := any(val).(type) {
				case float64:
					if math.IsNaN(vval) {
						nanSafeVals[i] = nil
					} else {
						nanSafeVals[i] = vval
					}
				case float32:
					if math.IsNaN(float64(vval)) {
						nanSafeVals[i] = nil
					} else {
						nanSafeVals[i] = vval
					}
				}
			}
			summaries[summaryname] = nanSafeVals
		}
		block.Summaries = summaries
	}

	//adds a new extract to a block section
	writerAccumulator[rw.blockName] = append(writerAccumulator[rw.blockName], map[string]any{input.datasetName: block})
	return nil
}

// =============================================================================
// JSON attribute writer
// =============================================================================

type JsonAttrOutputBlock struct {
	Dataset string   `json:"dataset"`
	Columns []string `json:"columns"`
	Data    [][]any  `json:"data"`
}

func NewJsonAttributeExtractor(blockname string, dataset string) (*JsonAttributeExtractWriter, error) {
	writer := JsonAttributeExtractWriter{blockname: blockname, dataset: dataset}
	return &writer, nil
}

type JsonAttributeExtractWriter struct {
	blockname string
	dataset   string
}

func (jw *JsonAttributeExtractWriter) Write(vals map[string]any) error {
	cols := make([]string, len(vals))
	databox := make([][]any, 1)
	data := make([]any, len(vals))
	count := 0
	for k, v := range vals {
		cols[count] = k
		data[count] = v
		count++
	}
	databox[0] = data
	writerAccumulator[jw.blockname] = []map[string]any{{"attributes": JsonAttrOutputBlock{Columns: cols, Data: databox, Dataset: jw.dataset}}}
	return nil
}

type NaNSafeMatrix[T RasExtractDataTypes] [][]T

func NewSafeMatrix[T RasExtractDataTypes](matrix [][]T) NaNSafeMatrix[T] {
	return NaNSafeMatrix[T](matrix)
}

func (m NaNSafeMatrix[T]) MarshalJSON() ([]byte, error) {

	result := make([][]interface{}, len(m))

	for i, row := range m {
		result[i] = make([]interface{}, len(row))
		for j, v := range row {
			switch val := any(v).(type) {
			case float32:
				if math.IsNaN(float64(val)) {
					result[i][j] = nil
				} else {
					result[i][j] = val
				}
			case float64:
				if math.IsNaN(val) {
					result[i][j] = nil
				} else {
					result[i][j] = val
				}
			default:
				result[i][j] = val
			}
		}
	}

	return json.Marshal(result)
}

// =============================================================================
// Console writer
// =============================================================================
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

// =============================================================================
// Console attribute writer
// =============================================================================

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
