package hdf

import (
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

	TestRasHdfFile := "/mnt/testdata/1/testdata.hdf"

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

	stringSizes, err := a.Action.Attributes.GetIntSlice("stringSizes")
	if err != nil {
		log.Println("no string sizes found")
		stringSizes = nil
	}

	var dt reflect.Kind
	var ok bool
	dataType := a.Action.Attributes.GetStringOrFail("dataType")
	if dt, ok = dataTypeMap[dataType]; !ok {
		return fmt.Errorf("invalid data type: %s", dataType)
	}

	input := RasExtractInput{
		DataPath:        a.Action.Attributes.GetStringOrDefault("dataPath", ""),
		GroupPath:       a.Action.Attributes.GetStringOrDefault("groupPath", ""),
		GroupSuffix:     a.Action.Attributes.GetStringOrDefault("groupSuffix", ""),
		MatchPattern:    a.Action.Attributes.GetStringOrDefault("matchPattern", ""),
		ExcludePattern:  a.Action.Attributes.GetStringOrDefault("excludePattern", ""),
		Postprocess:     postprocessing,
		Colnames:        colnames,
		ColNamesDataset: a.Action.Attributes.GetStringOrDefault("colNamesDataset", ""),
		ColData:         a.Action.Attributes.GetStringOrDefault("coldata", ""),
		StringSizes:     stringSizes,
		DataType:        dt,
		WriteData:       a.Action.Attributes.GetBooleanOrDefault("writedata", false),
		WriteSummary:    a.Action.Attributes.GetBooleanOrDefault("writeSummary", false),
		WriterType:      ConsoleWriter,
	}

	return DataExtract(input, TestRasHdfFile)

}
