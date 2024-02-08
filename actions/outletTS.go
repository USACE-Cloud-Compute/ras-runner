package actions

import (
	"errors"
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

func InitOutletTS(rows []string) (OutletTS, error) {
	name := rows[0][len(TS_OUTFLOW_HEADER):len(rows[0])]
	output := OutletTS{Name: name}
	rowCount, err := strconv.Atoi(strings.TrimLeft(rows[1], " "))
	if err != nil {
		return output, err
	}
	output.RowCount = rowCount
	flowdata := make([]FlowData, rowCount)
	for idx, rowstring := range rows {
		if idx != 0 || idx != 1 {
			tmpFlowData, err := parseRowString(rowstring)
			if err != nil {
				return output, err
			}
			for _, fd := range tmpFlowData {
				flowdata[fd.Index] = fd
			}
		}
	}
	output.TimeSeries = flowdata
	return output, nil
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
