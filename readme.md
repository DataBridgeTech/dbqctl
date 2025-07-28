# DataBridge Quality Control

`dbqctl` is a free, open-source data quality CLI checker that provides a set of tools to profile, validate and test data in your data warehouse or databases. 
It is designed to be flexible, fast, easy to use and integrate seamlessly into your existing workflow.

---

## Features

- Effortless dataset import: pull in datasets (e.g. tables) from your chosen DWH with filters
- Comprehensive data profiling: gain instant insights into your data with automatic profiling, including:
  - Data Types
  - Null & Blank Counts
  - Min/Max/Avg Values
  - Standard Deviation
  - Most Frequent Values
  - Rows Sampling
- Data quality checks with built-in support for:
  - Row Count
  - Null Count
  - Average, Max, Min, Sum
- Flexible custom SQL checks: you can define and run your own SQL-based quality rules to meet unique business requirements.

## Supported databases
- [ClickHouse](https://clickhouse.com/)
- [PostgreSQL](https://www.postgresql.org/)

## Usage

### Installation

Download the latest binaries from [GitHub Releases](https://github.com/DataBridgeTech/dbqctl/releases).

### Configuration

Create `dbqctl` configuration file (default lookup directory is $HOME/.dbq.yaml or ./dbq.yaml). Alternatively,
you can specify configuration during the launch via `--config` parameter:

```bash
dbq --config /path/to/dbq.yaml import
```

```yaml
# dbq.yaml
version: "1"
datasources:
    - id: clickhouse
      type: clickhouse
      configuration:
        host: 0.0.0.0:19000
        port: 19000
        username: default
        password: changeme
        database: default
      datasets:
        - nyc_taxi.trips_small
    - id: pg
      type: postgresql
      configuration:
        host: localhost
        port: 5432
        username: default
        password: changeme
        database: uk_dbq_test
      datasets:
        - public.land_registry_price_paid_uk
        - public.test_table_name
```

### Checks example

```yaml
# checks.yaml
version: "1"
validations:
  # https://clickhouse.com/docs/getting-started/example-datasets/nyc-taxi
  - dataset: ch@[nyc_taxi.trips_small]
    # common pre-filter for every check, e.g. to run daily check only for yesterday
    where: "pickup_datetime > '2014-01-01'"
    checks:
      - id: row_count > 0
        description: "data should be present" # optional
        on_fail: error # optional (error, warn), default "error"

      - id: row_count between 100 and 30000
        description: "expected rows count"
        on_fail: warn

      - id: null_count(pickup_ntaname) == 0
        description: "no nulls are allowed in column: pickup_ntaname"

      - id: min(pickup_datetime) < now() - interval 3 day
        description: "min(pickup_datetime) should not be earlier than 3 days"

      - id: stddevPop(trip_distance) < 100_000
        description: "check stddev value"

      - id: sum(fare_amount) <= 10_000_000
        description: "sum of value"

      - id: countIf(trip_id == 1) == 1
        description: "check trip id"

      - id: raw_query
        description: "raw query quality test"
        query: |
          select countIf(trip_distance == 0) > 0 from {{table}} where 1=1
          
  # https://wiki.postgresql.org/wiki/Sample_Databases
  - dataset: pg@[public.land_registry_price_paid_uk]
    # exclude January for example
    where: "transfer_date >= '2025-02-01 00:00:00.000000'"
    checks:
      - id: row_count > 0
        description: "data should be present"
        on_fail: error
        
      - id: row_count between 200000 and 300000
        description: "expected rows count"
        on_fail: warn
        
      - id: min(price) > 0
        description: "min(price) should be greater than zero"
        
      - id: max(price) < 100000000
        description: "max(price) should be less than 100_000_000"
        
      - id: stddev_pop(price) < 500000
        description: "price stddev"
```

### Commands

```bash
$ dbqctl help

dbqctl is a CLI tool for profiling data and running quality checks across various data sources

Usage:
  dbqctl [command]

Available Commands:
  check       Runs data quality checks defined in a configuration file against a datasource
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  import      Connects to a data source and imports all available tables as datasets
  ping        Checks if the data source is reachable
  profile     Collects dataset`s information and generates column statistics
  version     Prints dbqctl and core lib version

Flags:
      --config string   config file (default is $HOME/.dbq.yaml or ./dbq.yaml)
  -h, --help            help for dbqctl
  -v, --verbose         enables verbose logging

Use "dbqctl [command] --help" for more information about a command.
```

### Quick usage examples
```bash
# check connection to datasource
$ dqbctl ping cnn-id

# automatically import datasets from datasource with applied filter and in-place update config file 
$ dbqctl import cnn-id --filter "reporting.*" --cfg checks.yaml --update-cfg

# run checks from checks.yaml file
$ dbqctl check --checks checks.yaml

# override default dbqctl config file
$ dbqctl --config /path/to/dbq.yaml import

# run dataset profile to collect general stats
$ dbqctl profile --datasource cnn-id --dataset table_name
```
