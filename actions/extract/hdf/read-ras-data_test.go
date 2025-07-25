package hdf

import (
	"fmt"
	"testing"
)

/*

	Group

	Dataset
	  - Array
	  - Compound

	Summary


	- is this a group op
	    - yes: open group
	    - no skip to dataset loading: extract name is based on the supplied name
	  - get objects in group
	  - for each object
	    - if object is a dataset
		  - load dataset: extract name is supplied name concatonated with the group name

    - dataset loading
	  - check dataset type (array or compound)
	    - array:
		  - read array into memory
		    - should summary function be invoked?
		      - yes
			    - create a new outout dataset
			    - run summary function (Max, Min, Mean)
		  - return output array
		- compound data type:
		 - read into struct?


///////////////

actions:[
  {
    "name": "extract model data to cloud"
    "type": "extract",
    "attributes": {
      "outputformat": "csv",
	  "exportconfig:[
	  	"boundary-conditions":{
			"match":"^*_flow",
			"postprocess:["max","min"],
			"omitdata":true,
			"colnames":["flow(cfs)","stage(ft)"]. //either use this, or coldata
			"coldata":"Name" //if path is group then looks for this dataset, if path is dataset, looks for this dataset is the same path prefix
		}
	  ],
    },
    "inputs":[
      {
        "name": "rasOutput",
        "paths": {
          "extractFile": "{ATTR::modelPrefix}.p{ATTR::plan}.hdf"
        },
        "datapaths": {
          "boundary-conditions": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",  << group
          "reference-lines": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines", << group
          "max-water-surface": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Max Water Surface". <<array dataset
        },
        "store_name": ""
      },
    ],
    "outputs":[
      {
        "name": "extractOutputTemplate",
        "paths": {
          "extract": "results/{ENV::EVENT_IDENTIFIER}/asdfasdf/{ATTR::modelPrefix}_{datapaths-key}_{post-process-name}.{ATTR::outputFormat}"
        },
        "store_name": "FFRD"
      },
    ]
  }
]

////////////////////////
Junk below

actions:[
  {
    "name": "extract model data to cloud"
    "type": "extract",
    "attributes": {
      "outputFormat": "csv",
      "post-processes": {
        "boundary-conditions":["max","min"]
      },
	  "exportConfig:[
	  	"boundary-conditions":{
			"match":"asdfadsf",
			"postProcess:["max","min"],
			"omitData":true
		}
	  ],
      "omit-data": [
		"boundary-conditions"
      ],
      "match": {
		"boundary-conditions":[
	  		"regex-match":"asdfasdf"
        ]
      }
    },
    "inputs":[
      {
        "name": "rasOutput",
        "paths": {
          "extractFile": "{ATTR::modelPrefix}.p{ATTR::plan}.hdf"
        },
        "datapaths": {
          "boundary-conditions": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",  << group
          "reference-lines": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines", << group
          "max-water-surface": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Max Water Surface". <<array dataset
        },
        "store_name": ""
      },
    ],
    "outputs":[
      {
        "name": "extractOutputTemplate",
        "paths": {
          "extract": "results/{ENV::EVENT_IDENTIFIER}/asdfasdf/{ATTR::modelPrefix}_{datapaths-key}_{post-process-name}.{ATTR::outputFormat}"
        },
        "store_name": "FFRD"
      },
    ]
  }
]

*/

func TestReadRasData1(t *testing.T) {
	a := "float32"
	config := RasExtractConfig{
		Colnames: []string{"flow", "stage"},
	}

	input := ReadRasDataInput{
		DataPath: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/HHRes_Outlet_DS",
		Config:   config,
	}

	hdffile := "/mnt/testdata/1/testdata.hdf"

	switch a {
	case "float32":
		reader, err := NewRasExtractReader[float32](hdffile)
		if err != nil {
			t.Fatal(err)
		}
		out, err := reader.Read(input)
		if err != nil {
			t.Fatal(err)
		}
		writer := ConsoleRasExtractWriter[float32]{}
		writer.Write(out)
		fmt.Println("done")
	}
}

// func TestReadRasData_ArrayOnly(t *testing.T) {
// 	config := RasExportConfig{
// 		Colnames: []string{"flow", "stage"},
// 	}

// 	input := ReadRasDataInput{
// 		DataPath: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/HHRes_Outlet_DS",
// 		Config:   config,
// 	}

// 	hdffile := "/mnt/testdata/1/testdata.hdf"

// 	reader, err := NewRasReader[float32](hdffile)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	array, err := reader.ReadArray(input)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	testLen := 1441
// 	if len(array) != testLen {
// 		t.Errorf("Expected %d got %d", testLen, len(array))
// 	}
// }

// func TestReadRasData_SingleSummary(t *testing.T) {
// 	config := RasExportConfig{
// 		Colnames: []string{"flow", "stage"},
// 	}

// 	input := ReadRasDataInput{
// 		DataPath: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/HHRes_Outlet_DS",
// 		Config:   config,
// 	}

// 	hdffile := "/mnt/testdata/1/testdata.hdf"

// 	reader, err := NewRasReader[float32](hdffile)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	data, err := reader.ReadArray(input)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(data) > 0 {
// 		cols := len(data[0])
// 		colSummary := make([]float32, cols)

// 		for k := 0; k < cols; k++ {
// 			colSummary[k] = colMax(data, k)
// 		}

// 		fmt.Println(colSummary)

// 	}

// }

// func TestExtractBoundaryConditionArray(t *testing.T) {
// 	hdffile := "/mnt/testdata/1/testdata.hdf"
// 	datapath := "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/HHRes_Outlet_DS"

// 	reader, err := NewRasReader[float32](hdffile)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	array, err := reader.ReadArray(RasReadInput{
// 		Datapath: datapath,
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(array)

// }

// func TestReadGroupsArray(t *testing.T) {
// 	hdffile := "/mnt/testdata/1/testdata.hdf"

// 	reader, err := NewRasReader[float32](hdffile)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	gm, err := reader.GroupMembers("/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	for i, grp := range gm {
// 		fmt.Println(i, grp)
// 		input := RasReadInput{
// 			Datapath: fmt.Sprintf("/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/%s", grp),
// 		}
// 		data, err := reader.ReadArray(input)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		//process if we have data//
// 		if len(data) > 0 {
// 			cols := len(data[0])
// 			colSummary := make([]float32, cols)

// 			//for j := 0; j < len(data); j++ {
// 			for k := 0; k < cols; k++ {
// 				colSummary[k] = colMax(data, k)
// 			}
// 			//}

// 			fmt.Println(colSummary)

// 		}

// 		//////////
// 		fmt.Println(fmt.Sprintf("%s::%v", grp, data))
// 	}

// 	fmt.Println(gm)

// 	f, err := hdf5utils.OpenFile(hdffile)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	namesDataSet, err := hdf5utils.NewHdfDataset(REFLINE_RESULT_PATH+"Name", hdf5utils.HdfReadOptions{
// 		Dtype:        reflect.String,
// 		Strsizes:     hdf5utils.NewHdfStrSet(36),
// 		File:         f,
// 		ReadOnCreate: true,
// 	})

// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(namesDataSet.Rows())
// 	for i := 0; i < namesDataSet.Rows(); i++ {
// 		dest := []string{}
// 		err = namesDataSet.ReadRow(i, &dest)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		fmt.Println(dest)
// 	}

// 	fmt.Println(namesDataSet)
// }
