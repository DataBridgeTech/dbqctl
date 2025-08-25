# DataBridge Quality Control

`dbqctl` is a free, open-source data quality CLI checker that provides a set of tools to profile, validate and test data in your data warehouse or databases. 
It is designed to be flexible, fast, easy to use and integrate seamlessly into your existing workflow.

---

## Features

- Effortless dataset import: pull in datasets (e.g. tables) from your chosen DWH with filters
- Comprehensive data profiling - gain instant insights into your data with automatic profiling, including:
  - Columns, positions, Data Types
  - Null & Blank Counts
  - Min/Max/Avg Values
  - Standard Deviation
  - Most Frequent Values
  - Rows Sampling
- Data quality checks with built-in support for various checks:
  - Schema-level checks:
    - `expect_columns_ordered`: Validate table columns match an ordered list
    - `expect_columns`: Validate table has one of columns from unordered list
    - `columns_not_present`: Validate table doesn't have any columns from the list or matching pattern
  - Table-level:
    - `row_count`: Count of rows in the table
    - `raw_query`: Custom SQL query for complex validations
  - Column-level:
    - `not_null`: Check for null values in a column
    - `freshness`: Check data recency based on timestamp column
    - `uniqueness`: Check for unique values in a column
    - `min/max`: Minimum and maximum values for numeric columns
    - `sum`: Sum of values in a column
    - `avg`: Average of values in a column
    - `stddev`: Standard deviation of values in a column
- Flexible custom SQL checks: you can define and run your own SQL-based quality rules to meet unique business requirements.

## Supported databases
- [ClickHouse](https://clickhouse.com/)
- [PostgreSQL](https://www.postgresql.org/)
- [MySQL](https://www.mysql.com/)

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

Refer to [checks.yaml](./checks.yaml) example for full configuration overview. 

```yaml
# checks.yaml
version: "1"
validations:
  # https://clickhouse.com/docs/getting-started/example-datasets/nyc-taxi
  - dataset: ch@[nyc_taxi.trips_small, nyc_taxi.trips_full]
    # common pre-filter for every check, e.g. to run daily check only for yesterday
    where: "pickup_datetime > '2014-01-01'"
    checks:
      # schema-level checks
      - schema_check:
          expect_columns_ordered:
            columns_order: [trip_id, pickup_datetime, dropoff_datetime, trip_distance, fare_amount]
        desc: "Ensure table columns are in the expected order"
        on_fail: error

      - schema_check:
          expect_columns:
            columns: [trip_id, fare_amount]
        desc: "Ensure required columns exist"
        on_fail: error

      - schema_check:
          columns_not_present:
            columns: [credit_card_number, credit_card_cvv]
            pattern: "pii_*"
        desc: "Ensure PII and credit card info is not present in the table"
        on_fail: error

      # table-level checks
      - row_count between 1000 and 50000:
          desc: "Dataset should contain a reasonable number of trips"
          on_fail: error

      # column existence and nullability
      - not_null(trip_id):
          desc: "Trip ID is mandatory"
      - not_null(pickup_datetime)
      - not_null(dropoff_datetime)

      # data freshness
      - freshness(pickup_datetime) < 7d:
          desc: "Data should be no older than 7 days"
          on_fail: warn

      # uniqueness constraints
      - uniqueness(trip_id):
          desc: "Trip IDs must be unique"
          on_fail: error

      # numeric validations
      - min(trip_distance) >= 0:
          desc: "Trip distance cannot be negative"
      - max(trip_distance) < 1000:
          desc: "Maximum trip distance seems unrealistic"
          on_fail: warn
      - avg(trip_distance) between 1.0 and 20.0:
          desc: "Average trip distance should be reasonable"
      - stddev(trip_distance) < 100:
          desc: "Trip distance variation should be within normal range"

      # fare validations
      - min(fare_amount) > 0:
          desc: "Fare amount should be positive"
      - max(fare_amount) < 1000:
          desc: "Maximum fare seems too high"
      - sum(fare_amount) between 10000 and 10000000:
          desc: "Total fare amount should be within expected range"

      # custom validation with raw query
      - raw_query:
          desc: "Check for trips with zero distance but positive fare"
          query: "select count() from {{dataset}} where trip_distance = 0 and fare_amount > 0"
          on_fail: warn

  # https://wiki.postgresql.org/wiki/Sample_Databases
  - dataset: pg@[public.land_registry_price_paid_uk]
    # exclude January for example
    where: "transfer_date >= '2025-02-01 00:00:00.000000'"
    checks:
      # schema validation
      - schema_check:
          expect_columns_ordered:
            columns: [transaction_id, price, transfer_date, property_type, address]
        desc: "Validate expected column order for data consistency"
        on_fail: warn

      - schema_check:
          expect_columns:
            columns: [transaction_id, price, property_type]
        desc: "Ensure critical columns exist"
        on_fail: error

      - row_count() between 100 and 100000:
          desc: "Recent property transactions should be within expected volume"

  # https://github.com/datacharmer/test_db
  - dataset: mysql@[employees.salaries]
    checks:
      - row_count between 100 and 10000:
          desc: "Monthly order volume should be within business expectations"
          on_fail: warn
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
# check connection for datasource
$ dqbctl ping -d cnn-id

# check connection for all configured datasources
$ dqbctl ping

# automatically import datasets from datasource with applied filter and in-place update config file 
$ dbqctl import -d cnn-id --filter "reporting" --update-config

# run checks from checks.yaml file
$ dbqctl check --checks ./checks.yaml

# override default dbqctl config file
$ dbqctl --config /path/to/dbq.yaml import

# run dataset profile to collect general stats (limit concurrent jobs to 8)
$ dbqctl profile -d cnn-id --dataset table_name -j 8
```
