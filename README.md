# Proton Grafana Datasource
Grafana data source plugin to connect to Proton and visualize streaming or batch queries.

## Overview / Introduction

[Proton](https://github.com/timeplus-io/proton) is a unified streaming and historical data processing engine in a single binary. It helps data engineers and platform engineers solve complex real-time analytics use cases, and powers the Timeplus streaming analytics platform.

## Requirements

Grafana v10.0.3 or above

A running Proton instance with TCP port 8463 (for database connection) and HTTP port 3218 (for query analazyer REST API).

## Getting Started

### Use the pre-built Docker Compose
The [docker-compose.yaml](docker-compose.yaml) in this folder ships a Grafana container, with the proton plugin pre-installed, as well as a data generator.

You start it with `docker compose up` and go to http://localhost:3000 to add a new data source for Proton, using `proton` as the hostname (because the Grafana container is running in the Docker Compose network. `proton` is the internal hostname for Proton database.)
Then create a new dashboard by copying the content of [carsharing_grafana_dashboard.json](carsharing_grafana_dashboard.json) in the dashboard JSON model.

### Use your own Grafana deployment

In the navigation menu, choose Connections -> Add new connection.

Search for Proton and accept the default settings (localhost,port 8463 and 3218 as proton connection). This plugin is expected to run in localhost or trusted network. Username and password for Proton will be added later. For Timeplus Cloud, API Key is supported for REST API, but this Grafana plugin doesn't support Timeplus Cloud at this point.

Create a new dashboard or explore data with this Proton data source.

There are unbounded streaming query and bounded historical query in proton, all queries like `select count(*) from stream_name` are streaming queries, and adding `table` function to the stream name will turn the query into bounded query, e.g. `select count(*) from table(stream_name)`.

![query editor](src/img/query.png)


## Development

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   brew install mage
   mage -v
   ```

   mage build:linuxARM will fail, which is okay. Only 64bit OS are supported.

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```
### Frontend

1. Install dependencies

   ```bash
   npm install
   ```

2. Build plugin in development mode and run in watch mode (Ctrl+C to stop)

   ```bash
   npm run dev
   ```

3. Build plugin in production mode

   ```bash
   npm run build
   ```

4. Sign the plugin
   ```bash
   export GRAFANA_ACCESS_POLICY_TOKEN=<YOUR_ACCESS_POLICY_TOKEN>
   npm run sign
   ```

4. Distribute the plugin
   ```bash
   make package
   ```

## Learn more

Below you can find source code for existing app plugins and other related documentation.

- [Basic data source plugin example](https://github.com/grafana/grafana-plugin-examples/tree/master/examples/datasource-basic#readme)
- [`plugin.json` documentation](https://grafana.com/developers/plugin-tools/reference-plugin-json)
- [How to sign a plugin?](https://grafana.com/docs/grafana/latest/developers/plugins/sign-a-plugin/)
