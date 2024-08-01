# fusionpbx_incoming_calls_exporter
[![Test, Build and publish app release](https://github.com/Apfelwurm/fusionpbx_incoming_calls_exporter/actions/workflows/test-and-build.yml/badge.svg)](https://github.com/Apfelwurm/fusionpbx_incoming_calls_exporter/actions/workflows/test-and-build.yml)

This is a prometheus exporter for fusionpbx, that reports the current sum of the incoming calls on all gateways.

## Installation

Download the latest release [here](https://github.com/Apfelwurm/fusionpbx_incoming_calls_exporter/releases/latest).

You can find a deb package there which can be installed by  `dpkg -i fusionpbx_incoming_calls_exporter_*_amd64.deb` on debian.

Alternativeley there is also a tar.gz file that can be unpacked using `tar xvf fusionpbx_incoming_calls_exporter-*-linux-amd64.tar.gz`. Make sure to set up some kind of Service yourself in that case.


## Prometheus Integration

Add the following job to your Prometheus configuration to scrape metrics from the FusionPBX CDR Exporter:

```yaml
scrape_configs:
  - job_name: 'fusionpbx_cdr_exporter'
    static_configs:
      - targets: ['localhost:8080'] # Adjust the target based on your setup
```

## Metrics

The following Prometheus metrics are exposed:

- `fusionpbx_individual_caller_destination_count`: Count of calls to individual caller destinations.
- `fusionpbx_total_caller_destination_count`: Total count of calls to all gateways.


## Configuration

### Environment Variables

- `FPB_IC_EXP_FUSION_CONFIG_FILE`: Path to the FusionPBX configuration file. Default is `/etc/fusionpbx/config.conf`.
- `FPB_IC_EXP_PORT`: Port on which to expose Prometheus metrics. Default is `8080`.


## Building it yourself

* Clone the repository
* install golang

Run `go build -o fusionpbx_incoming_calls_exporter` to build it.


## Contributing

Feel free to open issues, submit pull requests, or suggest new features. Your contributions are welcome!

<!-- ## Testing

* Clone the repository
* install golang

Run `go test` to run it. -->

