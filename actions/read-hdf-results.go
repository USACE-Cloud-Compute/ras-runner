package actions

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/hdf5utils"
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

const BCLINE_RESULT_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/"

// ReadBCLinePeakStage reads the peak stage for each bc line element provided.
func ReadBCLinePeak(action cc.Action, modelDir string) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("bcLineDataSource")
	variableType := action.Parameters.GetStringOrFail("stage_or_flow")
	startEventIndex := action.Parameters.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := action.Parameters.GetInt64OrFail("end_event_index")
	outputDataSourceName := action.Parameters.GetStringOrFail("output_file_dataSource")
	bucketPrefix := action.Parameters.GetStringOrFail("bucket_prefix")
	dataPaths, err := action.Parameters.GetStringSlice("bclines")
	if err != nil {
		return err
	}
	hdfDataSource, err := pm.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}
	//for bclines stage is column index 0.
	col := 0
	if variableType == "flow" {
		col = 1
	}

	//eventCount := endEventIndex - startEventIndex
	simulation := SimulationMaxResult{
		DataPaths: dataPaths,
		Rows:      []EventMaxResult{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	//index := 0
	for event := startEventIndex; event <= endEventIndex; event++ {
		err = func() error {
			hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], event)
			log.Println("searching for " + hdfPath)
			f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)

			if err != nil {
				return err
			}
			defer f.Close()
			options := hdf5utils.HdfReadOptions{
				Dtype:        reflect.Float32,
				ReadOnCreate: true,
				File:         f,
			}
			eventRow := make([]float32, len(dataPaths))

			for idx, bcline := range dataPaths {
				err = func() error {
					datapath := fmt.Sprintf("%s/%s", BCLINE_RESULT_PATH, bcline)
					ds, err := hdf5utils.NewHdfDataset(datapath, options)
					if err != nil {
						log.Println(fmt.Sprintf("%v %v", hdfPath, bcline))
						return err
					}
					defer ds.Close()
					column := []float32{}
					ds.ReadColumn(col, &column)
					var mv float32 = -901.0

					for _, v := range column {
						//fmt.Printf("%f\n", v)
						if v >= mv {
							mv = v
						}
					}
					eventRow[idx] = mv
					return nil
				}()
				if err != nil {
					log.Fatal(err)
				}
			}
			bcEventRow := EventMaxResult{
				EventId:   event,
				DataPaths: &simulation.DataPaths,
				Values:    eventRow,
			}
			simulation.Rows = append(simulation.Rows, bcEventRow)
			return nil
		}()
		if err != nil {
			continue
		}
	}
	outputDataSource, err := pm.GetOutputDataSource(outputDataSourceName)
	if err != nil {
		return err
	}
	b := simulation.ToBytes()
	reader := bytes.NewReader(b)
	//fmt.Println(string(b))
	err = pm.FileWriter(reader, outputDataSource, 0)
	//err = pm.PutFile(b, outputDataSource, 0)
	if err != nil {
		return err
	}
	return nil

}

const REFLINE_RESULT_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/"

// ReadRefLinePeakStage reads the peak stage for each bc line element provided.
func ReadRefLinePeak(action cc.Action, modelDir string) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("refLineDataSource")
	variableType := action.Parameters.GetStringOrFail("wsel_or_flow")
	startEventIndex := action.Parameters.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := action.Parameters.GetInt64OrFail("end_event_index")
	dsetNameStringLen := action.Parameters.GetIntOrFail("names_string_length")
	outputDataSourceName := action.Parameters.GetStringOrFail("output_file_dataSource")
	bucketPrefix := action.Parameters.GetStringOrFail("bucket_prefix")
	hdfDataSource, err := pm.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}
	//for reflines we have Water Surface or Flow
	dsName := "Water Surface"
	if variableType == "flow" {
		dsName = "Flow"
	}

	//eventCount := endEventIndex - startEventIndex

	hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], startEventIndex)
	f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
	if err != nil {
		return err
	}
	namesDataSet, err := hdf5utils.NewHdfDataset(REFLINE_RESULT_PATH+"Name", hdf5utils.HdfReadOptions{
		Dtype:        reflect.String,
		Strsizes:     hdf5utils.NewHdfStrSet(dsetNameStringLen),
		File:         f,
		ReadOnCreate: true,
	})
	if err != nil {
		return err
	}
	defer namesDataSet.Close()
	dataPaths := make([]string, namesDataSet.Rows())
	for i := 0; i < namesDataSet.Rows(); i++ {
		name := []string{}
		err := namesDataSet.ReadRow(i, &name)
		if err != nil {
			return err
		}
		dataPaths[i] = name[0]
	}

	simulation := SimulationMaxResult{
		DataPaths: dataPaths,
		Rows:      []EventMaxResult{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	for event := startEventIndex; event <= endEventIndex; event++ {
		//read the names from the Names Table.
		err = func() error {
			hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], event)
			f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
			if err != nil {
				log.Println(hdfPath + " not found")
				return err
			}
			defer f.Close()
			var destVals *hdf5utils.HdfDataset
			err = func() error {
				destoptions := hdf5utils.HdfReadOptions{
					Dtype:        reflect.Float32,
					File:         f,
					ReadOnCreate: true,
				}
				destVals, err = hdf5utils.NewHdfDataset(REFLINE_RESULT_PATH+dsName, destoptions)
				if err != nil {
					return err
				}
				defer destVals.Close()
				return nil
			}()
			if err != nil {
				return err
			}
			eventRow := make([]float32, len(dataPaths))
			column := []float32{}
			for idx := range dataPaths {
				destVals.ReadColumn(idx, &column)
				var mv float32 = 0.0
				for _, v := range column {
					if mv <= v {
						mv = v
					}
				}
				eventRow[idx] = mv
			}
			bcEventRow := EventMaxResult{
				EventId:   event,
				DataPaths: &simulation.DataPaths,
				Values:    eventRow,
			}
			simulation.Rows = append(simulation.Rows, bcEventRow)
			return nil
		}()
		if err != nil {
			continue
		}

	}
	outputDataSource, err := pm.GetOutputDataSource(outputDataSourceName)
	if err != nil {
		return err
	}
	b := simulation.ToBytes()
	reader := bytes.NewReader(b)

	err = pm.FileWriter(reader, outputDataSource, 0)
	//err = pm.PutFile(b, outputDataSource, 0)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
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
		builder.WriteString(string(row.EventId))
		for _, value := range row.Values {
			builder.WriteString(fmt.Sprintf(",%v", value))
		}
		builder.WriteString("\n")
	}

	return []byte(builder.String())
}

/*
const SUMMARY_PATH = "/Results/Unsteady/Summary"
const TWOD_FLOW_AREA_PATH = "/Results/Unsteady/Output/Output Blocks/Summary Output/2D Flow Areas/"

func ReadSimulationMetadata(action cc.Action, modelDir string) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("simulationDataSource")
	startEventIndex := action.Parameters.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := action.Parameters.GetInt64OrFail("end_event_index")
	twoDStorageAreaNames, err := action.Parameters.GetStringSlice("2d_flow_areas")
	if err != nil {
		return err
	}
	outputDataSourceName := action.Parameters.GetStringOrFail("output_file_dataSource")
	hdfDataSource, err := pm.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}

	eventCount := endEventIndex - startEventIndex

	dataPaths := []string{"Computation Time Total", "Maximum WSEL Error", "Solution", "Time Stamp Solution Went Unstable"}
	for _, area := range twoDStorageAreaNames {
		dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, "Cumulative Net PRecip Inches"))
		dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, "Volume Accounting Error"))
		dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, "Volume Accounting Error Percentage"))
		dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, "Volume Accounting External Inflow"))
		dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, "Volume Accounting External Outflow"))
		dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, "Volume Accounting Inflow from Net Precipitation"))
	}

	simulation := SimulationMaxResult{
		DataPaths: dataPaths,
		Rows:      make([]EventMaxResult, eventCount),
	}
	//crack open a hdf file and read the values for each specified datapath.
	for event := startEventIndex; event <= endEventIndex; event++ {
		//read the names from the Names Table.
		hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], event)
		f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
		if err != nil {
			log.Println(hdfPath + " not found")
			continue
		}
		var destVals *hdf5utils.HdfDataset
		err = func() error {
			destoptions := hdf5utils.HdfReadOptions{
				Dtype:        reflect.Float32,
				File:         f,
				ReadOnCreate: true,
			}
			destVals, err = hdf5utils.NewHdfDataset(REFLINE_RESULT_PATH+dsName, destoptions)
			if err != nil {
				return err
			}
			defer destVals.Close()
			return nil
		}()
		if err != nil {
			return err
		}
		eventRow := make([]float32, len(dataPaths))
		column := []float32{}
		for idx := range dataPaths {
			destVals.ReadColumn(idx, column)
			var mv float32 = 0.0
			for _, v := range column {
				if mv <= v {
					mv = v
				}
			}
			eventRow[idx] = mv
		}
		bcEventRow := EventMaxResult{
			EventId:   event,
			DataPaths: &simulation.DataPaths,
			Values:    eventRow,
		}
		simulation.Rows[startEventIndex-event] = bcEventRow
	}
	outputDataSource, err := pm.GetInputDataSource(outputDataSourceName)
	if err != nil {
		return err
	}
	b := simulation.ToBytes()
	pm.PutFile(b, outputDataSource, 0)
	return nil
}
*/
