# Create RAS TMP Action Documentation

This action creates a temporary HEC RAS input file for running RAS models with the Linux runner. It copies essential datasets from a source RAS file to a new temporary file.

## Overview

The `create-ras-tmp` action is designed to prepare a plan RAS model input file for execution by copying the necessary structure and data from an existing RAS hdf file.

## Configuration

### Action Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `src` | string | Yes | Name of the source RAS file in the model directory |
| `local_dest` | string | Yes | Name of the temporary file to create |
| `save_to_remote` | string | No | Whether to save the temporary file to a remote destination (default: "false") |
| `remote_dest` | string | No | Name of the remote data source if `save_to_remote` is true |

### JSON Configuration Example

```json
{
  "action": "create-ras-tmp",
  "attributes": {
    "src": "source_model.ras",
    "local_dest": "temp_model.ras",
    "save_to_remote": "true",
    "remote_dest": "remote_storage"
  }
}
```

## Functionality

### Main Process

1. **Input Validation**: 
   - Verifies that both `src` and `local_dest` attributes are provided
   - Parses `save_to_remote` to boolean value
   - Checks for remote destination name if remote saving is requested

2. **File Creation**:
   - Creates a new temporary file at the specified location
   - Copies core RAS datasets: "Geometry", "Plan Data", "Event Conditions"
   - Preserves essential file metadata attributes

3. **Remote Storage**:
   - If `save_to_remote` is true, copies the temporary file to the specified remote destination

### Key Datasets Copied

The action copies the following datasets from the source file:
- Geometry
- Plan Data
- Event Conditions

### Metadata Preservation

The action preserves these key file attributes:
- File Type
- File Version
- Projection
- Units System

## Error Handling

The action returns descriptive error messages for:
- Missing required attributes
- File I/O errors during creation or copying
- Invalid boolean conversion for `save_to_remote`
- Remote storage failures

## Usage Examples

### Basic Usage
```json
{
  "action": "create-ras-tmp",
  "attributes": {
    "src": "model.ras",
    "local_dest": "temp.ras"
  }
}
```

### With Remote Storage
```json
{
  "action": "create-ras-tmp",
  "attributes": {
    "src": "model.ras",
    "local_dest": "temp.ras",
    "save_to_remote": "true",
    "remote_dest": "ras_models"
  }
}
```

## Dependencies

This action requires the following dependencies:
- `github.com/usace/cc-go-sdk`
- `github.com/usace/go-hdf5`
- Standard Go libraries: `fmt`, `log`, `os`, `strconv`, `unsafe`

## Notes

- The temporary file is created in the model directory specified by `actions.MODEL_DIR`
- All copied datasets and attributes maintain their original structure and data types
- The action ensures proper file handle cleanup using defer statements
- Remote storage functionality requires proper plugin configuration for the `CopyFileToRemote` operation