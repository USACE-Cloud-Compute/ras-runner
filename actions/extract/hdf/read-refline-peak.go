package hdf

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"ras-runner/actions/extract"
	"reflect"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/hdf5utils"
)

const REFLINE_RESULT_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/"

func init() {
	cc.ActionRegistry.RegisterAction("refline-peak-outputs", &ReadRefLinePeakAction{})
}

type ReadRefLinePeakAction struct {
	cc.ActionRunnerBase
}

func (a *ReadRefLinePeakAction) Run() error {
	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := a.Action.Attributes.GetStringOrFail("refLineDataSource")
	variableType := a.Action.Attributes.GetStringOrFail("wsel_or_flow")
	startEventIndex := a.Action.Attributes.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := a.Action.Attributes.GetInt64OrFail("end_event_index")
	dsetNameStringLen := a.Action.Attributes.GetIntOrFail("names_string_length")
	outputDataSourceName := a.Action.Attributes.GetStringOrFail("output_file_dataSource")
	bucketPrefix := a.Action.Attributes.GetStringOrFail("bucket_prefix")
	hdfDataSource, err := a.PluginManager.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}
	//for reflines we have Water Surface or Flow
	dsName := "Water Surface"
	if variableType == "flow" {
		dsName = "Flow"
	}

	//eventCount := endEventIndex - startEventIndex

	hdfPath := fmt.Sprintf(hdfDataSource.Paths["0"], startEventIndex)
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

	simulation := extract.SimulationMaxResult{
		DataPaths: dataPaths,
		Rows:      []extract.EventMaxResult{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	for event := startEventIndex; event <= endEventIndex; event++ {
		//read the names from the Names Table.
		err = func() error {
			hdfPath := fmt.Sprintf(hdfDataSource.Paths["0"], event)
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
			bcEventRow := extract.EventMaxResult{
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

	b := simulation.ToBytes()
	reader := bytes.NewReader(b)

	_, err = a.PluginManager.Put(cc.PutOpInput{
		SrcReader: reader,
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: outputDataSourceName,
			PathKey:        "0",
		},
	})
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
