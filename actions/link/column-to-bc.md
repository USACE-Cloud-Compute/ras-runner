# Column to Boundary Condition Action

The **column-to-boundary-condition** action reads columnar data from HDF5 RAS output files and writes it to boundary condition datasets in HDF5 RAS input files.

## Description

This action facilitates the transfer of flow or other hydrological data from RAS output results to boundary conditions in RAS input models. It maps time-series data from a source dataset to corresponding boundary condition entries in a destination dataset based on matching timestamps.

## Implementation Details/Process Flow

The action performs the following steps:

1. **Input Validation**: Validates the column index and configuration parameters
2. **Source Data Access**: Opens the source HDF5 file and retrieves the specified dataset
3. **Destination Data Access**: Opens the destination HDF5 file and prepares the target dataset
4. **Data Processing**:
   - Reads time-series data from the source file
   - Matches timestamps between source and destination datasets
   - Extracts specified column data from source for each time step
   - Writes updated boundary condition data to destination

5. Time tolerance for matching is defined by the `actions.Tolerance` constant if the user does not supply a value
6. Column indexing is 1-based (first data column = column 1)

## Configuration

## Store
A single store is configured for this action.

## Inputs
An input data source must be configured with a single file path key of "hdf" representing the relative path from the store to the hdf5 input file.

## Action Attributes

1. **`column_index`** (string)
   - Description: The column index of data to extract from the source file (1-based indexing)
   - Example: `"3"`
   - Required: Yes

2. **`time-tolerance`** (float)
   - Description: The tolerance used for matching time values between the two models in decimal days. 
   - Example: `0.00005`
   - Required: No.  Default value is `0.000001`

3. **`src`** (map)
   - Description: Source configuration parameters
   - Required: Yes
   - Fields:
     * `name` (string): Name of the input data source
     * `datapath` (string): Path to the dataset within the source file
   - Notes: 
     - Only S3 stores are currently supported and they must include a "root" parameter
     - The source dataset path is accessed via the "hdf" key in the Paths map

4. **`dest`** (map)
   - Description: Destination configuration parameters for the local destination hdf5
   - Required: Yes
   - Fields:
     * `name` (string): Name of the output data source
     * `datapath` (string): Path to the dataset within the destination file
   - Notes: 
     - Only S3 stores are currently supported and they must include a "root" parameter
     - The dest dataset path is accessed locally

## Configuration Example

```json
stores:[
  {
      "name": "src_data",
      "store_type": "S3",
      "profile": "FFRD",
      "params": {
          "root": "/testing"
      }
  },
],

"inputs": [
  {
    "name": "input-model",
    "store_name": "src_data",
    "paths": {
      "hdf":"input-model.p01.hdf"
    }
  }
],    

{
  "action": "column-to-boundary-condition",
  "attributes": {
    "column_index": "3",
    "time_tolerance": 0.00001,
    "src": {
      "name": "input-model",
      "datapath": "/results/flow"
    },
    "dest": {
      "name": "output-model.p01.tmp.hdf",
      "datapath": "/boundary/flow"
    }
  }
}
```



## Data Format Requirements

### Source Dataset Structure
- Expected format: 2D hdf5 dataset

### Destination Dataset Structure
- Expected format: 2D array with time as first column and boundary condition values as second column
- Each row represents a time-step boundary condition entry

## Time Dataset Path

The source time dataset path is derived from the source data dataset path using the `actions.TimePath` function. The time dataset is separate from the data dataset and contains timestamp information.

## Supported Store Types

This action currently supports S3 store types only. Support for other store types will be added in future releases.

## Error Handling

The action returns descriptive error messages for:
- Invalid column index specification
- Missing or invalid configuration parameters
- File access errors (source/destination)
- Data read/write failures
- Timestamp matching failures

## Usage Notes

- Column indexing is 0-based (first data column = column 0)
- Time tolerance for matching is defined by the `actions.Tolerance` constant
- Destination file must exist in the container's local model directory

## Patterns and Best Practices

- Ensure destination files exist before running the action
- Validate that source datasets contain the expected time-series structure
- Use consistent timestamp formats between source and destination
- Test with small datasets before processing large files
- Maintain proper S3 store configuration with required "root" parameter