# Update Breach Data Action

This action updates breach elevations in a b-file based on fragility curve results.

## Overview

The `update-breach-bfile` action reads a fragility curve output file and amends breach elevations in a b-file. It's designed to update the breach data in a RAS model, particularly when dealing with dam breach scenarios where failure elevations need to be updated.

## Action Configuration

### Required Parameters

| Parameter      | Description                              |
|----------------|------------------------------------------|
| `bFile`        | Name of the b-file to update             |
| `fcFile`       | Name of the fragility curve results file |
| `geoHdfFile`   | Name of the geometry HDF file            |

### Example Configuration

```json
{
  "action": {
    "type": "update-breach-bfile",
    "description": "Update breach elevations based on fragility curves",
    "attributes": {
      "bFile": "Duwamish_17110013.b01",
      "fcFile": "failure_elevations.json",
      "geoHdfFile": "Duwamish_17110013.g01.hdf"
    }
  }
}
```

## How It Works

1. **Initialization**: The action reads the b-file and initializes a BFile object
2. **Geometry Mapping**: Creates an SNET ID to name map from the geometry HDF file
3. **Fragility Curve Loading**: Reads and parses the fragility curve results
4. **Breach Elevation Update**: Amends breach elevations in the b-file based on fragility curve results
5. **Output**: Writes the updated b-file back to disk

## File Requirements

### Input Files
- **bFile**: The RAS b-file containing breach data to be updated
- **fcFile**: JSON file containing fragility curve results with failure elevations
- **geoHdfFile**: HDF file containing geometry information for SNET ID mapping

### Output
The action modifies the b-file in place, updating breach elevation data.

## Error Handling

If any step fails during execution:
- The action returns an error and stops processing
- Specific errors include file not found, invalid JSON format, or failure to update breach elevations

## Usage Notes

This action should be run after copying local files using the `copy-local` action. It requires all input files to exist in the model directory.

## Example fragility curve results format (fcFile)
```json
{
  "Results": [
    {
      "Name": "Dam1",
      "FailureElevation": 250.5
    }
  ]
}
```

## Test Requirements

The action can be tested with unit tests that:
- Verify file paths are correctly constructed
- Confirm fragility curve results are properly loaded and parsed
- Validate breach elevation updates occur successfully
- Check error handling for missing or invalid files
