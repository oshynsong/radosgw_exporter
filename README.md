# Radosgw Exporter

The prometheus exporter that scrapes the radosgw service information of a ceph cluster.


## Introduction

The radosgw is the object storage service of the [ceph](https://github.com/ceph/ceph) project, which is compatible with the AWS S3 API.
The high-level user needs to monitor the metrics such as thoughput, IOPS, and total object number and so on.
`radosgw_exporter` implements a exporter agent of the [prometheus](https://github.com/prometheus/) system in order to draw the figure with grafana.

## Metrics

It scrapes the following information of the radosgw service:

- `radosgw_bytes_sent_total`: accumulated sent bytes
- `radosgw_bytes_recv_total`: accumulated received bytes
- `radosgw_ops_total`: accumulated calling times of the given API
- `radosgw_ops_ok_total`: accumulated successfull calling times of the given API
- `radosgw_num_objects`: accumulated total object number
- `radosgw_capacity`: accumulated total space usage

Above each metric has three different labels: `user`, `bucket`, `api`.
User can sum up by one or more given label(s) to generate different figure with different concerns.


One can get the IOPS of uploading objects to the radosgw service on bucket `abc` with given user `admin` by the following query:

```
sum(delta(radosgw_bytes_recv_total{user='admin',bucket='abc'}[2m])) by (put_obj)
```

You can easily write other query statements to get your concerns of the radosgw service of a ceph cluster.

## Usage


### Installation

`go install github.com/oshynsong/radosgw_exporter`

### Running

The command arguments list as follows:

```
Usage of radosgw_exporter:
  -addr string
    	listen address for radosgw exporter (default "127.0.0.1:9129")
  -ak string
    	access key id of the admin user of radosgw service
  -endpoint string
    	endpoint of the radosgw service (default "127.0.0.1:8080")
  -path string
    	URL path for collecting radosgw metrics (default "/metrics")
  -sk string
    	secret access key of the admin user of radosgw service
```

One can just start the program with the endpoint and AK/SK of the radosgw service, and config
the prometheus like following snippet:

```
scrape_configs:
  - job_name: 'radosgw'
    static_configs
    - targets: ['127.0.0.1:9129']
```

---
Copyright @2018 [Oshyn Song](https://github.com/oshynsong)
