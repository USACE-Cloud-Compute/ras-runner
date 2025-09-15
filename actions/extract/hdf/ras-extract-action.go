package hdf

import (
	"bytes"
	"fmt"
	"log"
	"reflect"

	"github.com/usace/cc-go-sdk"
)

func init() {
	cc.ActionRegistry.RegisterAction("ras-extract", &RasExtractAction{})
}

var dataTypeMap map[string]reflect.Kind = map[string]reflect.Kind{
	"float32": reflect.Float32,
	"float64": reflect.Float64,
	"int32":   reflect.Int32,
	"int64":   reflect.Int64,
	"string":  reflect.String,
}

type RasExtractAction struct {
	cc.ActionRunnerBase
}

func (a *RasExtractAction) Run() error {

	TestRasHdfFile := "/workspaces/cc-ras-runner/testData/duwamish-test.hdf"

	colnames, err := a.Action.Attributes.GetStringSlice("colnames")
	if err != nil {
		log.Println("error reading or no column names for extraction")
		colnames = nil
	}

	postprocessing, err := a.Action.Attributes.GetStringSlice("postprocess")
	if err != nil {
		log.Println("error or no postprocessing requested")
		postprocessing = nil
	}

	stringSizes, err := a.Action.Attributes.GetIntSlice("stringsizes")
	if err != nil {
		log.Println("no string sizes found")
		stringSizes = nil
	}

	var dt reflect.Kind
	var ok bool
	dataType := a.Action.Attributes.GetStringOrFail("datatype")
	if dt, ok = dataTypeMap[dataType]; !ok {
		return fmt.Errorf("invalid data type: %s", dataType)
	}

	outputAccumulator := ByteBufferWriteAccumulator{}

	input := RasExtractInput{
		DataPath:        a.Action.Attributes.GetStringOrDefault("datapath", ""),
		GroupPath:       a.Action.Attributes.GetStringOrDefault("grouppath", ""),
		GroupSuffix:     a.Action.Attributes.GetStringOrDefault("groupsuffix", ""),
		MatchPattern:    a.Action.Attributes.GetStringOrDefault("match", ""),
		ExcludePattern:  a.Action.Attributes.GetStringOrDefault("exclude", ""),
		Postprocess:     postprocessing,
		Colnames:        colnames,
		ColNamesDataset: a.Action.Attributes.GetStringOrDefault("coldata", ""),
		//ColData:         a.Action.Attributes.GetStringOrDefault("coldata", ""),
		StringSizes:      stringSizes,
		DataType:         dt,
		WriteData:        a.Action.Attributes.GetBooleanOrDefault("writedata", false),
		WriteSummary:     a.Action.Attributes.GetBooleanOrDefault("writesummary", false),
		WriterType:       JsonWriter,
		WriteAccumulator: &outputAccumulator,
	}

	//return DataExtract(input, TestRasHdfFile)
	err = DataExtract(input, TestRasHdfFile)
	if err != nil {
		return err
	}

	_, err = a.Action.Put(cc.PutOpInput{
		SrcReader: bytes.NewReader(outputAccumulator.data.Bytes()),
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: "extractOutputTemplate",
			PathKey:        "extract",
		},
	})

	return err

}
