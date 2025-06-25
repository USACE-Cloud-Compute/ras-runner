package hdf

import (
	"bytes"
	"fmt"
	"log"
	"ras-runner/actions/extract"
	"reflect"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/hdf5utils"
)

type ReadRefLineTimeSeriesAction struct {
	cc.ActionRunnerBase
}

func (a *ReadRefLineTimeSeriesAction) Run() error {
	dataSourceName := a.Action.Attributes.GetStringOrFail("refLineDataSource")
	variableType := a.Action.Attributes.GetStringOrFail("wsel_or_flow")
	EventIndex := a.Action.Attributes.GetInt64OrDefault("event_index", 1)
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

	hdfPath := fmt.Sprintf(hdfDataSource.Paths["0"], EventIndex)
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

	result := extract.EventTimeSeriesResult{
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

	outputDataSource, err := a.PluginManager.GetOutputDataSource(outputDataSourceName)
	if err != nil {
		return err
	}
	outputDataSource.Paths["0"] = fmt.Sprintf("%v_event_%v.%v", outputDataSource.Paths["0"], EventIndex, "csv")
	b := result.ToBytes()
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
