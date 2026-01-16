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

2. **`source`** (map)
   - Description: Source configuration parameters
   - Required: Yes
   - Fields:
     * `name` (string): Name of the input data source
     * `datapath` (string): Path to the dataset within the source file
   - Notes: 
     - Only S3 stores are currently supported and they must include a "root" parameter
     - The source dataset path is accessed via the "hdf" key in the Paths map

3. **`destination`** (map)
   - Description: Destination configuration parameters
   - Required: Yes
   - Fields:
     * `name` (string): Name of the output data source
     * `datapath` (string): Path to the dataset within the destination file
   - Notes: 
     - Only S3 stores are currently supported and they must include a "root" parameter
     - The dest dataset path is accessed via the "hdf" key in the Paths map

## Action Configuration Example

```json
{
  "action": "refline-to-boundary-condition",
  "attributes": {
    "refline": "MainChannel",
    "source": {
      "name": "source_data",
      "datapath": "/results/refline"
    },
    "destination": {
      "name": "boundary_condition",
      "datapath": "/boundary/flow"
    }
  }
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
- Expected format: 2D hdf5 dataset containing reference line names and flow values
- The source dataset path should contain both `/Name` and `/Flow` datasets
- Reference line names are stored in the `/Name` dataset
- Flow values are stored in the `/Flow` dataset

### Destination Dataset Structure
- Expected format: 2D array with time as first column and boundary condition values as second column
- Each row represents a time-step boundary condition entry
- The action assumes identical time arrays between source and destination

### General Usage notes
- The action assumes that time arrays in source and destination datasets are identical
- Reference line names must exactly match those present in the source dataset
- Destination file must exist in the container's local model directory
- The action processes all rows in the datasets, matching by row index rather than timestamp