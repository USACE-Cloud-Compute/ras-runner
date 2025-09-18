package hdf

import (
	"fmt"
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
	Dataset        string            `json:"dataset"`
	Columns        []string          `json:"columns"`
	Data           [][]T             `json:"data,omitempty"`
	Summaries      map[string][]T    `json:"summaries,omitempty"`
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
		summaries := make(map[string][]T)
		for summaryname, vals := range input.Data.summaries {
			summaries[summaryname] = vals
		}
		block.Summaries = summaries
	}

	// var outputname string
	// if input.datasetName != "" {
	// 	outputname = input.datasetName
	// } else {
	// 	outputname = path.Base(input.OutputName)
	// }

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
