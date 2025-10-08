# update-bfile-skip-dss

## Description

The `update-bfile-skip-dss` action modifies a specified RAS bFile by appending a predefined SKIPDSS command if it's not already present. This is used to instruct the HECRAS linux runtime to skip exporting DSS following a model run.

## Implementation Details

This action modifies files in place and assumes that the necessary files are already present in a local directory accessible to the runner. The appended text is a specific format expected by HEC RAS linux runtime version 6.x.

### Process Flow

1. **Log Initialization**: Logs that the action is ready to run.
2. **Directory Setup**: If `ModelDir` is not set, it defaults to `actions.MODEL_DIR`.
3. **File Path Construction**: Builds the full path to the bFile using `ModelDir` and the `bFile` attribute.
4. **File Validation**:
   - Checks if the file exists at the constructed path.
   - If not found, returns an error indicating that the copy-local action should be run first.
5. **File Reading**: Reads the entire content of the bFile as a string.
6. **Skip Command Check**:
   - If `SKIPDSS` is not found within the file content, it appends `SKIPDSS` to the end of the file.
7. **File Writing**:
   - Writes the modified content back to the same file with permissions set to 0600 (read/write for owner only).

## Configuration

### Environment

- `MODEL_DIR`: Default directory path in the running container where the bFile is located (used when `ModelDir` is not specified)

### Attributes

- `bFile` (required): The name of the bFile to be updated
- `ModelDir` (optional): Directory path where the bFile is located

### Action

- Action type: `update-bfile-skip-dss`

## Configuration Example

```json
{
 "actions": [
  "update-bfile-skip-dss": {
    "type": "update-bfile-skip-dss",
    "attributes": {
      "bFile": "my_model.b"
    }
  }
 ]
}
```

### Error Handling

- Returns errors if:
  - The specified bFile is not found in the local directory.
  - File reading fails.
  - File writing fails.


## Usage Notes
This action should be run after ensuring the bFile has been copied locally using the `copy-local` action to ensure the file exists in the expected local directory.

- This action modifies files in place.
- It assumes that the necessary files are already present in a local directory accessible to the runner.
- The appended text is a specific format expected by HEC RAS linux runtime version 6.x