# digitalocean_exporter
[![Build Status](https://travis-ci.org/andrewsomething/digitalocean_exporter.svg)](https://travis-ci.org/andrewsomething/digitalocean_exporter) [![Go Report Card](https://goreportcard.com/badge/github.com/andrewsomething/digitalocean_exporter)](https://goreportcard.com/report/github.com/andrewsomething/digitalocean_exporter) [![Docker Build Status](https://img.shields.io/docker/build/andrewsomething/digitalocean_exporter.svg)](https://hub.docker.com/r/andrewsomething/digitalocean_exporter/)



`digitalocean_exporter` is a Prometheus exporter for DigitalOcean resources.

Initially written as a way to explore Prometheus exporters, I am sharing
it in the hope that it will be useful to others. Matt Layher's
[`rtorrent_exporter`](https://github.com/mdlayher/rtorrent_exporter) was
a great help.

## Usage

`digitalocean_exporter` has a number of runtime options that can be set
using flags including the listen addresss and the path for the metrics
end point. A DigitalOcean API token is also required. It is recommended
to use a read-only token as `digitalocean_exporter` has no need for write
access to your account.

As calls to the DigitalOcean API can be expensive, `digitalocean_exporter`
maintains a local cache that is periodically refreshed based on the
`refresh-interval` value provided. The default is every 60 seconds.

```
$ ./digitalocean_exporter -help
Usage of ./digitalocean_exporter:
  -debug
        Print debug logs
  -listen string
        Listen address for DigitalOcean exporter (default "localhost:9292")
  -metrics-path string
        URL path for surfacing metrics (default "/metrics")
  -refresh-interval int
        Interval (in seconds) between subsequent requests against DigitalOcean API (default 60)
  -token string
        DigitalOcean API token (read-only)
  -v    Prints current digitalocean_exporter version
```

### Docker

This exporter is also available as a Docker image: [`andrewsomething/digitalocean_exporter`](https://hub.docker.com/r/andrewsomething/digitalocean_exporter/)

Example usage:

```
$ docker run -p 127.0.0.1:9292:9292 andrewsomething/digitalocean_exporter \
    -listen 0.0.0.0:9292 -token $DIGITALOCEAN_API_TOKEN
```

## Metrics

Here is an example of the metrics exposed by `digitalocean_exporter`:

```
$ curl --silent localhost:9292/metrics | grep digitalocean
# HELP digitalocean_droplets_count Number of Droplets by region, size, and status.
# TYPE digitalocean_droplets_count gauge
digitalocean_droplets_count{region="lon1",size="1gb",status="active"} 1
digitalocean_droplets_count{region="nyc2",size="1gb",status="active"} 1
digitalocean_droplets_count{region="nyc2",size="2gb",status="active"} 3
digitalocean_droplets_count{region="nyc3",size="16gb",status="active"} 1
digitalocean_droplets_count{region="nyc3",size="1gb",status="active"} 1
digitalocean_droplets_count{region="nyc3",size="1gb",status="off"} 1
digitalocean_droplets_count{region="nyc3",size="4gb",status="active"} 1
digitalocean_droplets_count{region="nyc3",size="512mb",status="active"} 6
digitalocean_droplets_count{region="nyc3",size="512mb",status="off"} 1
# HELP digitalocean_floating_ips_count Number of Floating IPs by region and status.
# TYPE digitalocean_floating_ips_count gauge
digitalocean_floating_ips_count{region="nyc3",status="assigned"} 1
digitalocean_floating_ips_count{region="nyc3",status="unassigned"} 1
# HELP digitalocean_load_balancers_count Number of Load Balancers by region and status.
# TYPE digitalocean_load_balancers_count gauge
digitalocean_load_balancers_count{region="nyc3",status="active"} 1
# HELP digitalocean_query_duration_seconds Time elapsed while querying the DigitalOcean API in seconds.
# TYPE digitalocean_query_duration_seconds gauge
digitalocean_query_duration_seconds 4.806081399
# HELP digitalocean_tags_count Count of tagged resources by name and resource type.
# TYPE digitalocean_tags_count gauge
digitalocean_tags_count{name="frontend",resource_type="droplets"} 0
digitalocean_tags_count{name="production",resource_type="droplets"} 7
digitalocean_tags_count{name="prometheus",resource_type="droplets"} 1
digitalocean_tags_count{name="swarm",resource_type="droplets"} 2
# HELP digitalocean_volumes_count Number of Volumes by region, size in GiB, and status.
# TYPE digitalocean_volumes_count gauge
digitalocean_volumes_count{region="fra1",size="100",status="unattached"} 1
digitalocean_volumes_count{region="nyc1",size="100",status="attached"} 1
```
