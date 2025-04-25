# DataBridge Quality Core

dbq is a data quality tool that provides a set of tools to validate and test data in your data pipeline. 
It is designed to be easy to use and integrate into your existing workflow.

## Help
- dqb ping cnn-id
- dbq import cnn-id --filter "reporting.*" --cfg checks.yaml --update-cfg
- dbq check --cfg checks.yaml
- dbq --config /Users/artem/code/dbq/dbq.yaml import 
- dbq profile --datasource cnn-id --dataset table_name

## 0.1
- [x] basic structure
- [x] define checks cfg v1
- [x] checks cfg parser v1
- [x] complete clickhouse support
  - [x] ping
  - [x] import datasets
  - [x] profile dataset
    - [x] rows in table
    - [x] min, max, avg, stddev for numeric columns
    - [x] count of nulls and blanks
    - [x] most frequent value in column
    - [x] JSON export
  - [x] run checks
- [x] implement support for custom sql check 
- [x] implement aliases for common checks based on raw sql check
- [x] fix cmd descriptions
- [x] review todos
- [x] improve output
- [ ] basic cross validation (dataset is defined)
- [ ] review logs
- [ ] review crashes (wrong arguments)
- [ ] default values (e.g. severity)
- [ ] quiet/verbose mode for logs
- [ ] docs

## 0.x
- config validation
- add postgres support
- CLI for adding more checks
- AirFlow integration (operator)
- output format flag

---

## Checks config specification
- raw_query(query = "...")
- row_count
- null_count(col)
- <aggr_function> <op> <rest>

### clickhouse 

```bash
docker run -d -p 18123:8123 -p19000:9000 -e CLICKHOUSE_PASSWORD=changeme --name some-clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server
```

# Supported Datasources
- Clickhouse

# dbq configuration

# checks configuration