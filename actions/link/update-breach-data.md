# Update Breach Data Action

## Description
This action updates breach elevations in a b-file based on fragility curve results. It is designed to update the breach data in a RAS model, particularly when dealing with dam breach scenarios where failure elevations need to be updated.

## Implementation Details
The action reads a fragility curve output file and amends breach elevations in a b-file. It initializes a BFile object from the b-file, creates an SNET ID to name map from the geometry HDF file, reads and parses the fragility curve results, and then updates breach elevations in the b-file based on fragility curve results.

## Process Flow
1. Initialize BFile object from b-file
2. Create SNET ID to name map from geometry HDF file
3. Load and parse fragility curve results from fcFile
4. Update breach elevations in b-file based on fragility curve results
5. Write updated b-file back to disk

## Configuration

### Environment
- Requires RAS model directory with all input files
- Files must be accessible and properly formatted

### Attributes

#### Action

| Attribute      | Description                              |
|----------------|------------------------------------------|
| `bFile`        | Name of the b-file to update             |
| `fcFile`       | Name of the fragility curve results file |
| `geoHdfFile`   | Name of the geometry HDF file            |

### Input Files
- **bFile**: The RAS b-file containing breach data to be updated
- **fcFile**: JSON file containing fragility curve results with failure elevations
- **geoHdfFile**: HDF file containing geometry information for SNET ID mapping

### Output
The action modifies the b-file in place, updating breach elevation data.

## Configuration Example
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
