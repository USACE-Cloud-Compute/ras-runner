# Post Outputs Action Documentation

## Description

The `post-outputs` action is a plugin action designed to copy the results of a RAS (River Analysis System) model run to a remote storage system, typically Amazon S3. This action facilitates the transfer of model output files while excluding specific reserved files, specifically `rasoutput` and temporary `.hdf` files.

## Process Flow

1. The action retrieves required attributes (`modelPrefix` and `plan`)
2. Constructs a reserved filename pattern: `{modelPrefix}.p{plan}.tmp.hdf`
3. Iterates through all output data sources in the plugin manager
4. Skips `rasoutput` and the reserved temporary file
5. Copies remaining files using `pm.CopyFileToRemote()`
6. Returns formatted errors if file copying fails

## Configuration

### Attributes

#### Global
- `modelPrefix` (string): A prefix used to identify model output files
- `plan` (string): The plan number associated with the model run


### Output Data Sources
The action processes all configured output data sources in the plugin manager's outputs, with the following exceptions:
- `rasoutput`: This datasource is ignored and not copied.
- Temporary HDF files matching the pattern `{modelPrefix}.p{plan}.tmp.hdf`: These are also ignored and not copied.
- the output copies to the dataset path of `default`

### Configuration Examples
####  - @TODO

### Error Handling
- Returns a formatted error if file copying fails
- Logs errors with descriptive messages for debugging