# RAS Extract Action

## Description

The **RAS Extract Action** is designed to extract user-defined datasets from RAS HDF5 files following a successful HEC-RAS model run. The extracted data can be written to more accessible formats such as JSON and Console STDOUT. Future support for cloud-native array stores (e.g., TileDB or Zarr) is planned.



# Implementation Details

The RAS Extract Action supports three primary extraction methods:

1. **Dataset Extraction**
2. **Group Extraction**
3. **Attribute Extraction**

Each method allows users to define how data should be extracted, processed, and outputted based on their specific needs.

## Process Flow

1. The action is triggered with a configuration specifying the extraction method and parameters.
2. The HDF5 file is accessed using the provided `modelPrefix` and `plan`.
3. Based on the extraction type (dataset, group, or attribute), the appropriate data is read from the HDF5 file.
4. Data processing occurs according to the specified `postprocess` functions if `writesummary` is enabled.
5. The results are either written directly to the specified output or accumulated for later writing via `outputDataSource`.

## Configuration

### Environment

### AttributesAttributes

These parameters are required for all extraction actions and can be included in either payload attributes or action attributes:

```json
{
  "modelPrefix": "muncie",
  "plan": "04"
}
```
| Attribute             | Description |
|-----------------------|-------------|
| `modelPrefix`         | The file name of the ras mdoel without any extension information |
| `plan`                | The plan number as a two digit string (e.g. plan 4 is '04')  |

---

#### 1. Dataset Extraction Attributes

A dataset can be extracted from an HDF5 file into the specified output format. Both raw data and summary values (max or min) are supported.

### Configuration Example:

```json
{
  "name": "ras-extract",
  "type": "extract",
  "description": "refline-flow",
  "attributes": {
    "outputformat": "json",
    "datapath": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
    "coldata": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name",
    "postprocess": ["max"],
    "writedata": false,
    "writesummary": true,
    "datatype": "float32",
    "block-name": "refline_peak_flow",
    "outputDataSource": "metadataout"
  }
}
```

#### Attributes

| Attribute             | Description |
|-----------------------|-------------|
| `name`                | Must be set to `"ras-extract"` to instruct the RAS Runner to execute this action. |
| `type`                | Must be set to `"extract"`; it defines the type of operation. |
| `description`         | A user-defined description for the extraction task. |
| `attributes`          | Contains all configuration options related to the extraction. |

#### Action-Attributes (`attributes`)

| Action-Attribute           | Description |
|-------------------------|-------------|
| `outputformat`          | Specifies output format. Supported values: `"json"` or `"console"`. The console writer prints directly to STDOUT; JSON writes to a structured document. |
| `datapath`              | Internal path in the HDF5 file pointing to the dataset to be extracted. |
| `coldata`               | Optional. Path to a separate string array dataset containing column names (e.g., `/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Name`). |
| `colnames`              | Optional. An array of strings defining column names if they are not stored in a dataset. Example: `["stage(ft)", "flow(cfs)"]`. |
| `writedata`             | Optional boolean flag. If `true`, exports the entire dataset. Default is `false`. |
| `writesummary`          | Optional boolean flag. If `true`, calculates summary statistics per column. Default is `false`. |
| `postprocess`           | Array of strings specifying post-processing functions when `writesummary=true`. Supported values: `"max"` and `"min"`. |
| `datatype`              | Required (planned to be removed in future versions). Supported data types: `"float32"`, `"float64"`, `"int32"`, `"int64"`. |
| `block-name`            | Name used for identifying the block in the output. |
| `accumulate-results`    | Boolean flag indicating whether to accumulate results rather than write them immediately. When set, extraction data is stored temporarily until explicitly written using `outputDataSource`. Mutually exclusive with `outputDataSource`. |
| `outputDataSource`      | Reference to a data source definition that specifies where accumulated or immediate output should be written. Example:
```json
{
  "name": "metadataout",
  "paths": {
    "extract": "ras-event-data.json"
  },
  "store_name": "FFRD"
}
```

> ⚠️ Note: Either `accumulate-results` or `outputDataSource` must be specified, but not both.

---

## 2. Group Extraction

Groups within the HDF5 file can be enumerated and each dataset within the group can be extracted.

### Configuration Example:

```json
{
  "name": "ras-extract",
  "type": "extract",
  "description": "boundary condition peak flow and stage",
  "attributes": {
    "outputformat": "json",
    "grouppath": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions",
    "colnames": ["stage(ft)", "flow(cfs)"],
    "match": "",
    "exclude": "Flow per Face|Stage per Face|Flow per Cell",
    "postprocess": ["max", "min"],
    "writesummary": true,
    "datatype": "float32",
    "block-name": "bcline_peak",
    "accumulate-results": true
  }
}
```

### Additional Attributes for Group Extraction

| Attribute             | Description |
|-----------------------|-------------|
| `grouppath`           | Path to the HDF5 group whose datasets will be extracted. |
| `groupsuffix`         | Optional string appended to each group object path during reading. Used when datasets reside in subdirectories of group objects (e.g., for 2D flow areas). Example:
```json
"grouppath": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/",
"groupsuffix": "Reference Lines/Flow"
```
| `match`               | Optional regular expression to select specific dataset objects within the group. |
| `exclude`             | Optional regular expression to filter out unwanted dataset objects from selection. When both `match` and `exclude` are present, `match` is applied first, followed by `exclude`. |

> 📝 Note: The behavior of matching/excluding follows standard regex logic.

---

## 3. Attribute Extraction

Attributes from datasets or groups can be extracted and formatted as key-value pairs in a record-style output.

### Configuration Example:

```json
{
  "name": "ras-extract",
  "type": "extract",
  "description": "summary attributes",
  "attributes": {
    "outputformat": "json",
    "datapath": "/Results/Unsteady/Summary",
    "attributes": true,
    "colnames": [
      "Computation Time DSS",
      "Computation Time Total",
      "Maximum WSEL Error",
      "Maximum number of cores",
      "Run Time Window",
      "Solution",
      "Time Solution Went Unstable",
      "Time Stamp Solution Went Unstable"
    ],
    "block-name": "summary_attributes",
    "accumulate-results": true
  }
}
```

### Additional Attributes for Attribute Extraction

| Attribute             | Description |
|-----------------------|-------------|
| `datapath` or `grouppath` | One of these must be provided to identify the target object whose attributes are to be extracted. |
| `attributes`          | Boolean flag indicating that attribute extraction should occur instead of data extraction. Set to `true`. |

> ⚠️ Important: Either `datapath` or `grouppath` is required.

---

### Inputs

- HDF5 file containing RAS model output data
- Configuration specifying extraction parameters

### Input Data Sources

- HDF5 file path derived from `modelPrefix` and `plan`

### Outputs

- Extracted data in JSON or console format
- Accumulated results ready for final writing via the configured output data source
- Direct output to console when `outputformat` is `"console"`

## Output Data Format
- For JSON output, refer to the [JSON Specification](#json-spec) for schema details.


## Error Handling

- Invalid `outputformat` values will result in an error
- Missing required attributes will cause configuration validation failures
- Invalid HDF5 paths will lead to read errors
- Mutually exclusive `accumulate-results` and `outputDataSource` will trigger an error

## Usage Notes

- Either `accumulate-results` or `outputDataSource` must be specified, but not both
- The `grouppath` and `match`/`exclude` parameters allow for fine-grained dataset selection
- When using `accumulate-results`, ensure that `outputDataSource` is configured to write the final results

## Future Enhancements

Support for additional formats like **TileDB** and **Zarr** is planned. These will allow seamless integration with modern cloud-native data stores and improve scalability for large datasets.

--- 

### Patterns and Best Practices
#### regular expression pattern matching for groups and datasets
When using group extractions, regex matching allows fine-grained control over which datasets are processed:
```json
"match": ".*Flow.*",
"exclude": ".*per Face.*"
```
This example matches all datasets containing “Flow” but excludes those with “per Face”.

#### Accumulation vs Direct Writing
Use `accumulate-results: true` when you want to collect multiple extractions before writing them to a final file. This is useful for aggregating results across many sub-extractions.

To write accumulated data, specify an output dataset in the configuration:
```json
outputs[
  {
    "name": "metadataout",
    "paths": {
      "extract": "final_output.json"
    },
    "store_name": "FFRD"
  }
]

...

actions[
  {
    "name": "ras-extract",
    "type": "extract",
    "description": "refline-flow",
    "attributes": {
      //...extract config
      "outputDataSource": "metadataout"
    }
  }
]
```

