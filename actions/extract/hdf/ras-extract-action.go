package hdf

import (
	"bytes"
	"encoding/json"
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

	modelResultsPath, err := a.Action.GetAbsolutePath("LOCAL", "rasOutput", "default")
	if err != nil {
		log.Printf("missing a LOCAL store/path to the RAS model output")
		return err
	}

	blockName := a.Action.Attributes.GetStringOrDefault("block-name", "data")

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

	input := RasExtractInput{
		DataPath:        a.Action.Attributes.GetStringOrDefault("datapath", ""),
		Attributes:      a.Action.Attributes.GetBooleanOrDefault("attributes", false),
		GroupPath:       a.Action.Attributes.GetStringOrDefault("grouppath", ""),
		GroupSuffix:     a.Action.Attributes.GetStringOrDefault("groupsuffix", ""),
		MatchPattern:    a.Action.Attributes.GetStringOrDefault("match", ""),
		ExcludePattern:  a.Action.Attributes.GetStringOrDefault("exclude", ""),
		Postprocess:     postprocessing,
		Colnames:        colnames,
		ColNamesDataset: a.Action.Attributes.GetStringOrDefault("coldata", ""),
		ColSize:         a.Action.Attributes.GetIntOrDefault("colsize", 0),
		StringSizes:     stringSizes,
		//DataType:        dt,
		WriteData:      a.Action.Attributes.GetBooleanOrDefault("writedata", false),
		WriteSummary:   a.Action.Attributes.GetBooleanOrDefault("writesummary", false),
		WriterType:     RasExtractWriterType(a.Action.Attributes.GetStringOrDefault("outputformat", "console")),
		WriteBlockName: blockName,
		Accumulate:     a.Action.Attributes.GetBooleanOrDefault("accumulate-results", false),
	}

	attrpath, err := a.Action.Attributes.GetString("attributepath")
	if err == nil {
		//this means we have an attr path, so we will process an attr extraction
		fmt.Println(attrpath)
		return err
	}

	if input.Attributes {
		aeinput := AttributeExtractInput{
			AttributePath:  input.DataPath,
			AttributeNames: input.Colnames,
			WriteBlockName: input.WriteBlockName,
		}
		err := AttributeExtract(aeinput, modelResultsPath)
		if err != nil {
			return err
		}

	} else {
		var dt reflect.Kind
		var ok bool
		dataType := a.Action.Attributes.GetStringOrFail("datatype")
		if dt, ok = dataTypeMap[dataType]; !ok {
			return fmt.Errorf("invalid data type: %s", dataType)
		}
		input.DataType = dt
		err = DataExtract(input, modelResultsPath)
		if err != nil {
			return err
		}
	}

	if input.Accumulate {
		//do nothing
		return nil
	} else {
		json, err := json.Marshal(&writerAccumulator)
		if err != nil {
			return err
		}
		_, err = a.Action.Put(cc.PutOpInput{
			SrcReader: bytes.NewReader(json),
			DataSourceOpInput: cc.DataSourceOpInput{
				DataSourceName: "extractOutputTemplate",
				PathKey:        "extract",
			},
		})
		//reset accumulator
		writerAccumulator = make(map[string][]map[string]any)
		return err
	}
}
