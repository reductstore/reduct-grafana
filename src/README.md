# ReductStore

## Overview

ReductStore is a time series data store optimized for **robotics** and **industrial IoT** workloads.
It ingests raw binary data such as **logs, JSON, CSV, and MCAP files** and organizes them with **time indexes and labels** for efficient querying, streaming, and retrieval.

This plugin integrates ReductStore as a **data source**, enabling analysis and visualization of unstructured, time-indexed data directly within Grafana.

## Features

* Ingests and queries data in its raw form (logs, JSON, CSV, MCAP).
* Attach metadata as labels for flexible filtering.
* Time-indexed storage for efficient temporal queries.
* Stream data from edge devices to the cloud.

## Requirements

* **Grafana**: v10.4.0 or higher
* **ReductStore**: v1.16 or higher (see [Releases](https://github.com/reductstore/reductstore/releases))
* Network access to your ReductStore instance

## Getting Started

1. Install the plugin from Grafana Marketplace or build from source.
2. Add **ReductStore Data Source** in Grafana’s *Connections → Data Sources* menu.
3. Configure your ReductStore server URL and authentication.
4. Start building dashboards with time-indexed queries.

## Screenshots

ReductStore can be queried using time ranges and labels (metadata) to filter data. Each record includes a timestamp and associated content (e.g., logs, JSON, CSV, MCAP) that can be queried and visualized.

For example, you can visualize labels:

![Query Labels](https://raw.githubusercontent.com/reductstore/reduct-grafana/main/src/img/screenshot-query-labels.png)

And you can use ReductStore's extensions to filter and extract specific fields from JSON, CSV, MCAP or logs. Here is an example of querying JSON content:

![Query Content](https://raw.githubusercontent.com/reductstore/reduct-grafana/main/src/img/screenshot-query-content.png)

For more information about querying JSON or CSV content, see the [ReductSelect Extension](https://www.reduct.store/docs/extensions/official/select-ext).

## Documentation

Official documentation is available at:
* [ReductStore Documentation](https://www.reduct.store/docs)

For more information about filtering and extracting fields from JSON, CSV or logs, see:
* [ReductSelect Extension](https://www.reduct.store/docs/extensions/official/select-ext)

For robotics-specific data, see:
* [ReductROS Extension](https://www.reduct.store/docs/extensions/official/ros-ext)

## Contributing

We welcome feedback, issues, and contributions!

* Open an issue on [GitHub Issues](https://github.com/reductstore/reduct-grafana/issues)
* Submit pull requests via [GitHub](https://github.com/reductstore/reduct-grafana)
