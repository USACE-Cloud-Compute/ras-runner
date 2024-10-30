package actions

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strings"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
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
func ReadBCLinePeak(action cc.Action) error {
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
					var mv float32 = math.SmallestNonzeroFloat32

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
func ReadRefLinePeak(action cc.Action) error {
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
				var mv float32 = math.SmallestNonzeroFloat32
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
			log.Println(err)
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

const REFPOINT_RESULT_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/"

// ReadRefLinePeakStage reads the peak stage for each bc line element provided.
func ReadRefPointPeak(action cc.Action) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("refPointDataSource")
	variableType := action.Parameters.GetStringOrFail("wsel_or_velocity")
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
	if variableType == "velocity" {
		dsName = "Velocity"
	}

	//eventCount := endEventIndex - startEventIndex

	hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], startEventIndex)
	f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
	if err != nil {
		return err
	}
	namesDataSet, err := hdf5utils.NewHdfDataset(REFPOINT_RESULT_PATH+"Name", hdf5utils.HdfReadOptions{
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
				destVals, err = hdf5utils.NewHdfDataset(REFPOINT_RESULT_PATH+dsName, destoptions)
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
				var mv float32 = math.SmallestNonzeroFloat32
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
			log.Println(err)
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
func ReadRefPointMinimum(action cc.Action) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("refPointDataSource")
	variableType := action.Parameters.GetStringOrFail("wsel_or_velocity")
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
	if variableType == "velocity" {
		dsName = "Velocity"
	}

	//eventCount := endEventIndex - startEventIndex

	hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], startEventIndex)
	f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
	if err != nil {
		return err
	}
	namesDataSet, err := hdf5utils.NewHdfDataset(REFPOINT_RESULT_PATH+"Name", hdf5utils.HdfReadOptions{
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
				destVals, err = hdf5utils.NewHdfDataset(REFPOINT_RESULT_PATH+dsName, destoptions)
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
				var mv float32 = math.MaxFloat32
				for _, v := range column {
					if mv >= v {
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
			log.Println(err)
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
		builder.WriteString(fmt.Sprintf("%v", row.EventId))
		for _, value := range row.Values {
			builder.WriteString(fmt.Sprintf(",%v", value))
		}
		builder.WriteString("\n")
	}

	return []byte(builder.String())
}

const SUMMARY_PATH = "/Results/Unsteady/Summary"
const TWOD_FLOW_AREA_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Summary Output/2D Flow Areas/"

func ReadSimulationMetadata(action cc.Action) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("simulationDataSource")
	startEventIndex := action.Parameters.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := action.Parameters.GetInt64OrFail("end_event_index")
	bucketPrefix := action.Parameters.GetStringOrFail("bucket_prefix")
	twoDStorageAreaNameString := action.Parameters.GetStringOrFail("flow_areas")
	twoDStorageAreaNames := strings.Split(twoDStorageAreaNameString, ", ")
	outputDataSourceName := action.Parameters.GetStringOrFail("output_file_dataSource")
	hdfDataSource, err := pm.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}

	dataPaths := []string{"Computation Time Total", "Maximum WSEL Error", "Solution", "Time Stamp Solution Went Unstable"}
	twoDpaths := []string{"Cum Net Precip Inches", "Vol Accounting Error", "Vol Accounting Error Percentage", "Vol Accounting External Inflow", "Vol Accounting External Outflow", "Vol Acct. Inflow from Net Precip"}
	for _, area := range twoDStorageAreaNames {
		for _, variable := range twoDpaths {
			dataPaths = append(dataPaths, fmt.Sprintf("%v - %v", area, variable))
		}
	}
	simulation := SimulationMetadata{
		DataPaths: dataPaths,
		Rows:      []EventMetadata{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	for event := startEventIndex; event <= endEventIndex; event++ {
		//read the HDF file.
		hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], event)
		values := make([]any, 6*len(twoDStorageAreaNames)+4)
		err = func() error {

			f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
			if err != nil {
				return err
			}
			defer f.Close()
			//read the Simulation group
			simulationGroup, err := f.OpenGroup(SUMMARY_PATH)
			if err != nil {
				return err
			}
			defer simulationGroup.Close()
			//read the attributes from the simulation group:
			compTime, err := stringAttribute(dataPaths[0], simulationGroup)
			if err != nil {
				return err
			}
			values[0] = compTime
			maxWSELError, err := floatAttribute(dataPaths[1], simulationGroup)
			if err != nil {
				return err
			}
			values[1] = maxWSELError
			solution, err := stringAttribute(dataPaths[2], simulationGroup)
			if err != nil {
				return err
			}
			values[2] = solution
			timeUnstable, err := stringAttribute(dataPaths[3], simulationGroup)
			if err != nil {
				return err
			}
			values[3] = timeUnstable

			for twodid, name := range twoDStorageAreaNames {
				idx := (twodid * 6) + 4
				twodGroup, err := f.OpenGroup(TWOD_FLOW_AREA_PATH + name)
				if err != nil {
					return err
				}
				defer twodGroup.Close()
				for i, attrName := range twoDpaths {
					index := idx + i
					val, err := floatAttribute(attrName, twodGroup)
					if err != nil {
						return err
					}
					values[index] = val
				}
			}
			bcEventRow := EventMetadata{
				EventId:   event,
				DataPaths: &simulation.DataPaths,
				Values:    values,
			}
			simulation.Rows = append(simulation.Rows, bcEventRow)
			return nil
		}()
		if err != nil {
			log.Println(hdfPath + " not found")
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

	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
func floatAttribute(name string, grp *hdf5.Group) (float32, error) {
	if grp.AttributeExists(name) {
		Attr, err := grp.OpenAttribute(name)
		if err != nil {
			return 0.0, err
		}
		defer Attr.Close()
		var AttrData float32
		//log.Println(Attr.Type().String())
		Attr.Read(&AttrData, hdf5.T_IEEE_F32LE) //not sure if this is a string
		//set value in the values array.
		return AttrData, nil
	}
	return 0.0, errors.New("attribute named " + name + " does not exist")
}
func stringAttribute(name string, grp *hdf5.Group) (string, error) {
	if grp.AttributeExists(name) {
		Attr, err := grp.OpenAttribute(name)
		if err != nil {
			return "", err
		}
		defer Attr.Close()
		var AttrData string
		Attr.Read(&AttrData, hdf5.T_GO_STRING) //not sure if this is a string
		//set value in the values array.
		return AttrData, nil
	}
	return "", errors.New("attribute named " + name + " does not exist")
}

const TWODSTORAGEAREA_STRUCTUREVARIABLES_RESULT_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/"

// ReadBCLinePeakStage reads the peak stage for each bc line element provided.
func ReadStructureVariablesPeak(action cc.Action) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("structurevariablesDataSource")
	startEventIndex := action.Parameters.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := action.Parameters.GetInt64OrFail("end_event_index")
	outputDataSourceName := action.Parameters.GetStringOrFail("output_file_dataSource")
	bucketPrefix := action.Parameters.GetStringOrFail("bucket_prefix")
	twoDFlowarea := action.Parameters.GetStringOrFail("twod_flow_area")
	dataPathString := action.Parameters.GetStringOrFail("twod_hyd_cons")
	dataPaths := strings.Split(dataPathString, ", ")
	if err != nil {
		return err
	}
	hdfDataSource, err := pm.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}
	header := make([]string, len(dataPaths)*3)
	for idx, hydcon := range dataPaths {
		header[idx*3] = hydcon + " - Total Flow"
		header[(idx*3)+1] = hydcon + " - Stage HW"
		header[(idx*3)+2] = hydcon + " - Stage TW"
	}
	//eventCount := endEventIndex - startEventIndex
	simulation := SimulationMaxResult{
		DataPaths: header,
		Rows:      []EventMaxResult{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	//index := 0
	rootPath := TWODSTORAGEAREA_STRUCTUREVARIABLES_RESULT_PATH + twoDFlowarea
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
			eventRow := make([]float32, len(dataPaths)*3)
			//0=Total Flow, 2=Stage HW, 3=Stage TW
			cols := []int{0, 2, 3}
			for idx, hydcon := range dataPaths {
				err = func() error {
					datapath := fmt.Sprintf("%s/2D Hyd Conn/%s/Structure Variables", rootPath, hydcon)
					ds, err := hdf5utils.NewHdfDataset(datapath, options)
					if err != nil {
						log.Println(fmt.Sprintf("%v %v", hdfPath, hydcon))
						return err
					}
					defer ds.Close()
					column := []float32{}
					for i, col := range cols {
						ds.ReadColumn(col, &column)
						var mv float32 = -901.0

						for _, v := range column {
							//fmt.Printf("%f\n", v)
							if v >= mv {
								mv = v
							}
						}
						eventRow[(idx*3)+i] = mv
					}

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

func ReadRefLineTimeSeries(action cc.Action) error {
	//get the plugin manager
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}

	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := action.Parameters.GetStringOrFail("refLineDataSource")
	variableType := action.Parameters.GetStringOrFail("wsel_or_flow")
	EventIndex := action.Parameters.GetInt64OrDefault("event_index", 1)
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

	hdfPath := fmt.Sprintf(hdfDataSource.Paths[0], EventIndex)
	f, err := hdf5utils.OpenFile(hdfPath, bucketPrefix)
	if err != nil {
		return err
	}
	defer f.Close()
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

	result := EventTimeSeriesResult{
		DataPaths: dataPaths,
		Values:    [][]float32{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	//read the names from the Names Table.
	err = func() error {

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
		data := make([][]float32, len(dataPaths))

		for idx := range dataPaths {
			column := []float32{}
			destVals.ReadColumn(idx, &column)
			data[idx] = column
		}
		result.Values = data
		return nil
	}()
	if err != nil {
		log.Println(err)
	}

	outputDataSource, err := pm.GetOutputDataSource(outputDataSourceName)
	if err != nil {
		return err
	}
	outputDataSource.Paths[0] = fmt.Sprintf("%v_event_%v.%v", outputDataSource.Paths[0], EventIndex, "csv")
	b := result.ToBytes()
	reader := bytes.NewReader(b)

	err = pm.FileWriter(reader, outputDataSource, 0)
	//err = pm.PutFile(b, outputDataSource, 0)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
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
