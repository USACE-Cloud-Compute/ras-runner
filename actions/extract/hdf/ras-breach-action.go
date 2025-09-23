package hdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"

	"ras-runner/actions"

	"github.com/usace/cc-go-sdk"
)

const (
	breachLocationField string = "SaConn"
)

func init() {
	cc.ActionRegistry.RegisterAction("ras-breach-extract", &RasBreachExtractAction{})
}

type RasBreachExtractAction struct {
	cc.ActionRunnerBase
}

func (a *RasBreachExtractAction) Run() error {
	/////////////////
	//event := 1
	//////////////

	modelResultsPath := fmt.Sprintf("%s/%s.p%s.hdf", actions.MODEL_DIR,
		a.Action.Attributes.GetStringOrFail("modelPrefix"),
		a.Action.Attributes.GetStringOrFail("plan"),
	)

	rb, err := NewRasBreachData(modelResultsPath)
	if err != nil {
		return err
	}
	defer rb.Close()

	flowAreas, err := rb.FlowAreas2D()
	if err != nil {
		return err
	}

	breachRecords := []BreachRecord{}
	for _, flowarea2d := range flowAreas {
		connectionNames, err := rb.ConnectionNames(flowarea2d)
		if err != nil {
			return err
		}
		for _, connectionName := range connectionNames {
			bd, err := rb.BreachData(flowarea2d, connectionName)
			if err != nil {
				log.Printf("No breach configuration for %s\n", connectionName)
			} else {
				br := GetBreachRecord(a.PluginManager.EventIdentifier, flowarea2d, connectionName, &bd)
				breachRecords = append(breachRecords, br)
			}
		}
	}

	writer := JsonBreachDataExtractWriter{blockname: "breach_records"}
	writer.Write(breachRecords)

	json, err := json.Marshal(&writerAccumulator)
	if err != nil {
		return err
	}
	outputDataSource := a.Action.Attributes.GetStringOrFail("outputDataSource")
	_, err = a.Action.Put(cc.PutOpInput{
		SrcReader: bytes.NewReader(json),
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: outputDataSource,
			PathKey:        "extract",
		},
	})
	return err
}

type BreachDataExtractWriter interface {
	Write(recs []BreachRecord) error
}

type JsonBreachDataExtractWriter struct {
	blockname string
}

func (writer JsonBreachDataExtractWriter) Write(recs []BreachRecord) error {
	jsonRecs := breachRecordsToJsonAccumulatorMap(recs)
	for _, br := range jsonRecs {
		datasetName := br[breachLocationField]
		dataset := fmt.Sprintf(breachPathTemplate, datasetName.(string))
		outputBlock := RasExtractorOutputBlock[float32]{Dataset: dataset, Record: br}
		writerAccumulator[writer.blockname] = append(writerAccumulator[writer.blockname], map[string]any{datasetName.(string): outputBlock})
	}
	//outputblock:=RasExtractorOutputBlock[float32]{Dataset: }
	//writerAccumulator[writer.blockname] = append(writerAccumulator["breach-records"], jsonRecs...)
	return nil
}

// func (bew *JsonBreachDataExtractWriter) Write(recs []BreachRecord) error {
// 	for
// }

func breachRecordsToJsonAccumulatorMap(recs []BreachRecord) []map[string]any {
	accumMaps := []map[string]any{}
	for _, r := range recs {
		recmap := make(map[string]any)
		v := reflect.ValueOf(r)
		t := v.Type()
		for j := 0; j < v.NumField(); j++ {
			field := t.Field(j)
			value := v.Field(j)
			if value.Kind() == reflect.Float32 || value.Kind() == reflect.Float64 {
				if math.IsNaN(value.Float()) {
					recmap[field.Name] = nil
				} else {
					recmap[field.Name] = value.Interface()
				}
			} else {
				recmap[field.Name] = value.Interface()
			}
		}
		accumMaps = append(accumMaps, recmap)
	}
	return accumMaps
}
