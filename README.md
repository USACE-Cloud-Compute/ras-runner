# cc-ras-runner

The cc-ras-runner is a cloud compute plugin for running HEC RAS models in a cloud batch processing environment. The plugin currently supports the following RAS Versions:
  - 6.3.1
  - 6.4.1
  - 6.5.0
  - 6.6.0

Both steady state and unsteady models are supported. In addition to running models, the plugin has a number of actions that break down into the following categories:

## Run
Run actions execute the RAS Linux commands. These include:
  - **unsteady-simulation**: This action runs the RAS Linux [Unsteady Simulation](actions/run/unsteady-simulation.md).
  - **steadystate-simulation**: This action runs the RAS Linux Steady State Simulation
  - **geometry-preprocessor**: This action runs the RAS Linux Geometry preprocessor 

## Link
Link actions facilitate linking data from other HEC products (HMS/RESSIM) or from upstream RAS models to a target model. For example, this might link upstream hydrographs to a downstream model boundary condition. The following link actions are available:
  - **column-to-bc**: The [column-to-bc](actions/link/column-to-bc.md) action links column-oriented data in HDF5 format to a boundary condition for a RAS model.
  - **refline-to-bc**: The [refline-to-bc](actions/link/refline-to-bc.md) action links reference line results from one RAS HDF file to the boundary condition of another RAS model. 
  - **update-bfile-skip-dss**: The [update-bfile-skip-dss](actions/link/update-bfile-skip-dss.md) action instructs RAS not to export a DSS file by setting a flag in the RAS B-file.
  - **update-breach-data**: The [update-breach-data](actions/link/update-breach-data.md) action updates breach elevations in a RAS B-file with output from the fragility curve plugin.
  - **update-outletts-data**: The [update-outletts-data](actions/link/update-outletts-data.md) action updates a RAS B-file with new observed flow data from an HDF file

## Extract
Extract actions help to extract various RAS HDF results into formats other than HDF5.
  - **ras-breach-extract**: The [ras-breach-extract](actions/extract/hdf/ras-breach-action.md) action extracts 2D Flow Area Connections data breaching conditions.
  - **ras-extract**: The [ras-extract](actions/extract/hdf/ras-extract-action.md) action is a tool to extract user-defined datasets and attributes from RAS HDF5 output.

## Utils
Utility actions
  - **copy-inputs**: The [copy-inputs](actions/utils/copy-inputs-action.md) action assists with bulk copying of model input files into the compute plugin.
  - **create-ras-tmp**: The [create-ras-tmp](actions/utils/create-ras-tmp.md) action creates a RAS TMP file from an input plan HDF file. The Linux RAS runner requires this file to run. 
  - **post-outputs**: The [post-outputs](actions/utils/post-outputs.md) action copies output files to an external store.
