package hdf

import (
	"bytes"
	"fmt"
	"log"
	"ras-runner/actions/extract"
	"reflect"
	"strings"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/hdf5utils"
)

const TWODSTORAGEAREA_STRUCTUREVARIABLES_RESULT_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/"

func init() {
	cc.ActionRegistry.RegisterAction(&ReadStructureVariablesPeak{
		ActionRunnerBase: cc.ActionRunnerBase{ActionName: "structure-variables-peak-output"},
	})
}

type ReadStructureVariablesPeak struct {
	cc.ActionRunnerBase
}

func (a *ReadStructureVariablesPeak) Run() error {
	dataSourceName := a.Action.Attributes.GetStringOrFail("structurevariablesDataSource")
	startEventIndex := a.Action.Attributes.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := a.Action.Attributes.GetInt64OrFail("end_event_index")
	outputDataSourceName := a.Action.Attributes.GetStringOrFail("output_file_dataSource")
	bucketPrefix := a.Action.Attributes.GetStringOrFail("bucket_prefix")
	twoDFlowarea := a.Action.Attributes.GetStringOrFail("twod_flow_area")
	dataPathString := a.Action.Attributes.GetStringOrFail("twod_hyd_cons")
	dataPaths := strings.Split(dataPathString, ", ")

	hdfDataSource, err := a.PluginManager.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
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
	simulation := extract.SimulationMaxResult{
		DataPaths: header,
		Rows:      []extract.EventMaxResult{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	//index := 0
	rootPath := TWODSTORAGEAREA_STRUCTUREVARIABLES_RESULT_PATH + twoDFlowarea
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
			eventRow := make([]float32, len(dataPaths)*3)
			//0=Total Flow, 2=Stage HW, 3=Stage TW
			cols := []int{0, 2, 3}
			for idx, hydcon := range dataPaths {
				err = func() error {
					datapath := fmt.Sprintf("%s/2D Hyd Conn/%s/Structure Variables", rootPath, hydcon)
					ds, err := hdf5utils.NewHdfDataset(datapath, options)
					if err != nil {
						log.Printf("%v %v", hdfPath, hydcon)
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
	if err != nil {
		return err
	}
	return nil
}
