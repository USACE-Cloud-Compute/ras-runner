package main

import "testing"


func TestMain(t *testing.T) {
	main()
}
/*
func TestHDF_ATTRS(t *testing.T) {
	err := MakeRasHdfTmp("/workspaces/cc-ras-runner/testData/Duwamish_17110013.p01.hdf", "output.p01.hdf")
	if err != nil {
		t.Fail()
	}
}
*/

/*
func Test_Read_HDF_Attributes(t *testing.T) {
	// set the manifest id environment variable.
	os.Setenv("CC_MANIFEST_ID", "99041f15-8274-4782-b67c-bf6216e9fd95")
	action := cc.Action{
		Name:        "simulation-attribute-metadata",
		Type:        "simulation-attribute-metadata", //ptr[string]("copy-inputs"),
		Description: "simulation-attribute-metadata",
		Parameters: map[string]any{
			"simulationDataSource":   "Duwamish_17110013.p01.hdf",
			"start_event_index":      1,
			"end_event_index":        2,
			"flow_areas":             "Perimeter 1",
			"output_file_dataSource": "simulation_metadata.csv",
			"bucket_prefix":          "FFRD",
		},
	}
	actions.ReadSimulationMetadata(action)
}
func Test_Read_HDF_StructureVariables(t *testing.T) {
	// set the manifest id environment variable.
	//os.Setenv("CC_MANIFEST_ID", "99041f15-8274-4782-b67c-bf6216e9fd95")
	action := cc.Action{
		Name:        "structure-variables-peak-output",
		Type:        "structure-variables-peak-output", //ptr[string]("copy-inputs"),
		Description: "structure-variables-peak-output",
		Parameters: map[string]any{
			"structurevariablesDataSource": "Duwamish_17110013.p01.hdf",
			"start_event_index":            1,
			"end_event_index":              2,
			"twod_flow_area":               "Perimeter 1",
			"twod_hyd_cons":                "GS_200thStreet, GS_277thStreet, GS_Auburn, GS_BigSoos, GS_BlackDiamond, GS_GolfCourse, GS_HowardTailwat, GS_MarginalWay, GS_MeekerStreet, GS_MillAtEarthwo, GS_MillAtOrilla, GS_Newaukum, GS_Purification, GS_SpringbrookOr, GS_Tukwila, HH_DamEmbankment, Headworks_LowHd, PumpSta_BlackRvr, YoungsLakeOutDam",
			"output_file_dataSource":       "2dConnection_maxes.csv",
			"bucket_prefix":                "FFRD",
		},
	}
	actions.ReadStructureVariablesPeak(action)
}
func Test_Read_HDF_EVENT_Data(t *testing.T) {
	// set the manifest id environment variable.
	//os.Setenv("CC_MANIFEST_ID", "99041f15-8274-4782-b67c-bf6216e9fd95")
	action := cc.Action{
		Name:        "structure-variables-peak-output",
		Type:        "structure-variables-peak-output", //ptr[string]("copy-inputs"),
		Description: "structure-variables-peak-output",
		Parameters: map[string]any{
			"refLineDataSource":      "Duwamish_17110013.p01.hdf",
			"event_index":            1,
			"names_string_length":    36,
			"wsel_or_flow":           "flow",
			"output_file_dataSource": "event_time_series",
			"bucket_prefix":          "FFRD",
		},
	}
	actions.ReadRefLineTimeSeries(action)
}
func Test_Read_HDF_RefPoint_Min_Data(t *testing.T) {
	// set the manifest id environment variable.
	//os.Setenv("CC_MANIFEST_ID", "99041f15-8274-4782-b67c-bf6216e9fd95")
	action := cc.Action{
		Name:        "a",
		Type:        "a", //ptr[string]("copy-inputs"),
		Description: "a",
		Parameters: map[string]any{
			"refPointDataSource":     "Duwamish_17110013.p01.hdf",
			"start_event_index":      1,
			"end_event_index":        2,
			"names_string_length":    33,
			"wsel_or_velocity":       "wsel",
			"output_file_dataSource": "refpoint_min_wsel.csv",
			"bucket_prefix":          "FFRD",
		},
	}
	actions.ReadRefPointMinimum(action)
}

*/
