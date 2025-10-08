# unsteady-simulation

## Description

The `unsteady-simulation` action executes a RAS (River Analysis System) unsteady simulation using a specified model, plan, and geometry file. This action supports preprocessing geometry files and running the RAS model script to produce simulation results.

## Implementation Details

This action is designed to automate the execution of RAS unsteady simulations by:
- Handling geometry preprocessing when enabled
- Executing RAS models
- Managing input/output data sources for simulation results and logs
- Capturing and logging all execution outputs

## Process Flow

1. **Geometry Preprocessing** (if enabled):
   - Runs the geometry preprocessor script (`geom_preproc.sh`) with the model directory, model prefix, and geometry file as arguments
   - Output from preprocessing is captured and added to the RAS output log

2. **Model Execution**:
   - Executes the RAS model script (`model.sh`) with the model directory, model prefix, plan, and geometry file as arguments
   - The combined output of the RAS model execution is captured and added to the RAS output log

3. **Results Saving**:
   - Saves the simulation results (`.p<plan>.tmp.hdf` file) to the configured output data source
   - Saves the RAS output log to a data source named `rasoutput`

## Configuration

### Environment

- **RAS_SCRIPT_PATH**: Path to the directory containing RAS scripts
  - If not set, defaults to the `actions.MODEL_SCRIPT_PATH` constant which is currently set to `/ras`

### Attributes

#### Action
- **geom_preproc**: Set to `"true"` to enable geometry preprocessing. Default is `"false"`
- **rasoutput**: The name of the output log data source. Defaults to `"rasoutput"`

#### Global (payload attributes)
- **modelPrefix**: The prefix for the RAS model files
- **plan**: The name of the RAS plan (`.cfile`)
- **geom**: The name of the geometry file (`.bfile`)


### Inputs

All of the RAS model files are assumed to have been copied to the local model directory (defaults to `/sim/model`) prior to running this action.


### Outputs

- Simulation Results: The `.p<plan>.tmp.hdf` file
- RAS Log Output: A log containing outputs from both geometry preprocessing and model execution

### Output Data Sources

- **rasoutput**: Contains the combined RAS output log
- **results**: The simulation results file in HDF format

## Configuration Examples

```json
{
  "attributes": {
    "modelPrefix": "Muncie",
    "plan": "04",
    "geom": "04"
  },
  "actions": [
    {
      "name": "unsteady-simulation",
      "type": "run",
      "description": "run a RAS 6.x unsteady simulation"
      "attributes": {
        "geom_preproc": "true"
      }
    }
  ]
}
```

## Error Handling

The action should handle:
- Missing required files or directories
- Execution failures of RAS scripts
- Invalid attribute values
- Data source access errors
- File permission issues

## Usage Notes

- Ensure that the RAS scripts (`geom_preproc.sh`, `model.sh`) are executable and located in the directory specified by `RAS_SCRIPT_PATH`
- The `modelPrefix` attribute should match the prefix used in the `.c`, `.b`, and other model files
- For more information on RAS model execution, refer to the HEC RAS documentation
- Geometry preprocessing should only be enabled when neccessary