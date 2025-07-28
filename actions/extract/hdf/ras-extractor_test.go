package hdf

import (
	"reflect"
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

const (
	TestRasHdfFile string = "/mnt/testdata/1/testdata.hdf"
)

func TestReadBcLinePeak(t *testing.T) {
	input := RasExtractInput{
		GroupPath:      "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",
		Colnames:       []string{"flow", "stage"},
		Postprocess:    []string{"max"},
		ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell",
		DataType:       reflect.Float32,
		WriteSummary:   true,
		WriterType:     ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineLinePeakWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineLinePeakFlow(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineTimeSeriesWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		DataType:        reflect.Float32,
		WriteData:       true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadReflineTimeSeriesFlow(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		DataType:        reflect.Float32,
		WriteData:       true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMinVelocity(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMinWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointPeakVelocity(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointPeakWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRefpointMaxAndMinWaterSurface(t *testing.T) {
	input := RasExtractInput{
		DataPath:        "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Water Surface",
		ColNamesDataset: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
		Postprocess:     []string{"max", "min"},
		DataType:        reflect.Float32,
		WriteSummary:    true,
		WriterType:      ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadStructureVariablePeak(t *testing.T) {
	input := RasExtractInput{
		GroupPath:    "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/Perimeter 1/2D Hyd Conn",
		GroupSuffix:  "Structure Variables",
		Colnames:     []string{"Total Flow", "Weir Flow", "Stage HW", "Stage TW", "Total Culv"},
		Postprocess:  []string{"max", "min"},
		DataType:     reflect.Float32,
		WriteSummary: true,
		WriterType:   ConsoleWriter,
	}

	err := RunExtract(input, TestRasHdfFile)
	if err != nil {
		t.Fatal(err)
	}
}
