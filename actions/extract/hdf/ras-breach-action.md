# RAS Breach Action

## Description
The RAS breach action generates a report that inspects 2D Hyd Conn datasets for breaching conditions. It extracts data locally from hdf5 and writes it to more accessible formats such as JSON.

## Implementation Details
The action processes 2D Hyd Conn datasets to identify breach conditions by analyzing breaching variables and time series data. It extracts specific data points related to breach progression, flow characteristics, and hydraulic conditions.

## Process Flow
1. Read 2D Hyd Conn datasets from HDF5 files
2. Analyze breaching variables and time series data
3. Identify breach conditions and progression duration
4. Extract maximum flow, stage HW, and stage TW values
5. Generate breach records in JSON format
6. Output results to configured data sources

## Configuration

### Environment
- there are no environment variables unique to `ras-breach-action`

### Attributes
#### Action
- **`name`**: Required field that instructs the RAS Runner to run a ras-breach-extract. This value should only be set to "ras-breach-extract"
- **`type`**: Fixed value that should be set to "extract"
- **`description`**: User-defined text describing what is being extracted
- **`accumulate-results`**: When added to a block, informs the plugin not to write out the block but to accumulate the extraction. To configure the extract action to write to output, use the "outputDataSource" attribute. An extraction configuration should include either "accumulate-results" or "outputDataSource" but not both.
- **`outputDataSource`**: Configures the action to write the current and any accumulated extractions to output. The outputDataSource is a reference to a data source name that describes the store and file name.


  ---
  ```json
  //sample json action configuration for the breach extract action
  "actions":[
    {
      "name": "ras-breach-extract",
      "type": "extract",
      "description": "breaching summary information for my model",
      "attributes": {
        "outputDataSource": "metadataout"
      }
    }
  ]
  ```
  ---

#### Global
this action requires two attributes to be populated globally (i.e. payload attributes)
- **modelPrefix**: This is the name of the model on the filesystem without an extension ( e.g. `Muncie` for `Muncie.prj`).
- **plan**: this is a two character string representing the plan that will be used (e.g. `04` for `p04`)  

### Inputs
- HDF5 files containing 2D Hyd Conn datasets.  The hdf5 is assumed to exist in the local model directory (`/sim/model`) prior to running this action.  Typically this action is run immediaty after running the ras model so the hdf5 plan output file is already in the model directory.  Alternatively the hdf5 can be copied from a remote resource to this local direction using a copy-inputs action.

### Input Data Sources
- te only input necessary for 

### Outputs
- JSON formatted breach records
- Console STDOUT output

### Output Data Sources
```json
{
    "name": "metadataout",
    "paths": {
        "extract": "ras-event-data.json"
    },
    "store_name": "FFRD"
}
```

## Configuration Examples
```json
{
    "name": "ras-breach-extract",
    "type": "extract",
    "description": "breach-extract",
    "attributes": {
        "outputDataSource": "metadataout"
    }
}
```

## Outputs

### Format
JSON record data using the [JSON Output Format Specification](./ras-extractor-json-output-file-specification.md)

```json
{
    "BreachIndex": -1,
    "BreachProgressionDuration": 0,
    "BreachStartTime": null,
    "Breached": false,
    "Event": "1",
    "FlowArea2D": "bardwell-creek",
    "HWAtBreach": null,
    "MaxBottomWidth": null,
    "MaxFlow": null,
    "MaxHW": 532.1331,
    "MaxTW": 483.99698,
    "SaConn": "nid_tx01255",
    "TWAtBreach": null
}
```

### Field Descriptions
- **`BreachIndex`**: The ordinal position (i.e. index) of the breach date in the time step data (`/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Time`)
- **`BreachProgressionDuration`**: The duration in timestep units that the *"Breaching Velocity"* from the *"Breaching Variables"* dataset exceeds a threshold velocity. The threshold velocity is currently a float32 constant value of "1.5".
- **`BreachStartTime`**: The hdf5 attribute value *"Breach at Time (Days)"* associated with the *"Breaching Variables"* dataset
- **`Breached`**: False if *BreachStartTime* is NaN. True if it is a value.
- **`Event`**: The event identifier for the compute run
- **`FlowArea2D`**: The 2D Flow Area name from the *"/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/"* group
- **`HWAtBreach`**: The *"Breaching Variables"* data value at the *"Breach Index"* for the Stage HW Column (constant value: 0)
- **`MaxBottomWidth`**: The maximum value from the *"Breaching Variables"* data Bottom Width Column (constant value: 2)
- **`MaxFlow`**: The maximum value from the *"Breaching Variables"* data Breach Flow Column (constant value: 6)
- **`MaxHW`**: The maximum value from the *"Breaching Variables"* data Stage HW Column (constant value: 0)
- **`MaxTW`**: The maximum value from the *"Breaching Variables"* data Stage TW Column (constant value: 1)
- **`SaConn`**: The name of the *"2D Hyd Conn"* group (`/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/{flow area name}/2D Hyd Conn/`)
- **`TWAtBreach`**: The *"Breaching Variables"* data value at the *"Breach Index"* for the Stage TW Column (constant value: 1)

**Note**: Column constant values are array ordinal positions. i.e. 0 is the 1st column.

## Error Handling
- Invalid HDF5 file paths will result in error messages and a compute run failure
- Missing required attributes defined above will result in a compute run failure

## Usage Notes
- Currently only Breaching Variables output for 2D flow areas with 2D Hyd Conn datasets can be extracted
- the attribute `"accumulate-results":true` from the `ras-extract-action` can be used with this action to defer writing results.  refer to the ras-extract-action documentation for instructions on using accumulate-results.

## Future Enhancements
- Additional output format options and support for cloud-native array stores like TileDB or Zarr

## Patterns and Best practices
- Consider using accumulate-results for batch processing of multiple datasets