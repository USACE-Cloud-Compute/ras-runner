# Post Outputs Action Documentation

## Overview

The `post-outputs` action is a plugin action designed to copy the results of a RAS (River Analysis System) model run to a remote storage system, typically Amazon S3. This action facilitates the transfer of model output files while excluding specific reserved files such as `rasoutput` and temporary `.hdf` files.

## Configuration

### Action Registration
The action is automatically registered with the action registry using the name `post-outputs`.

### Required Attributes
The action requires the following attributes to be configured:

- `modelPrefix` (string): A prefix used to identify model output files.
- `plan` (string): The plan number associated with the model run.

### Output Data Sources
The action processes all configured output data sources in the plugin manager's outputs, with the following exceptions:
- `rasoutput`: This datasource is ignored and not copied.
- Temporary HDF files matching the pattern `{modelPrefix}.p{plan}.tmp.hdf`: These are also ignored and not copied.

### Remote Storage Configuration
The action copies files to a remote store using the following settings:
- `RemoteStoreName`: Set to the datasource name.
- `RemotePath`: Default path is set to `default`.

## Development

### File Structure
The action is implemented in `actions/utils/post-outputs.go` and follows the standard CC (Cloud Connector) plugin structure.

### Key Functions

#### `Run()`
The main execution function that calls `postOutputFiles()` to perform the file copying operation.

#### `postOutputFiles(pm *cc.PluginManager)`
Main function that iterates through all output data sources and copies eligible files to remote storage.

### Implementation Details

The action uses the following logic:
1. Retrieves required attributes (`modelPrefix` and `plan`)
2. Constructs a reserved filename pattern: `{modelPrefix}.p{plan}.tmp.hdf`
3. Iterates through all output data sources
4. Skips `rasoutput` and the reserved temporary file
5. Copies remaining files using `pm.CopyFileToRemote()`

### Error Handling
- Returns a formatted error if file copying fails
- Logs errors with descriptive messages for debugging

### Example Usage
```yaml
actions:
  - name: post-outputs
    attributes:
      modelPrefix: "my-model"
      plan: "1"
```

### Notes for Developers
- Ensure that the plugin manager has proper remote storage configuration
- The action assumes all output files are located in `actions.MODEL_DIR`
- The default remote path is hardcoded to "default"