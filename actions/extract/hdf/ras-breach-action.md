# RAS-Runner: RAS Breach Action

The RAS breach action generates a report that inspects 2D Hyd Conn datasets for breaching conditions. It extracts data locally and writes it to more accessible formats such as JSON and Console STDOUT. Support for cloud-native array stores like TileDB or Zarr is anticipated in the near future.

## Configuration

The action is configured via JSON configuration:

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

### Configuration Parameters

- **`name`**: Required field that instructs the RAS Runner to run a ras-breach-extract. This value should only be set to "ras-breach-extract"
- **`type`**: Fixed value that should be set to "extract"
- **`description`**: User-defined text describing what is being extracted
- **`accumulate-results`**: When added to a block, informs the plugin not to write out the block but to accumulate the extraction. To configure the extract action to write to output, use the "outputDataSource" attribute. An extraction configuration should include either "accumulate-results" or "outputDataSource" but not both.
- **`outputDataSource`**: Configures the action to write the current and any accumulated extractions to output. The outputDataSource is a reference to a data source name that describes the store and file name.

### Output Data Source Example

```json
{
    "name": "metadataout",
    "paths": {
        "extract": "ras-event-data.json"
    },
    "store_name": "FFRD"
}
```

## Output Format

The breach action outputs JSON data using the [JSON Output Format Specification](./ras-extractor-json-output-file-specification.md), producing one `record` per 2D Hyd Conn location.

### Breach Record Fields

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