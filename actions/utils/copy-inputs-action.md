# Copy Inputs Action

The **copy-inputs-action** action copies all data source inputs and their respecive paths to a local plugin data folder.

## Description

This action facilitates the transfer of files from a remote store to the local directory being used for model processing.  Most commonly it is used to transfer the RAS Model files for local execution.

## Action Attributes

### Required Attributes

1. **`inputs`** (array)
   - Description: List of input data source configurations
   - Required: Yes
   - Fields:
     * `name` (string): Name of the input data source
     * `paths` (map): Mapping of source path keys to local file paths
   - Notes: 
     - Each input source can specify multiple file paths to copy
     - Source paths are accessed via the "hdf" key in the Paths map

## Action Configuration Example

```json
{
  "action": "copy-inputs",
  "attributes": {
    "inputs": [
      {
        "name": "input_data",
        "paths": {
          "hdf": "input/model.hdf"
        }
      }
    ]
  }
}
```

## Implementation Details

The action performs the following steps:

1. **Input Validation**: Validates the inputs configuration and required attributes
2. **Source Data Access**: Opens the source data files and retrieves the specified datasets
3. **Destination Data Access**: Creates destination files in the local model directory
4. **Data Processing**:
   - Reads data from source files
   - Writes data to destination files in the model directory
   - Maintains original filenames during copying

## Data Format Requirements

### Source Dataset Structure
- Expected format: Any file type supported by the data source

### Destination Dataset Structure
- Expected format: Local file system directory (actions.MODEL_DIR)

## Supported Store Types

This action supports all store types that are configured in the plugin manager.

## Error Handling

The action returns descriptive error messages for:
- Invalid inputs specification
- Missing or invalid configuration parameters
- File access errors (source/destination)
- Data read/write failures
- Path resolution errors