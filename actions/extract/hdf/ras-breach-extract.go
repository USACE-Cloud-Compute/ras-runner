package hdf

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"reflect"

	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const (
	timeStepInDaysPath string = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Time"
	breachPathTemplate string = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/%s/2D Hyd Conn"
	breachDataPath     string = "/%s/Breaching Variables"
	//breachFlowDataPath          string  = "/%s/HW TW Segments/Flow"
	BreachFlowColumn            int     = 6
	BreachVelocityFpsColumn     int     = 7
	BottomWidthColumn           int     = 2
	HwColumn                    int     = 0
	TwColumn                    int     = 1
	BreachFlowVelocityThreshold float32 = 1.5
)

type BreachData struct {
	BreachAt               string
	BreachAtTime           float32
	BreachingVariablesData [][]float32
	TimeInDays             []float64
}

type RasBreach struct {
	f *hdf5.File
	//rasversion string
}

func NewRasBreachData(filepath string) (*RasBreach, error) {
	f, err := hdf5utils.OpenFile(filepath)
	if err != nil {
		return nil, err
	}
	rbd := RasBreach{f: f}

	return &rbd, nil
}

func (rb *RasBreach) Close() {
	rb.f.Close()
}

func (rb *RasBreach) FlowAreas2D() ([]string, error) {
	groupPath := "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas"
	group, err := hdf5utils.NewHdfGroup(rb.f, groupPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the hdf group '%s': %s", groupPath, err)
	}
	defer group.Close()
	return group.ObjectNames()
}

func (rb *RasBreach) ConnectionNames(name string) ([]string, error) {
	path := fmt.Sprintf(breachPathTemplate, name)

	gp, err := rb.f.OpenGroup(path)
	if err != nil {
		return nil, err
	}
	defer gp.Close()

	numobj, err := gp.NumObjects()
	if err != nil {
		return nil, err
	}
	connNames := make([]string, numobj)
	for obj := 0; obj < int(numobj); obj++ {
		t, err := gp.ObjectTypeByIndex(uint(obj))
		if err != nil {
			return nil, err
		}
		if t == hdf5.H5G_GROUP {
			gname, err := gp.ObjectNameByIndex(uint(obj))
			if err != nil {
				return nil, err
			}
			connNames[obj] = gname
		}
	}
	return connNames, err
}

func (rb *RasBreach) BreachData(name string, connName string) (BreachData, error) {
	bd := BreachData{}
	datapath := fmt.Sprintf(breachPathTemplate, name) + fmt.Sprintf(breachDataPath, connName)
	err := rb.readBreachAttributes(&bd, datapath)
	if err != nil {
		return bd, errors.New("no breach data")
	} else {
		data, err := rb.readBreachData(datapath)
		if err != nil {
			return bd, errors.New("missing or unreadable breach flow data")
		}
		bd.BreachingVariablesData = data

		//if there is data, there should ALWAYS be timesteps, so I'm logging but not handling timestep read errors.
		timesteps, err := rb.readTimeSteps()
		if err != nil {
			log.Printf("Failed to read timesteps for %s->%s: %s\n", err, name, connName)
		}
		bd.TimeInDays = timesteps
	}
	return bd, nil
}

func (rb *RasBreach) readBreachAttributes(bd *BreachData, datapath string) error {
	ds, err := rb.f.OpenDataset(datapath)
	if err != nil {
		return err
	}
	defer ds.Close()
	var breachAt string
	var breachTime float32

	getattr(ds, "Breach at", &breachAt) //ignore errors and return default on error
	bd.BreachAt = breachAt
	getattr(ds, "Breach at Time (Days)", &breachTime) //ignore errors and return default on error
	bd.BreachAtTime = breachTime
	return nil
}

func (rb *RasBreach) readTimeSteps() ([]float64, error) {
	options := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float64,
		ReadOnCreate: true,
		File:         rb.f,
	}

	data, err := hdf5utils.NewHdfDataset(timeStepInDaysPath, options)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	//Timesteps are a float64 array buffer, so I can cast the buffer rather than reading it into another typed array
	return *(data.Data.Buffer.(*[]float64)), err
}

func (rb *RasBreach) readBreachData(datapath string) ([][]float32, error) {
	options := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float32,
		ReadOnCreate: true,
		File:         rb.f,
	}

	data, err := hdf5utils.NewHdfDataset(datapath, options)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	breachingVarsData := make([][]float32, data.Cols())
	for c := 0; c < data.Cols(); c++ {
		coldata := make([]float32, data.Rows())
		err = data.ReadColumn(c, &coldata)
		breachingVarsData[c] = coldata
	}
	return breachingVarsData, err
}

func getattr(ds *hdf5.Dataset, attrname string, dest any) error {
	attr, err := ds.OpenAttribute(attrname)
	if err != nil {
		return err
	}
	defer attr.Close()
	destType := reflect.TypeOf(dest).Elem()
	switch destType.Kind() {
	case reflect.String:
		return attr.Read(dest, hdf5.T_GO_STRING)
	default:
		dtype, err := hdf5.NewDataTypeFromType(destType)
		if err != nil {
			return err
		}
		return attr.Read(dest, dtype)
	}
}

/////////////////////////////////////////////////////////////////

type BreachRecord struct {
	Event                     string
	FlowArea2D                string
	SaConn                    string
	Breached                  bool
	BreachStartTime           float32
	BreachIndex               int
	MaxHW                     float32
	MaxTW                     float32
	MaxFlow                   float32
	MaxBottomWidth            float32
	BreachProgressionDuration float64
	HWAtBreach                float32
	TWAtBreach                float32
}

var recordHeaders []string = []string{
	"Event",
	"2D_Flow_Area",
	"Connection",
	"Breached",
	"BreachStartTime",
	"BreachIndex",
	"MaxHW",
	"HwAtBreach",
	"MaxTw",
	"TwAtBreach",
	"MaxFlow",
	"MaxBottomWidth",
	"BreachProgressionDuration",
}

func GetBreachRecord(event string, flowarea2d string, connectionname string, bd *BreachData) BreachRecord {
	br := BreachRecord{}
	br.FlowArea2D = flowarea2d
	br.SaConn = connectionname
	breachTimeIndex := breachAtTimeIndex(bd)
	if bd != nil {
		br.Event = event
		br.Breached = !math.IsNaN(float64(bd.BreachAtTime))
		br.BreachStartTime = bd.BreachAtTime
		br.BreachIndex = breachTimeIndex
		br.MaxHW = Max(bd.BreachingVariablesData[HwColumn])
		br.MaxTW = Max(bd.BreachingVariablesData[TwColumn])
		if breachTimeIndex == -1 {
			br.HWAtBreach = float32(math.NaN())
			br.TWAtBreach = float32(math.NaN())
		} else {
			br.HWAtBreach = bd.BreachingVariablesData[HwColumn][breachTimeIndex]
			br.TWAtBreach = bd.BreachingVariablesData[TwColumn][breachTimeIndex]
		}
		br.MaxBottomWidth = Max(bd.BreachingVariablesData[BottomWidthColumn])
		br.MaxFlow = Max(bd.BreachingVariablesData[BreachFlowColumn])
		br.BreachProgressionDuration = getThresholdDuration(bd.BreachingVariablesData[BreachVelocityFpsColumn], bd.TimeInDays, BreachFlowVelocityThreshold)
	}
	return br
}

func breachAtTimeIndex(bd *BreachData) int {
	if math.IsNaN(float64(bd.BreachAtTime)) {
		return -1
	}
	for i := 1; i < len(bd.TimeInDays); i++ {
		if bd.TimeInDays[i-1] <= float64(bd.BreachAtTime) && bd.TimeInDays[i] >= float64(bd.BreachAtTime) {
			return i - 1
		}
	}
	return -1
}

func getThresholdDuration(data []float32, time []float64, threshold float32) float64 {
	startIndex := 0
	endIndex := 0
	for i, v := range data {
		if v > threshold && startIndex == 0 {
			startIndex = i
		}
		if v < threshold && startIndex > 0 {
			endIndex = i
			break
		}
	}
	return time[endIndex] - time[startIndex]
}

type Number interface {
	float64 | float32 | int | int8 | int16 | int32 | int64
}

func Max[T Number](t []T) T {
	max := t[0]
	for _, v := range t {
		mf := float64(max)
		if math.IsNaN(mf) || v > max {
			max = v
		}
	}
	return max
}

type BreachRecordWriter interface {
	Write(record BreachRecord) error
	Close()
}

type CsvBreachRecordWriter struct {
	writer io.WriteCloser
}

func NewCsvBreachRecordWriter(csvPath string) (*CsvBreachRecordWriter, error) {
	writer, err := os.Create(csvPath)
	if err != nil {
		return nil, err
	}
	bw := CsvBreachRecordWriter{writer}
	bw.writeHeaders()
	return &bw, err
}

func (cbw *CsvBreachRecordWriter) Write(r BreachRecord) error {
	row := fmt.Sprintf("%s,%s,%s,%v,%f,%d,%f,%f,%f,%f,%f,%f,%f\n",
		r.Event,
		r.FlowArea2D,
		r.SaConn,
		r.Breached,
		r.BreachStartTime,
		r.BreachIndex,
		r.MaxHW,
		r.HWAtBreach,
		r.MaxTW,
		r.TWAtBreach,
		r.MaxFlow,
		r.MaxBottomWidth,
		r.BreachProgressionDuration,
	)
	_, err := cbw.writer.Write([]byte(row))
	return err
}

func (cbw *CsvBreachRecordWriter) writeHeaders() {
	for i, h := range recordHeaders {
		if i > 0 {
			cbw.writer.Write([]byte(","))
		}
		cbw.writer.Write([]byte(h))
	}
	cbw.writer.Write([]byte("\n"))

}

func (cbw *CsvBreachRecordWriter) Close() {
	cbw.writer.Close()
}
