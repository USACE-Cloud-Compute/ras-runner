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

const (
	BCLINE_RESULT_PATH string = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/"
	STAGE_COLUMN       int    = 0
	FLOW_COLUMN        int    = 0
)

func init() {
	cc.ActionRegistry.RegisterAction("bcline-peak-outputs", &ReadBcLinePeakAction{})
}

type ReadBcLinePeakAction struct {
	cc.ActionRunnerBase
}

func (a *ReadBcLinePeakAction) Run() error {
	//hdf file and data paths are specified by a keyword in the input datasets (since im on the older sdk that doesnt have input datasources in actions.)
	dataSourceName := a.Action.Attributes.GetStringOrFail("bcLineDataSource")
	variableType := a.Action.Attributes.GetStringOrFail("stage_or_flow")
	startEventIndex := a.Action.Attributes.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := a.Action.Attributes.GetInt64OrFail("end_event_index")
	outputDataSourceName := a.Action.Attributes.GetStringOrFail("output_file_dataSource")
	bucketPrefix := a.Action.Attributes.GetStringOrFail("bucket_prefix")
	dataPaths, err := a.Action.Attributes.GetStringSlice("bclines")
	if err != nil {
		return err
	}
	hdfDataSource, err := a.PluginManager.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
	if err != nil {
		return err
	}

	//for bclines stage is column index 0.
	col := STAGE_COLUMN
	if variableType == "flow" {
		col = FLOW_COLUMN
	}

	simulation := extract.SimulationMaxResult{
		DataPaths: dataPaths,
		Rows:      []extract.EventMaxResult{},
	}

	//crack open a hdf file and read the values for each specified datapath.
	//index := 0
	for event := startEventIndex; event <= endEventIndex; event++ {
		err = func() error {
			hdfPath := fmt.Sprintf(hdfDataSource.Paths["0"], event)
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
						log.Printf("%v %v", hdfPath, bcline)
						return err
					}
					defer ds.Close()
					column := []float32{}
					ds.ReadColumn(col, &column)
					var mv float32 = math.SmallestNonzeroFloat32

					for _, v := range column {
						if v >= mv {
							mv = v
						}
					}
					eventRow[idx] = mv
					return nil
				}()
				if err != nil {
					return err
				}
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
			continue
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

	return err
}
