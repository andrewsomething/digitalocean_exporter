# digitalocean_exporter

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

```
$ ./digitalocean_exporter
Usage of ./digitalocean_exporter:
  -listen string
        Listen address for DigitalOcean exporter (default "localhost:9292")
  -metrics-path string
        URL path for surfacing metrics (default "/metrics")
  -token string
        DigitalOcean API token (read-only)
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
# HELP digitalocean_volumes_count Number of Volumes by region, size in GiB, and status.
# TYPE digitalocean_volumes_count gauge
digitalocean_volumes_count{region="fra1",size="100",status="unattached"} 1
digitalocean_volumes_count{region="nyc1",size="100",status="attached"} 1
```
