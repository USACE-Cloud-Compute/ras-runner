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
	// timeStepInDaysPath is the HDF5 path to the time step data.
	timeStepInDaysPath string = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Time"

	// breachPathTemplate is the template for constructing 2D Hyd Conn dataset paths.
	breachPathTemplate string = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/%s/2D Hyd Conn"

	// breachDataPath is the path suffix for breaching variables data.
	breachDataPath string = "/%s/Breaching Variables"

	BreachFlowColumnName     string = "Breach Flow"
	BreachVelocityColumnName string = "Breach Velocity"
	BottomWidthColumnName    string = "Bottom-Width"
	HwColumnName             string = "Stage HW"
	TwColumnName             string = "Stage TW"

	// // BreachFlowColumn is the column index for flow data in breaching variables.
	// BreachFlowColumn int = 6

	// // BreachVelocityFpsColumn is the column index for velocity data in breaching variables.
	// BreachVelocityFpsColumn int = 7

	// // BottomWidthColumn is the column index for bottom width data in breaching variables.
	// BottomWidthColumn int = 2

	// // HwColumn is the column index for HW (Stage) data in breaching variables.
	// HwColumn int = 0

	// // TwColumn is the column index for TW (Stage) data in breaching variables.
	// TwColumn int = 1

	// BreachFlowVelocityThreshold is the velocity threshold to determine breach progression duration.
	BreachFlowVelocityThreshold float32 = 1.5
	BreachFields                string  = "Variable_Unit"
)

// BreachData represents the breach data extracted from an HDF5 file.
type BreachData struct {
	// BreachAt is the RAS attr location where the breach occurred.
	BreachAt string

	// BreachAtTime is the RAS attr time (in days) when the breach occurred.
	BreachAtTime float32

	// BreachingVariablesData contains the breaching variables data.
	BreachingVariablesData [][]float32

	// TimeInDays contains the time steps in days.
	TimeInDays []float64

	//private variable for Variable_Units field definitions
	fields [][]string
}

func (bd *BreachData) ColumnIndexMap() map[string]int {
	bnames := []string{
		BreachFlowColumnName,
		BreachVelocityColumnName,
		BottomWidthColumnName,
		HwColumnName,
		TwColumnName,
	}
	indexMap := make(map[string]int)

	for i, cname := range bd.fields {
		for _, bname := range bnames {
			if cname[0] == bname {
				indexMap[bname] = i
			}
		}
	}

	return indexMap
}

type RasBreach struct {
	f *hdf5.File
	//rasversion string
}

// NewRasBreachData creates a new RasBreach instance from an HDF5 file.
//
// It opens the specified HDF5 file and returns a pointer to RasBreach.
// Returns an error if the file cannot be opened or read.
func NewRasBreachData(filepath string) (*RasBreach, error) {
	f, err := hdf5utils.OpenFile(filepath)
	if err != nil {
		return nil, err
	}
	rbd := RasBreach{f: f}

	return &rbd, nil
}

// Close closes the underlying HDF5 file.
func (rb *RasBreach) Close() {
	rb.f.Close()
}

// FlowAreas2D returns a list of 2D flow area names from the HDF5 file.
//
// It reads the 2D Flow Areas group and returns a slice of strings representing
// each flow area name.
// Returns an error if the group cannot be accessed or read.
func (rb *RasBreach) FlowAreas2D() ([]string, error) {
	groupPath := "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas"
	group, err := hdf5utils.NewHdfGroup(rb.f, groupPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the hdf group '%s': %s", groupPath, err)
	}
	defer group.Close()
	return group.ObjectNames()
}

// ConnectionNames returns a list of connection names for a given 2D flow area.
//
// It reads the specified flow area's 2D Hyd Conn group and returns a slice of
// strings representing each connection name.
// Returns an error if the group cannot be accessed or read.
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

// BreachData returns the breach data for a given flow area and connection.
//
// It reads the breaching variables data for the specified connection within
// the flow area and returns a BreachData struct containing all relevant information.
// Returns an error if the data cannot be read or is missing.
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

// readBreachAttributes reads the attributes of a breach dataset.
//
// It populates the BreachData struct with the "Breach at" and "Breach at Time (Days)" attributes.
// Returns an error if the attributes cannot be read.
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

	fields, err := get2dStringArrayAttr(ds, BreachFields)
	if err != nil {
		return err
	}
	bd.fields = fields

	return nil
}

// readTimeSteps reads the time step data from the HDF5 file.
//
// It returns a slice of float64 values representing the time steps in days.
// Returns an error if the time step data cannot be read.
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

// readBreachData reads the breaching variables data from the HDF5 file.
//
// It returns a 2D slice of float32 values representing the breaching variables data.
// Returns an error if the data cannot be read.
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

// getattr reads an attribute from an HDF5 dataset.
//
// It handles both string and numeric attributes by detecting the type of the destination
// and reading the appropriate data format.
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

func get2dStringArrayAttr(ds *hdf5.Dataset, attrname string) ([][]string, error) {
	attr, err := ds.OpenAttribute(attrname)
	if err != nil {
		return nil, err
	}
	defer attr.Close()
	return attr.ReadFixedStringArray()
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

// GetBreachRecord creates a BreachRecord from the extracted breach data.
//
// It takes event identifier, flow area name, connection name, and breach data
// to construct a complete BreachRecord with all relevant fields populated.
func GetBreachRecord(event string, flowarea2d string, connectionname string, bd *BreachData) BreachRecord {
	br := BreachRecord{}
	br.FlowArea2D = flowarea2d
	br.SaConn = connectionname
	breachTimeIndex := breachAtTimeIndex(bd)
	breachColIndexMap := bd.ColumnIndexMap()
	HwColumn := breachColIndexMap[HwColumnName]
	TwColumn := breachColIndexMap[TwColumnName]
	BottomWidthColumn := breachColIndexMap[BottomWidthColumnName]
	BreachFlowColumn := breachColIndexMap[BreachFlowColumnName]
	BreachVelocityFpsColumn := breachColIndexMap[BreachVelocityColumnName]
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

// breachAtTimeIndex finds the index of the breach time in the time step data.
//
// It returns -1 if no breach occurred or if the breach time is not found in the time steps.
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

// getThresholdDuration calculates the duration for which the velocity exceeds a threshold.
//
// It takes the velocity data, time steps, and threshold value to compute how long
// the velocity remained above the threshold.
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

// Number is a type constraint for numeric types.
type Number interface {
	float64 | float32 | int | int8 | int16 | int32 | int64
}

// Max returns the maximum value in a slice of numbers.
//
// It handles NaN values properly and returns the first non-NaN value if all are NaN.
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

// BreachRecordWriter defines the interface for writing breach records.
type BreachRecordWriter interface {
	Write(record BreachRecord) error
	Close()
}

// NewCsvBreachRecordWriter creates a new CSV breach record writer.
//
// It opens or creates the specified file path for writing CSV data.
type CsvBreachRecordWriter struct {
	writer io.WriteCloser
}

// Write writes a breach record to the CSV file.
//
// It formats the record data as a CSV row and writes it to the underlying writer.
func NewCsvBreachRecordWriter(csvPath string) (*CsvBreachRecordWriter, error) {
	writer, err := os.Create(csvPath)
	if err != nil {
		return nil, err
	}
	bw := CsvBreachRecordWriter{writer}
	bw.writeHeaders()
	return &bw, err
}

// writeHeaders writes the CSV column headers to the output file.
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

// Close closes the underlying CSV writer.
func (cbw *CsvBreachRecordWriter) Close() {
	cbw.writer.Close()
}
