# Update Outlet TS Action

## Description
The `update-outlet-ts-bfile` action updates a RAS bFile with new observed flow data from an HDF file. This action modifies the outlet time series data within a RAS bFile using values extracted from an HDF dataset.

## Implementation Details
This action is implemented as a link-type action that processes RAS bFiles and HDF5 datasets. It reads flow data from a specified HDF dataset and updates corresponding outlet time series data in the bFile.

## Process Flow
1. Validate required attributes are present
2. Resolve file paths for bFile and HDF file
3. Read flow data from specified HDF dataset path
4. Locate and update the specified outlet time series in the bFile
5. Save modified bFile in-place

## Configuration

### Environment
- Action requires access to RAS bFile and HDF5 files
- Files must be accessible in the model directory at runtime

### Attributes
#### Action
| Attribute | Required | Description |
|-----------|----------|-------------|
| `bFile` | Yes | Name of the bFile to update (relative path) |
| `outletTS` | Yes | Name of the outlet time series to update |
| `hdfFile` | Yes | Name of the HDF file containing source data |
| `hdfDataPath` | Yes | Path to the dataset within the HDF file |


## Configuration Examples

```json
{
  "name": "update-outlet-ts-bfile",
  "type": "link",
  "description": "Update outlet flow data with observed values",
  "attributes": {
    "bFile": "MyModel.b01",
    "outletTS": "River Outlet: Flow Hydrograph",
    "hdfFile": "observed_flows.p01.hdf",
    "hdfDataPath": "/FlowData/TimeSeries"
  }
}
```

## Error Handling
The action returns specific error messages for:
- Missing required attributes
- File not found errors
- Invalid HDF5 dataset paths
- Failure to read or write data

## Usage Notes
- The action assumes input files are already copied to the local model directory
- The outlet TS name must exactly match the one in the bFile
- The HDF dataset must contain a column of flow values to update the outlet TS with
- Action modifies the input bFile in-place