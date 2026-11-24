//go:build fixture

package hdf

import (
	"encoding/json"
	"math"
	"ras-runner/internal/testutil"
	"reflect"
	"strings"
	"testing"
)

const (
	duwamishHdfFixture = "Duwamish_17110013.p01.hdf"

	tsBase = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series"
)

// duwamishPath returns the fixture path to the Duwamish results file and skips
// the test if the fixture is not present.
func duwamishPath(t *testing.T) string {
	t.Helper()
	testutil.SkipIfNoFile(t, duwamishHdfFixture)
	return testutil.FixturePath(t, duwamishHdfFixture)
}

// resetAccumulator clears the package-level writer accumulator so tests do not
// observe state leaked from other tests (Change C in the testing plan).
func resetAccumulator(t *testing.T) {
	t.Helper()
	writerAccumulator.Reset()
	t.Cleanup(func() { writerAccumulator.Reset() })
}

// accumulatorJson marshals the writer accumulator to JSON for assertions.
func accumulatorJson(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(&writerAccumulator)
	if err != nil {
		t.Fatalf("failed to marshal writer accumulator: %v", err)
	}
	return string(b)
}

// TestDatasetExtractGroupFloat32 exercises the generic DataExtract over an
// HDF5 group of float32 boundary condition datasets.
func TestDatasetExtractGroupFloat32(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	input := RasExtractInput{
		GroupPath:      tsBase + "/Boundary Conditions",
		Colnames:       []string{"stage", "flow"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriterType:     ConsoleWriter,
	}

	if err := DataExtract[float32](input, fpath); err != nil {
		t.Fatalf("DataExtract returned error: %v", err)
	}
}

// TestDatasetExtractGroupWithPostprocess asserts that a group extraction with
// Postprocess ["max"] produces a summary entry for every extracted dataset,
// and that the exclude pattern keeps out the per-face/per-cell datasets.
func TestDatasetExtractGroupWithPostprocess(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	groupPath := tsBase + "/Boundary Conditions"
	input := RasExtractInput{
		GroupPath:      groupPath,
		Colnames:       []string{"stage", "flow"},
		Postprocess:    []string{"max"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriteSummary:   true,
		WriterType:     JsonWriter,
		WriteBlockName: "bcline_peak",
	}

	if err := DataExtract[float32](input, fpath); err != nil {
		t.Fatalf("DataExtract returned error: %v", err)
	}

	// every extracted dataset gets an entry keyed by its dataset name
	if len(writerAccumulator.Data["bcline_peak"]) == 0 {
		t.Fatal("expected at least one extracted dataset in accumulator block 'bcline_peak'")
	}
	for _, entry := range writerAccumulator.Data["bcline_peak"] {
		for dsname, v := range entry {
			if strings.ContainsAny(dsname, "Flow per Face|Stage per Face|Flow per Cell") {
				t.Errorf("excluded dataset %q was extracted", dsname)
			}
			block, ok := v.(RasExtractorOutputBlock[float32])
			if !ok {
				t.Fatalf("expected RasExtractorOutputBlock[float32] for %s, got %T", dsname, v)
			}
			if block.Dataset != groupPath+"/"+dsname {
				t.Errorf("dataset path = %q, want %q", block.Dataset, groupPath+"/"+dsname)
			}
			if len(block.Summaries) == 0 {
				t.Errorf("dataset %s: expected max summary, got none", dsname)
			}
		}
	}
}

// TestDatasetExtractReflineFloat64 exercises a dataset (not group) extraction
// with the float64 Time dataset.
func TestDatasetExtractReflineFloat64(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	input := RasExtractInput{
		DataPath:    tsBase + "/Time",
		DataType:    reflect.Float64,
		WriterType:  JsonWriter,
		WriteBlockName: "time",
	}

	if err := DataExtract[float64](input, fpath); err != nil {
		t.Fatalf("DataExtract returned error: %v", err)
	}

	entries := writerAccumulator.Data["time"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 accumulated entry, got %d", len(entries))
	}
	// the JSON writer stores the 2D slice under the dataset name
	raw, ok := entries[0]["Time"].([]any)
	if !ok {
		t.Fatalf("expected slice for Time, got %T", entries[0]["Time"])
	}
	if len(raw) == 0 {
		t.Fatal("expected non-empty float64 data")
	}
}

// TestDatasetExtractInt32 exercises int32 extraction against a known dataset
// whose first four values were verified with h5dump (1, 2, 37, 38).
func TestDatasetExtractInt32(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	input := RasExtractInput{
		DataPath:       "Event Conditions/Meteorology/Precipitation/2D Flow Areas/Perimeter 1/Cell Indexes",
		DataType:       reflect.Int32,
		WriterType:     JsonWriter,
		WriteBlockName: "int32",
	}

	if err := DataExtract[int32](input, fpath); err != nil {
		t.Fatalf("DataExtract returned error: %v", err)
	}

	entries := writerAccumulator.Data["int32"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 accumulated entry, got %d", len(entries))
	}
	// the JSON writer stores the 2D slice under the dataset name
	raw, ok := entries[0]["Cell Indexes"].([]any)
	if !ok {
		t.Fatalf("expected slice for Cell Indexes, got %T", entries[0]["Cell Indexes"])
	}
	if len(raw) == 0 {
		t.Fatal("expected non-empty int32 data")
	}
}

// TestDatasetExtractString exercises string extraction against the reference
// line name dataset.
func TestDatasetExtractString(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	input := RasExtractInput{
		DataPath:       tsBase + "/Reference Lines/Name",
		DataType:       reflect.String,
		WriterType:     JsonWriter,
		WriteBlockName: "refline-names",
	}

	if err := DataExtract[string](input, fpath); err != nil {
		t.Fatalf("DataExtract returned error: %v", err)
	}

	entries := writerAccumulator.Data["refline-names"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 accumulated entry, got %d", len(entries))
	}
}

// TestAttributeExtractGroup exercises attribute extraction mode against the
// root group of the results file.
func TestAttributeExtractGroup(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	input := AttributeExtractInput{
		AttributePath:  "/",
		AttributeNames: []string{"File Type", "File Version"},
		WriteBlockName: "file-attrs",
	}

	if err := AttributeExtract(input, fpath); err != nil {
		t.Fatalf("AttributeExtract returned error: %v", err)
	}

	entries := writerAccumulator.Data["file-attrs"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 accumulated entry, got %d", len(entries))
	}
	block, ok := entries[0]["attributes"].(RasExtractorOutputBlock[float32])
	if !ok {
		t.Fatalf("expected RasExtractorOutputBlock[float32], got %T", entries[0]["attributes"])
	}
	if block.Dataset != "/" {
		t.Errorf("dataset = %q, want %q", block.Dataset, "/")
	}
	if _, ok := block.Record["File Version"]; !ok {
		t.Error("expected 'File Version' attribute in record")
	}
}

// TestAttributeExtractSummaryGroup exercises attribute extraction mode against
// the unsteady summary group.
func TestAttributeExtractSummaryGroup(t *testing.T) {
	resetAccumulator(t)
	fpath := duwamishPath(t)

	input := AttributeExtractInput{
		AttributePath:  "/Results/Unsteady/Summary",
		AttributeNames: []string{"Solution"},
		WriteBlockName: "summary-attrs",
	}

	if err := AttributeExtract(input, fpath); err != nil {
		t.Fatalf("AttributeExtract returned error: %v", err)
	}

	entries := writerAccumulator.Data["summary-attrs"]
	if len(entries) != 1 {
		t.Fatalf("expected 1 accumulated entry, got %d", len(entries))
	}
	block, ok := entries[0]["attributes"].(RasExtractorOutputBlock[float32])
	if !ok {
		t.Fatalf("expected RasExtractorOutputBlock[float32], got %T", entries[0]["attributes"])
	}
	if _, ok := block.Record["Solution"]; !ok {
		t.Error("expected 'Solution' attribute in record")
	}
}
