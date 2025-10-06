# cc-ras-runner

The cc-ras-runner is a cloud compute plugn for running HEC RAS models in a cloud batch processing environment.  The pluign currently support the following RAS Versions:
  - 6.3.1
  - 6.4.1
  - 6.5.0
  - 6.6.0

Both steady state and  unsteady models are supported.  In addition to running models, the plugin has a number of actions that break down intotohe following categories:

 ## Run
 Run actions execute the RAS Linux execution commands.  these include:
  - **unsteady-simulation**: this action runs the RAS linux [Unsteady Simulation](actions/run/unsteady-simulation.md).
  - **steadystate-simulation**: this action runs the RAS linux Steady State Simulation
  - **geometry-preprocessor**: this action runs the RAS Linux Geometry preprocessor 

## Link
Link actions facilitate linking data from other HEC products (HMS/RESSIM) or from upstream RAS models to a target model.  For example this might link upstream hydrographs to a downstream model boundary condition.  The following link actions are available:
  - **column-to-bc**: the [column-to-bc](actions/link/column-to-bc.md) action links column oriented data in hdf5 format to a boundary condition for a RAS model.
  - **refline-to-bc**: the [refline-to-bc](actions/link/refline-to-bc.md) action links reference line results from one RAS hdf file to the boundary condition of another RAS model. 
  - **update-bfile-skip-dss**: the [update-bfile-skip-dss](actions/link/update-bfile-skip-dss.md) action instructs RAS not to export a DSS file by setting a flag in the RAS Bfile.
  - **update-breach-data**: the [update-breach-data](actions/link/update-breach-data.md) action updates breach elevations in a RAS B-file with output from the fragility curve plugin.
  - **update-outletts-data**: the [update-outletts-data](actions/link/update-outletts-data.md) action updates a RAS bFile with new observed flow data from an HDF file

## Extract
Extract actions help to extract various ras hdf results into formats other than HDF5.
  - **ras-breach-extract**: the [ras-breach-extract](actions/extract/hdf/ras-breach-action.md) action extracts 2D Flow Area Connections data breaching conditions.
  - **ras-extract**: the [ras-extract](actions/extract/hdf/ras-extract-action.md) action is a tool to extract user defined datasets and attributes from RAS hdf5 output.

## Utils
Utility actions
  - **copy-inputs**: the [copy-inputs](actions/utils/copy-inputs-action.md) action assists with bulk copying of model input files into the compute plugin.
  - **create-ras-tmp**: the [create-ras-tmp](actions/utils/create-ras-tmp.md) action creates a ras tmp file from an input plan hdf file.  The linux ras runner requires this file to run. 
  - **post-outputs**: the [post-outputs](actions/utils/post-outputs.md) action copies output files to an external store.
---



## Key Features

- **Simulation Execution**: Executes RAS unsteady simulations with support for geometry preprocessing and model running.
- **Flexible Input/Output Management**: Reads/Saves simulation files and results to configured data sources.
- **Integration Ready**: 

For detailed configuration and usage instructions, refer to the [full documentation](unsteady-simulation.md).


Considerations
 - there are minor solver differences between the 2D flow solver compiled on Windows vs compiled on Linux....