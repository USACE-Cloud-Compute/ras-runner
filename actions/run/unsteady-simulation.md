# Unsteady Simulation Action

## Overview

The `unsteady-simulation` action executes a RAS (River Analysis System) unsteady simulation using a specified model, plan, and geometry file. This action supports preprocessing geometry files and running the RAS model script to produce simulation results.

## Configuration

### Required Attributes

- **modelPrefix**: The prefix for the RAS model files.
- **plan**: The name of the RAS plan (`.cfile`).
- **geom**: The name of the geometry file (`.bfile`).

### Optional Attributes

- **geom_preproc**: Set to `"true"` to enable geometry preprocessing. Default is `"false"`.
- **rasoutput**: The name of the output log data source. Defaults to `"rasoutput"`.

## Environment Variables

- **RAS_SCRIPT_PATH**: Path to the directory containing RAS scripts.
  - If not set, defaults to `actions.MODEL_SCRIPT_PATH`.

## Execution Flow

1. **Geometry Preprocessing** (if enabled):
   - Runs the geometry preprocessor script (`geom_preproc.sh`) with the model directory, model prefix, and geometry file as arguments.
   - Output from preprocessing is captured and added to the RAS output log.

2. **Model Execution**:
   - Executes the RAS model script (`model.sh`) with the model directory, model prefix, plan, and geometry file as arguments.
   - The combined output of the RAS model execution is captured and added to the RAS output log.

3. **Results Saving**:
   - Saves the simulation results (`.p<plan>.tmp.hdf` file) to the configured output data source.
   - Saves the RAS output log to a data source named `rasoutput`.

## Example JSON Configuration

```json


{
  "attributes": {
    "modelPrefix": "my_model",
    "plan": "my_plan",
    "geom": "my_geom.b",
    "geom_preproc": "true"
  },  
  "actions":[
    {
      "name": "unsteady-simulation",
      "type": "run",
      "description": "run a RAS 6.x unsteady simulation",
    }  
  ]
}
```

## Output

- **Simulation Results**: The `.p<plan>.tmp.hdf` file is saved to the configured output data source.
- **RAS Log Output**: A log containing outputs from both geometry preprocessing and model execution is saved to a data source named `rasoutput`.

## Notes

- Ensure that the RAS scripts (`geom_preproc.sh`, `model.sh`) are executable and located in the directory specified by `RAS_SCRIPT_PATH`.
- The `modelPrefix` attribute should match the prefix used in the `.c`, `.b`, and other model files.
- For more information on RAS model execution, refer to the HEC RAS documentation.