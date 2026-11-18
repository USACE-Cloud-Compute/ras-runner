package hdf

import (
	"encoding/json"
	"fmt"
	"log"
	"ras-runner/actions"
	"reflect"
	"testing"
)

const (
	TestRasHdfFile string = "/workspaces/ras-runner/testData/Duwamish_17110013.p01.hdf"
)

//@TODO made DataExtract generic and broke all of the tests.  Fix the tests!

func TestReadBcLinePeak(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		GroupPath:      "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",
		Colnames:       []string{"stage", "flow"},
		Postprocess:    []string{"max"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriteSummary:   true,
		WriterType:     ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadBcLinePeakJson(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		GroupPath:      "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",
		Colnames:       []string{"stage", "flow"},
		Postprocess:    []string{"max"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriteSummary:   true,
		WriteData:      true,
		WriterType:     JsonWriter,
		WriteBlockName: "boundary-condition-lines",
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
	json, err := json.Marshal(&writerAccumulator)
	fmt.Println(string(json))
}

func TestReadReflineLinePeakWaterSurface(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Water Surface",
		ColNamesDataset: "refline-peak",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		if ec, ok := err.(errorContainer); ok {
			for _, e := range ec.Unwrap() {
				log.Println(e)
			}
		}
		t.Fatal(err)
	}
}

func TestReadReflineLinePeakFlow(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineTimeSeriesWaterSurface(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		DataType:        reflect.Float32,
		WriteData:       true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineTimeSeriesFlow(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		DataType:        reflect.Float32,
		WriteData:       true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMinVelocity(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMinWaterSurface(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Name",
		Postprocess:     []string{"min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

// @TODO...this test seems incomplete
func TestReadRefpointPeakVelocity(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointPeakWaterSurface(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMaxAndMinWaterSurface(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max", "min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadStructureVariablePeak(t *testing.T) {
	actions.InitLogger()
	input := RasExtractInput{
		GroupPath:    "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/Perimeter 1/2D Hyd Conn",
		GroupSuffix:  "Structure Variables",
		Colnames:     []string{"Total Flow", "Weir Flow", "Stage HW", "Stage TW", "Total Culv"},
		Postprocess:  []string{"max", "min"},
		DataType:     reflect.Float32,
		WriteSummary: true,
		WriterType:   ConsoleWriter,
	}

	err := DataExtract[float32](input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

// @TODO attribute extract is only wired for JSON Extractor.  Fix this.
func TestAttributeReader(t *testing.T) {
	actions.InitLogger()
	input := AttributeExtractInput{
		AttributePath:  "/Results/Unsteady/Summary",
		AttributeNames: []string{"Computation Time DSS", "Computation Time Total", "Maximum WSEL Error", "Maximum number of cores"},
		WriterType:     ConsoleWriter,
	}

	err := AttributeExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAttributeReader2(t *testing.T) {
	actions.InitLogger()
	input := AttributeExtractInput{
		AttributePath:  "/",
		AttributeNames: []string{"File Type", "File Version"},
		WriterType:     ConsoleWriter,
	}

	err := AttributeExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}
