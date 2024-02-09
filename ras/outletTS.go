package ras

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type OutletTS struct {
	Name       string
	RowCount   int
	TimeSeries []FlowData
}
type FlowData struct {
	Index int
	Flow  float64
}

func InitOutletTS(rows []string) (*OutletTS, error) {
	name := rows[0][len(TS_OUTFLOW_HEADER):len(rows[0])]
	output := OutletTS{Name: name}
	rowCount, err := strconv.Atoi(strings.TrimLeft(rows[1], " "))
	if err != nil {
		return &output, err
	}
	output.RowCount = rowCount
	flowdata := make([]FlowData, rowCount)
	for idx, rowstring := range rows {
		if idx != 0 && idx != 1 {
			tmpFlowData, err := parseRowString(rowstring)
			if err != nil {
				return &output, err
			}
			for _, fd := range tmpFlowData {
				flowdata[fd.Index] = fd
			}
		}
	}
	output.TimeSeries = flowdata
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
		index, err := strconv.Atoi(strings.TrimLeft(rowString[0+i*valueLength*2:valueLength+i*valueLength*2], " "))
		if err != nil {
			return []FlowData{}, errors.New("could not parse index")
		}
		flow, err := strconv.ParseFloat(strings.TrimLeft(rowString[valueLength+i*valueLength*2:valueLength*2+i*valueLength*2], " "), 64)
		if err != nil {
			return []FlowData{}, errors.New("could not parse flow")
		}
		result[i] = FlowData{
			Index: index,
			Flow:  flow,
		}
	}
	return result, nil
}
func (ots *OutletTS) UpdateFlows(flows []float64) error {
	if len(ots.TimeSeries) != len(flows) {
		return errors.New("Flow data was not the same length as the target in the b file")
	}
	for idx := range ots.TimeSeries {
		ots.TimeSeries[idx] = FlowData{idx, flows[idx]}
	}
	return nil
}
func (ots *OutletTS) ToBytes() []byte {
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
		result = append(result, fmt.Sprintf("%*d%*f", 8, fd.Index, 8, fd.Flow)...) //this wont be exactly the same. it will right pad with zeros up to 8, mixed precision is hard.
	}
	result = append(result, "\n"...)
	return result
}
