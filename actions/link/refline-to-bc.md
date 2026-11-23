# Refline to Boundary Condition Action

The **refline-to-boundary-condition** action reads reference line data from HDF5 RAS output files and writes it to boundary condition datasets in HDF5 RAS input files.

## Description

This action facilitates the transfer of reference line flow data from RAS output results to boundary conditions in RAS input models. Unlike the `column-to-boundary-condition` action, this action assumes that the time arrays in source and destination datasets are identical and does not perform time matching between datasets.

## Implementation Details

The action performs the following steps:

1. **Input Validation**: Validates the reference line name and configuration parameters
2. **Source Data Access**: Opens the source HDF5 file and retrieves the specified reference line dataset
3. **Destination Data Access**: Opens the destination HDF5 file and prepares the target dataset
4. **Data Processing**:
   - Reads reference line names and flow data from the source file
   - Identifies the specified reference line column
   - Copies boundary condition time values from destination
   - Combines destination time values with reference line flow data for corresponding rows
   - Writes updated boundary condition data to destination

## Configuration

### Environment

- Only S3 stores are currently supported
- S3 stores must include a "root" parameter
- Source and destination dataset paths are accessed via the "hdf" key in the Paths map

## Attributes

### Action Attributes

1. **`refline`** (string)
   - Description: The reference line name to extract data from the source file
   - Example: `"MainChannel"`
   - Required: Yes

2. **`source`** (input data source)
   - Description: Source configuration, resolved from the action's `inputs` array via `GetInputDataSource("source")`. Not an attribute map; it is a data source entry.
   - Required: Yes
   - Fields (on the input data source entry):
     * `name` (string): Must be `"source"` — this is the name used to look up the input data source
     * `store_name` (string): Name of the store containing the source HDF5 file
     * `paths` (map): The `hdf` key holds the relative path from the store root to the source HDF5 file
     * `data_paths` (map): The `refline` key holds the path to the reference line group within the source file (the group containing the `Name` and `Flow` datasets)
   - Notes: 
     - The source dataset path is taken from `data_paths["refline"]`; the `hdf` key holds the file path, not the dataset path

3. **`destination`** (output data source)
   - Description: Destination configuration, resolved from the action's `outputs` array via `GetOutputDataSource("destination")`. Not an attribute map; it is a data source entry.
   - Required: Yes
   - Fields (on the output data source entry):
     * `name` (string): Must be `"destination"` — this is the name used to look up the output data source
     * `paths` (map): The `hdf` key holds the path of the destination HDF5 file relative to the local model directory
     * `data_paths` (map): The `bcline` key holds the path to the boundary condition dataset within the destination file
   - Notes: 
     - The dest dataset path is taken from `data_paths["bcline"]`; the destination is accessed locally, not via a store

## Action Configuration Example

```json
{
  "stores": [
    {
      "name": "data",
      "store_type": "S3",
      "profile": "FFRD",
      "params": {
        "root": "/testing"
      }
    }
  ],
  "inputs": [
    {
      "name": "source",
      "store_name": "data",
      "paths": {
        "hdf": "ray-roberts.p01.hdf"
      },
      "data_paths": {
        "refline": "Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/Unsteady Time Series/Reference Lines"
      }
    }
  ],
  "outputs": [
    {
      "name": "destination",
      "store_name": "data",
      "paths": {
        "hdf": "modified/lewisville.p01.tmp.hdf"
      },
      "data_paths": {
        "bcline": "Event Conditions/Unsteady/Boundary Conditions/Flow Hydrographs/2D: Lewisville BCLine: ray-roberts_to_lewisville_1"
      }
    }
  ],
  "actions": [
    {
      "name": "refline-to-boundary-condition",
      "description": "updating bclines based on upstream ras reflines",
      "type": "link",
      "attributes": {
        "refline": "ray-roberts_to_lewisville_1|ray-roberts"
      }
    }
  ]
}
```
## Supported Store Types

This action currently supports S3 store types only. Support for other store types will be added in future releases.

## Error Handling

The action returns descriptive error messages for:
- Invalid reference line name specification
- Missing or invalid configuration parameters
- File access errors (source/destination)
- Data read/write failures
- Missing reference line data

## Usage Notes
### Source Dataset Structure
- Expected format: an hdf5 group containing two sibling datasets, `/Name` and `/Flow`
- The `data_paths["refline"]` value points to this group, not to a single dataset
- Reference line names are stored in the `<group>/Name` dataset (string values, one row per reference line)
- Flow values are stored in the `<group>/Flow` dataset (2D array of float values, one column per reference line)

### Destination Dataset Structure
- Expected format: 2D array with time as first column and boundary condition values as second column
- Each row represents a time-step boundary condition entry
- The action assumes identical time arrays between source and destination

### General Usage notes
- The action assumes that time arrays in source and destination datasets are identical
- Reference line names must exactly match those present in the source dataset
- Destination file must exist in the container's local model directory
- The action processes all rows in the datasets, matching by row index rather than timestamp