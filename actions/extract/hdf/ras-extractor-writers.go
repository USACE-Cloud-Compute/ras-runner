package hdf

import (
	"fmt"
	"log"
	"path"
	"reflect"

	"github.com/usace/go-hdf5"
)

type RasExtractWriterAction string

const (
	RasExtractWrite      RasExtractWriterAction = "write"
	RasExtractAccumulate RasExtractWriterAction = "accumulate"
)

type RasDataExtractWriter[T RasExtractDataTypes] interface {
	Write(WriteRasDataInput[T]) error
}

var writerAccumulator map[string][]map[string]any = make(map[string][]map[string]any)

// ///////////////JSON Writer//////////////////////////
type RasExtractorOutputBlock[T RasExtractDataTypes] struct {
	Dataset        string                  `json:"dataset"`
	Columns        []string                `json:"columns"`
	Data           [][]T                   `json:"data,omitempty"`
	Summaries      map[string]map[string]T `json:"summaries,omitempty"`
	RasExtractData RasExtractData[T]       `json:"-"`
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
		summaries := make(map[string]map[string]T)
		for summaryname, vals := range input.Data.summaries {
			summary := make(map[string]T)
			for i, colname := range input.Colnames {
				summary[colname] = vals[i]
			}
			summaries[summaryname] = summary
		}
		block.Summaries = summaries
	}
	//writerAccumulator = append(writerAccumulator, map[string]any{rw.blockName: block})
	writerAccumulator[rw.blockName] = append(writerAccumulator[rw.blockName], map[string]any{path.Base(input.OutputName): block})
	return nil
}

// Console Writer
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
// type WriteAccumulator interface {
// 	Write(data []byte) error
// 	Flush() []byte
// }

// type ByteBufferWriteAccumulator struct {
// 	data bytes.Buffer
// }

// func (wa *ByteBufferWriteAccumulator) Write(data []byte) error {
// 	_, err := wa.data.Write(data)
// 	return err
// }

// func (wa *ByteBufferWriteAccumulator) Flush() []byte {
// 	return wa.data.Bytes()
// }
