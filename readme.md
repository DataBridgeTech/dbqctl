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
- [ ] config parser
- [ ] clickhouse support
  - [x] ping
  - [x] import datasets
  - [x] profile dataset
    - [x] rows in table
    - [x] min, max, avg, stddev for numeric columns
    - [x] count of nulls and blanks
    - [x] most frequent value in column
  - [ ] run checks
- [ ] implement support for custom sql check 
- [ ] implement aliases for common checks based on raw sql check
- [ ] basic cross validation (dataset is defined)
- [ ] fix cmd descriptions
- [ ] review todos
- [ ] review logs

## 0.2
- config validation
- add postgres support
- CLI for adding more checks
- schema changes checks


---

### clickhouse 

```bash
docker run -d -p 18123:8123 -p19000:9000 -e CLICKHOUSE_PASSWORD=changeme --name some-clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server
```