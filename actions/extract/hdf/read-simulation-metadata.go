package hdf

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"ras-runner/actions/extract"
	"strings"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

const SUMMARY_PATH = "/Results/Unsteady/Summary"
const TWOD_FLOW_AREA_PATH = "/Results/Unsteady/Output/Output Blocks/Base Output/Summary Output/2D Flow Areas/"

type ReadSimulationMetadata struct {
	cc.ActionRunnerBase
}

func (a *ReadSimulationMetadata) Run() error {
	dataSourceName := a.Action.Attributes.GetStringOrFail("simulationDataSource")
	startEventIndex := a.Action.Attributes.GetInt64OrDefault("start_event_index", 1)
	endEventIndex := a.Action.Attributes.GetInt64OrFail("end_event_index")
	bucketPrefix := a.Action.Attributes.GetStringOrFail("bucket_prefix")
	twoDStorageAreaNameString := a.Action.Attributes.GetStringOrFail("flow_areas")
	twoDStorageAreaNames := strings.Split(twoDStorageAreaNameString, ", ")
	outputDataSourceName := a.Action.Attributes.GetStringOrFail("output_file_dataSource")
	hdfDataSource, err := a.PluginManager.GetInputDataSource(dataSourceName) // expected to look something like this "https://bucket-name.s3.re-gio-n.amazonaws.com/model-library/ffrd-duwamish/simulations/validation/%v/Hydraulics/Duwamish_17110013.p01.hdf"
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
	simulation := extract.SimulationMetadata{
		DataPaths: dataPaths,
		Rows:      []extract.EventMetadata{},
	}
	//crack open a hdf file and read the values for each specified datapath.
	for event := startEventIndex; event <= endEventIndex; event++ {
		//read the HDF file.
		hdfPath := fmt.Sprintf(hdfDataSource.Paths["0"], event)
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
			compTime, err := getAttribute[string](dataPaths[0], simulationGroup)
			if err != nil {
				return err
			}
			values[0] = compTime
			maxWSELError, err := getAttribute[float32](dataPaths[1], simulationGroup)
			if err != nil {
				return err
			}
			values[1] = maxWSELError
			solution, err := getAttribute[float32](dataPaths[2], simulationGroup)
			if err != nil {
				return err
			}
			values[2] = solution
			timeUnstable, err := getAttribute[float32](dataPaths[3], simulationGroup)
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
					val, err := getAttribute[float32](attrName, twodGroup)
					if err != nil {
						return err
					}
					values[index] = val
				}
			}
			bcEventRow := extract.EventMetadata{
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

func getAttribute[T any](name string, grp *hdf5.Group) (T, error) {
	var val T
	if grp.AttributeExists(name) {
		Attr, err := grp.OpenAttribute(name)
		if err != nil {
			return val, err
		}
		defer Attr.Close()
		//Attr.Type()
		Attr.Read(&val, hdf5.T_IEEE_F32LE)
		return val, nil
	}
	return val, errors.New("attribute named " + name + " does not exist")
}

/*
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
*/
