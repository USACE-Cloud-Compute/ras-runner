package extract

import (
	"fmt"
	"strings"
)

type EventMaxResult struct {
	EventId   int64
	DataPaths *[]string
	Values    []float32
}

type SimulationMaxResult struct {
	DataPaths []string
	Rows      []EventMaxResult
}

func (bclsm SimulationMaxResult) ToBytes() []byte {

	builder := strings.Builder{}
	header := fmt.Sprintf("Event ID, %v\n", strings.Join(bclsm.DataPaths, ", "))
	builder.WriteString(header)
	for _, row := range bclsm.Rows {
		builder.WriteString(fmt.Sprintf("%v", row.EventId))
		for _, value := range row.Values {
			builder.WriteString(fmt.Sprintf(",%f", value))
		}
		builder.WriteString("\n")
	}

	return []byte(builder.String())
}

type EventMetadata struct {
	EventId   int64
	DataPaths *[]string
	Values    []any
}
type SimulationMetadata struct {
	DataPaths []string
	Rows      []EventMetadata
}

func (bclsm SimulationMetadata) ToBytes() []byte {

	builder := strings.Builder{}
	header := fmt.Sprintf("Event ID, %v\n", strings.Join(bclsm.DataPaths, ", "))
	builder.WriteString(header)
	for _, row := range bclsm.Rows {
		builder.WriteString(fmt.Sprintf("%v", row.EventId))
		for _, value := range row.Values {
			builder.WriteString(fmt.Sprintf(",%v", value))
		}
		builder.WriteString("\n")
	}

	return []byte(builder.String())
}

type EventTimeSeriesResult struct {
	EventId   int64
	DataPaths []string
	Values    [][]float32
}

func (etsr EventTimeSeriesResult) ToBytes() []byte {

	builder := strings.Builder{}
	header := fmt.Sprintf("TimeStep, %v\n", strings.Join(etsr.DataPaths, ", "))
	builder.WriteString(header)

	for j := range etsr.Values[0] {
		builder.WriteString(fmt.Sprintf("%v", j))
		for i := range etsr.Values {
			builder.WriteString(fmt.Sprintf(",%f", etsr.Values[i][j]))
		}
		builder.WriteString("\n")
	}
	return []byte(builder.String())
}
