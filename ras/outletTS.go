package ras

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const EndOfFlow = " 3.4E+38"

type OutletTS struct {
	Name       string
	RowCount   int
	TimeSeries []FlowData
	ExtraLines []string
}
type FlowData struct {
	Index float32
	Flow  float32
}

func InitOutletTS(rows []string) (*OutletTS, error) {
	name := rows[0][len(TS_OUTFLOW_HEADER):len(rows[0])]
	output := OutletTS{Name: name}
	stringLines := make([]string, 0)
	rowCount, err := strconv.Atoi(strings.TrimLeft(rows[1], " "))
	if err != nil {
		return &output, err
	}
	output.RowCount = rowCount
	flowdata := make([]FlowData, rowCount)
	hasReachedEndOfFlow := false
	for idx, rowstring := range rows {
		if idx != 0 && idx != 1 {
			// for multiple dams the last line in the returned block is actually the first line of the incoming block - it has text in it.
			if hasReachedEndOfFlow {
				stringLines = append(stringLines, rowstring)
			} else {
				if rowstring == EndOfFlow {
					hasReachedEndOfFlow = true
					stringLines = append(stringLines, rowstring)
				} else {
					tmpFlowData, err := parseRowString(rowstring)
					if err != nil {
						return &output, err
					}
					for jdx, fd := range tmpFlowData {
						flowdata[(idx-2)*5+jdx] = fd
					}
				}
			}
		}
	}
	output.TimeSeries = flowdata
	output.ExtraLines = stringLines
	return &output, nil
}
func parseRowString(rowString string) ([]FlowData, error) {
	valueLength := 8
	rowLength := len(rowString)
	if rowLength == 0 {
		return []FlowData{}, errors.New("row length is zero")
	}
	if rowLength%valueLength != 0 {
		return []FlowData{}, errors.New(CELL_SIZE_ERROR)
	}
	values := rowLength / valueLength
	flowDataCount := values / 2
	result := make([]FlowData, flowDataCount)
	for i := 0; i < flowDataCount; i++ {
		index, err := strconv.ParseFloat(strings.TrimLeft(rowString[0+i*valueLength*2:valueLength+i*valueLength*2], " "), 32)
		index32 := float32(index)
		if err != nil {
			return []FlowData{}, errors.New("could not parse index")
		}
		flow, err := strconv.ParseFloat(strings.TrimLeft(rowString[valueLength+i*valueLength*2:valueLength*2+i*valueLength*2], " "), 32)
		flow32 := float32(flow)
		if err != nil {
			return []FlowData{}, errors.New("could not parse flow")
		}
		result[i] = FlowData{
			Index: index32,
			Flow:  flow32,
		}
	}
	return result, nil
}

func (ots *OutletTS) Header() string {
	return fmt.Sprintf("%v%v\n", TS_OUTFLOW_HEADER, ots.Name)
}

func (ots *OutletTS) UpdateFloat(value float64) error {
	return errors.New("cannot update float on outlet timeseries")
}

func (ots *OutletTS) UpdateFloatArray(values []float32) error {
	if len(ots.TimeSeries) != len(values) {
		return errors.New("flow data was not the same length as the target in the b file")
	}
	for idx, fd := range ots.TimeSeries {
		ots.TimeSeries[idx] = FlowData{fd.Index, values[idx]}
	}
	return nil
}

func (ots *OutletTS) ToBytes() ([]byte, error) {
	result := make([]byte, 0)
	result = append(result, fmt.Sprintf("%v%v\n", TS_OUTFLOW_HEADER, ots.Name)...)
	result = append(result, fmt.Sprintf("%*d\n", 8, ots.RowCount)...)
	//write out 5 pairs then newline
	for idx, fd := range ots.TimeSeries {
		if idx != 0 {
			if idx%5 == 0 {
				result = append(result, "\n"...) //zero based will return on the 6th element (after the 5th) before writing the 6th
			}
		}
		result = append(result, fmt.Sprintf("%s%s", convertFloatToBfileCellValue(float64(fd.Index)), convertFloatToBfileCellValue(float64(fd.Flow)))...) //
	}
	for _, r := range ots.ExtraLines {
		result = append(result, fmt.Sprintf("%s\n", r)...)
	}
	//result = append(result, "\n"...)
	return result, nil
}
