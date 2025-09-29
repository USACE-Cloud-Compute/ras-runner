# Update Outlet TS Action

## Overview
The `update-outlet-ts-bfile` action updates a RAS bFile with new observed flow data from an HDF file. This action modifies the outlet time series data within a RAS bFile using values extracted from an HDF dataset.

## Action Type
`update-outlet-ts-bfile`

## Attributes

| Attribute | Required | Description |
|-----------|----------|-----------|
| `bFile` | Yes | Name of the bFile to update (relative path) |
| `outletTS` | Yes | Name of the outlet time series to update |
| `hdfFile` | Yes | Name of the HDF file containing source data |
| `hdfDataPath` | Yes | Path to the dataset within the HDF file |

## Input Files

### bFile
- A RAS bFile that contains the outlet TS data to be updated
- Must exist in the model directory before running the action

### HDF File
- An HDF5 file containing the new flow data
- The HDF file must be accessible at runtime and contain the specified `hdfDataPath`

## Output

The action modifies the input bFile in-place, updating the outlet TS data with values from the HDF dataset.

## Example Usage (JSON)

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

## Testing

The action includes unit tests that verify:
- Correct file path resolution
- Proper HDF5 dataset reading
- Successful bFile modification
- Error handling for missing files or datasets

## Error Handling

The action returns specific error messages for:
- Missing required attributes
- File not found errors
- Invalid HDF5 dataset paths
- Failure to read or write data

## Notes

- The action assumes the input files are already copied to the local model directory
- The outlet TS name must exactly match the one in the bFile
- The HDF dataset must contain a column of flow values to update the outlet TS with
```