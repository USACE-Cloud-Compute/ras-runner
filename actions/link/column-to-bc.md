# Column to Boundary Condition Action

The **column-to-boundary-condition** action reads columnar data from HDF5 RAS output files and writes it to boundary condition datasets in HDF5 RAS input files.

## Description

This action facilitates the transfer of flow or other hydrological data from RAS output results to boundary conditions in RAS input models. It maps time-series data from a source dataset to corresponding boundary condition entries in a destination dataset based on matching timestamps.

## Action Attributes

### Required Attributes

1. **`column_index`** (string)
   - Description: The column index of data to extract from the source file (1-based indexing)
   - Example: `"3"`
   - Required: Yes

2. **`src`** (map)
   - Description: Source configuration parameters
   - Required: Yes
   - Fields:
     * `name` (string): Name of the input data source
     * `datapath` (string): Path to the dataset within the source file
   - Notes: 
     - Only S3 stores are currently supported and they must include a "root" parameter
     - The source dataset path is accessed via the "0" key in the Paths map

3. **`dest`** (map)
   - Description: Destination configuration parameters
   - Required: Yes
   - Fields:
     * `name` (string): Name of the output data source
     * `datapath` (string): Path to the dataset within the destination file
   - Notes: 
     - Only S3 stores are currently supported and they must include a "root" parameter
     - The dest dataset path is accessed via the "0" key in the Paths map

## Action Configuration Example

```json
{
  "action": "column-to-boundary-condition",
  "attributes": {
    "column_index": "3",
    "src": {
      "name": "source_data",
      "datapath": "/results/flow"
    },
    "dest": {
      "name": "boundary_condition",
      "datapath": "/boundary/flow"
    }
  }
}
```

## Implementation Details

The action performs the following steps:

1. **Input Validation**: Validates the column index and configuration parameters
2. **Source Data Access**: Opens the source HDF5 file and retrieves the specified dataset
3. **Destination Data Access**: Opens the destination HDF5 file and prepares the target dataset
4. **Data Processing**:
   - Reads time-series data from the source file
   - Matches timestamps between source and destination datasets
   - Extracts specified column data from source for each time step
   - Writes updated boundary condition data to destination

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