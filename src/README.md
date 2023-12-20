
## Overview / Introduction

Grafana data source plugin to connect to Proton and visualize streaming or batch queries.

[Proton](https://github.com/timeplus-io/proton) is a unified streaming and historical data processing engine in a single binary. It helps data engineers and platform engineers solve complex real-time analytics use cases, and powers the Timeplus streaming analytics platform.

## Requirements
Grafana v10.0.3 or above

A running Proton instance with TCP port 8463 (for database connection) and HTTP port 3218 (for query analazyer REST API).

## Getting Started

### Use the pre-built Docker Compose
The [docker-compose.yaml](https://github.com/timeplus-io/proton-grafana-source/blob/main/docker-compose.yaml) ships a Grafana container, with the proton plugin pre-installed, as well as a data generator.

You start it with `docker compose up` and go to http://localhost:3000 to add a new data source for Proton, using `proton` as the hostname (because the Grafana container is running in the Docker Compose network. `proton` is the internal hostname for Proton database.)

### Use your own Grafana deployment

Before the plugin is approved by Grafana, you need to set your Grafana running in development mode via changing /usr/local/etc/grafana/grafana.ini, setting `app_mode = development`

In the navigation menu, choose Connections -> Add new connection.

Search for Proton and accept the default settings (localhost,port 8463 and 3218 as proton connection). This plugin is expected to run in localhost or trusted network. Username and password for Proton will be added later. For Timeplus Cloud, API Key is supported for REST API, but this Grafana plugin doesn't support Timeplus Cloud at this point.

Create a new dashboard or explore data with this Proton data source.

There are unbounded streaming query and bounded historical query in proton, all queries like `select count(*) from stream_name` are streaming queries, and adding `table` function to the stream name will turn the query into bounded query, e.g. `select count(*) from table(stream_name)`.
 
## Documentation
For more information about Timeplus Proton, please check https://docs.timeplus.com/proton.

## Contributing
https://github.com/timeplus-io/proton-grafana-source is open-sourced with Apache v2.0 License. Welcome to create issues or pull requests.