# JSON Specification Document for Data Structure

## Overview

This document describes the JSON format used to represent structured data in a standardized way, specifically designed for scientific simulation outputs. The format supports various types of data including boundary conditions, breach records, reference lines, reference points, and summary attributes.

### File Format
- **Format**: JSON (JavaScript Object Notation)
- **Encoding**: UTF-8
- **Structure**: Hierarchical map with arrays of data objects
- **Naming Convention**: `.json` extension

### DataObject Properties
Each data object must have:
1. A valid dataset path
2. Optional but consistent columns array when present
3. Optional summary statistics (max/min values)
4. Optional structured record data
5. Optional two-dimensional numerical data

### Validation Rules
1. **Required Fields**:
   - Every data object must have a `dataset` field
   - The structure of the file is a valid JSON map with proper nesting

2. **Optional Fields**:
   - `columns`, `record`, `data`, and `summaries` are optional
   - When present, they must follow their respective formats
   - Summary arrays must have matching lengths for max and min values
   - Column arrays must contain strings
   - Data arrays must be two-dimensional with numerical values

3. **Data Consistency**:
   - If both `max` and `min` arrays are provided, they should have the same length
   - Column count in `columns` should match array lengths in `data`
   - All numeric values in arrays should be valid floating-point numbers

4. **Null Handling**:
   - Null values in `record` indicate missing data for that field
   - Empty arrays are valid when no data exists

## File Structure

The root of the JSON file is a **map/object** where:
- Keys are **strings** representing categories or groups (e.g., `"bcline_peak"`, `"breach_records"`)
- Values are **arrays** of objects containing individual data entries

Each category array contains multiple data object entries, each with a unique key that identifies the specific dataset within that category.

## Data Object Structure

Each entry in a category array follows this structure:

```json
{
  "key": {
    "dataset": "/path/to/dataset",
    "columns": ["column1", "column2", ...],
    "record": {
      // Key-value pairs of record data
    },
    "data": [[], [], ...],
    "summaries": {
      "max": [value1, value2, ...],
      "min": [value1, value2, ...]
    }
  }
}
```

## Field Definitions

### Root Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `dataset` | string | Yes | Path to the dataset location |
| `columns` | array of strings | No | Column names for data fields |
| `record` | object | No | Key-value pairs of record information |
| `data` | array of arrays | No | Two-dimensional array of numeric values |
| `summaries` | object | No | Summary statistics with max and min values |

### Summary Object

The `summaries` object contains:
```json
{
  "max": [float, float, ...],
  "min": [float, float, ...]
}
```

- **max**: Array of maximum values for each column
- **min**: Array of minimum values for each column

## Detailed Examples by Category

### Boundary Condition Peak Data (`bcline_peak`)

```json
{
  "bcline_peak": [
    {
      "bc_bardwell_s010_base": {
        "dataset": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/bc_bardwell_s010_base",
        "columns": ["stage(ft)", "flow(cfs)"],
        "summaries": {
          "max": [421.44797, 0],
          "min": [420.73865, 0]
        }
      }
    }
  ]
}
```

### Breach Records (`breach_records`)

```json
{
  "breach_records": [
    {
      "nid_tx00001": {
        "dataset": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/nid_tx00001/2D Hyd Conn",
        "record": {
          "BreachIndex": -1,
          "BreachProgressionDuration": 0,
          "BreachStartTime": null,
          "Breached": false,
          "Event": "1",
          "FlowArea2D": "bardwell-creek",
          "HWAtBreach": null,
          "MaxBottomWidth": null,
          "MaxFlow": null,
          "MaxHW": 421.44797,
          "MaxTW": 390.91776,
          "SaConn": "nid_tx00001",
          "TWAtBreach": null
        }
      }
    }
  ]
}
```

### Reference Line Flow (`refline_peak_flow`)

```json
{
  "refline_peak_flow": [
    {
      "Flow": {
        "dataset": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Lines/Flow",
        "columns": ["gage_usgs_08063590_waxahachiecrk|bardwell-creek", ...],
        "summaries": {
          "max": [1278.9534, 600.0005, 0.0095818015]
        }
      }
    }
  ]
}
```

### Reference Point Velocity (`refpoint_velocity`)

```json
{
  "refpoint_velocity": [
    {
      "Velocity": {
        "dataset": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Reference Points/Velocity",
        "columns": ["ref-pt_bardwell-lake|bardwell-creek", ...],
        "summaries": {
          "max": [0.0012651726, 0.0011478694, 2.9275854, 0.027877262, 1.4710057],
          "min": [0, 0, 0, 0, 0]
        }
      }
    }
  ]
}
```

### Summary Attributes (`summary_attributes`)

```json
{
  "summary_attributes": [
    {
      "attributes": {
        "dataset": "/Results/Unsteady/Summary",
        "record": {
          "Computation Time DSS": "00:00:00",
          "Computation Time Total": "01:08:57",
          "Maximum WSEL Error": 0,
          "Maximum number of cores": 2,
          "Run Time Window": "17SEP2025 14:37:50 to 17SEP2025 15:46:46",
          "Solution": "Unsteady Finished Successfully",
          "Time Solution Went Unstable": null,
          "Time Stamp Solution Went Unstable": "Not Applicable"
        }
      }
    }
  ]
}
```

## Data Type Specifications

### String Fields
- `dataset`: Full path to dataset location
- Column names: Descriptive strings representing data columns
- All other string fields: Free-form text or identifiers

### Number Fields
- All numeric values are `float64` type
- Arrays in `max` and `min` contain floating-point numbers
- Null values represent missing data

### Array Fields
- `columns`: Array of strings (column names)
- `max`, `min`: Arrays of floats representing summary statistics
- `data`: Two-dimensional array of floats
- `record`: Object with key-value pairs of any type

## JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": {
    "type": "array",
    "items": {
      "type": "object",
      "minProperties": 1,
      "maxProperties": 1,
      "additionalProperties": {
        "$ref": "#/definitions/DataObject"
      }
    }
  },
  "definitions": {
    "Summary": {
      "type": "object",
      "properties": {
        "max": {
          "type": "array",
          "items": {
            "type": "number"
          }
        },
        "min": {
          "type": "array",
          "items": {
            "type": "number"
          }
        }
      },
      "additionalProperties": false
    },
    "DataObject": {
      "type": "object",
      "properties": {
        "dataset": {
          "type": "string"
        },
        "columns": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "record": {
          "type": "object",
          "additionalProperties": true
        },
        "data": {
          "type": "array",
          "items": {
            "type": "array",
            "items": {
              "type": "number"
            }
          }
        },
        "summaries": {
          "$ref": "#/definitions/Summary"
        }
      },
      "required": ["dataset"],
      "additionalProperties": false
    }
  }
}
```

## Detailed Documentation

### Root Level Structure

The top-level JSON object is a map where:
- Keys are category identifiers (strings)
- Values are arrays of data objects
- Each array contains one or more objects with a single key-value pair

### DataObject Structure

Each `DataObject` represents a single data record and contains the following fields:

#### Required Fields
- **dataset** (`string`): Path identifier for the dataset
  - Example: `/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/bc_bardwell_s010_base`

#### Optional Fields
- **columns** (`array[string]`): List of column names
  - Example: `["stage(ft)", "flow(cfs)"]`
- **record** (`object`): Key-value pairs representing structured data
  - Can contain any valid JSON data types (string, number, boolean, null, array, object)
- **data** (`array[array[number]]`): Two-dimensional array of numerical data points
  - Example: `[[1.0, 2.0], [3.0, 4.0]]`
- **summaries** (`object`): Contains maximum and minimum values for each column
  - **max** (`array[number]`): Maximum values for each column
  - **min** (`array[number]`): Minimum values for each column

### Example Structure

```json
{
  "bcline_peak": [
    {
      "bc_bardwell_s010_base": {
        "dataset": "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions/bc_bardwell_s010_base",
        "columns": ["stage(ft)", "flow(cfs)"],
        "summaries": {
          "max": [421.44797, 0],
          "min": [420.73865, 0]
        }
      }
    }
  ]
}
```



