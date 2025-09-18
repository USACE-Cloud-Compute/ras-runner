package hdf

import (
	"reflect"
	"testing"
)

const (
	TestRasHdfFile string = "/workspaces/cc-ras-runner/testData/Duwamish_17110013.p01.hdf"
)

func TestReadBcLinePeak(t *testing.T) {
	input := RasExtractInput{
		GroupPath:      "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",
		Colnames:       []string{"stage", "flow"},
		Postprocess:    []string{"max"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriteSummary:   true,
		WriterType:     ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadBcLinePeakJson(t *testing.T) {
	input := RasExtractInput{
		GroupPath:      "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",
		Colnames:       []string{"stage", "flow"},
		Postprocess:    []string{"max"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriteSummary:   true,
		WriteData:      true,
		WriterType:     JsonWriter,
		//WriteAccumulator: &ByteBufferWriteAccumulator{},
		WriteBlockName: "boundary-condition-lines",
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(string(input.WriteAccumulator.Flush()))
}

func TestReadReflineLinePeakWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Water Surface",
		ColNamesDataset: "refline-peak",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineLinePeakFlow(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineTimeSeriesWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		DataType:        reflect.Float32,
		WriteData:       true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineTimeSeriesFlow(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		DataType:        reflect.Float32,
		WriteData:       true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMinVelocity(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMinWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Name",
		Postprocess:     []string{"min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointPeakVelocity(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      JsonWriter,
		ColSize:         40,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointPeakWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMaxAndMinWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max", "min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadStructureVariablePeak(t *testing.T) {
	input := RasExtractInput{
		GroupPath:    "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/Perimeter 1/2D Hyd Conn",
		GroupSuffix:  "Structure Variables",
		Colnames:     []string{"Total Flow", "Weir Flow", "Stage HW", "Stage TW", "Total Culv"},
		Postprocess:  []string{"max", "min"},
		DataType:     reflect.Float32,
		WriteSummary: true,
		WriterType:   ConsoleWriter,
	}

	err := DataExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAttributeReader(t *testing.T) {
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
