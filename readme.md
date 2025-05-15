# DataBridge Quality Core

`dbq` is a free, open-source data quality CLI checker that provides a set of tools to profile, validate and test data in your data warehouse or databases. 
It is designed to be flexible, fast, easy to use and integrate seamlessly into your existing workflow.

## Features

data profiling

v1 supported checks
---
row_count > 10
null_count(col) == 0
avg(col) <= 24.2
max(col) < 1000
min(col) == 0
sum(col) > 0
stddevPop(col) between 1 and 100_000_000
custom

## Supported databases
- [ClickHouse](https://clickhouse.com/)

## Usage

### Installation

Download the latest binaries from [GitHub Releases](https://github.com/DataBridgeTech/dbq/releases).

### Configuration

Create dbq configuration file (default lookup directory is $HOME/.dbq.yaml or ./dbq.yaml). Alternatively,
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
```

### Checks example

```yaml
# checks.yaml

```

### Commands

```bash
$ dbq help

dbq is a CLI tool for profiling data and running quality checks across various data sources

Usage:
  dbq [command]

Available Commands:
  check       Runs data quality checks defined in a configuration file against a datasource
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  import      Connects to a data source and imports all available tables as datasets
  ping        Checks if the data source is reachable
  profile     Collects dataset`s information and generates column statistics
  version     Prints dbq version

Flags:
      --config string   config file (default is $HOME/.dbq.yaml or ./dbq.yaml)
  -h, --help            help for dbq
  -v, --verbose         Enables verbose logging

Use "dbq [command] --help" for more information about a command.
```

### Quick start
- dqb ping cnn-id
- dbq import cnn-id --filter "reporting.*" --cfg checks.yaml --update-cfg
- dbq check --cfg checks.yaml
- dbq --config /Users/artem/code/dbq/dbq.yaml import 
- dbq profile --datasource cnn-id --dataset table_name