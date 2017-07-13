package digitaloceanexporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// A DigitalOceanSource is an interface which can retrieve information about a
// resources in a DigitalOcean account. It is implemented by
// *digitaloceanexporter.DigitalOceanService.
type DigitalOceanSource interface {
	Droplets() map[DropletCounter]int
	FloatingIPs() map[FlipCounter]int
	LoadBalancers() map[LoadBalancerCounter]int
	Tags() map[TagCounter]int
	Volumes() map[VolumeCounter]int
}

// A DigitalOceanCollector is a Prometheus collector for metrics regarding
// DigitalOcean.
type DigitalOceanCollector struct {
	Droplets      *prometheus.Desc
	FloatingIPs   *prometheus.Desc
	LoadBalancers *prometheus.Desc
	Tags          *prometheus.Desc
	Volumes       *prometheus.Desc

	dos DigitalOceanSource
}

// Verify that DigitalOceanCollector implements the prometheus.Collector interface.
var _ prometheus.Collector = &DigitalOceanCollector{}

// NewDigitalOceanCollector creates a new DigitalOceanCollector which collects
// metrics about resources in a DigitalOcean account.
func NewDigitalOceanCollector(dos DigitalOceanSource) *DigitalOceanCollector {
	return &DigitalOceanCollector{
		Droplets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "droplets", "count"),
			"Number of Droplets by region, size, and status.",
			[]string{"region", "size", "status"},
			nil,
		),
		FloatingIPs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "floating_ips", "count"),
			"Number of Floating IPs by region and status.",
			[]string{"region", "status"},
			nil,
		),
		LoadBalancers: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "load_balancers", "count"),
			"Number of Load Balancers by region and status.",
			[]string{"region", "status"},
			nil,
		),
		Tags: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tags", "count"),
			"Count of tagged resources by name and resource type.",
			[]string{"name", "resource_type"},
			nil,
		),
		Volumes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volumes", "count"),
			"Number of Volumes by region, size in GiB, and status.",
			[]string{"region", "size", "status"},
			nil,
		),

		dos: dos,
	}
}

// collect begins a metrics collection task for all metrics related to
// resources in a DigitalOcean account.
func (c *DigitalOceanCollector) collect(ch chan<- prometheus.Metric) {
	c.collectDropletCounts(ch)
	c.collectFipsCounts(ch)
	c.collectLoadBalancerCounts(ch)
	c.collectTagCounts(ch)
	c.collectVolumeCounts(ch)
}

func (c *DigitalOceanCollector) collectDropletCounts(ch chan<- prometheus.Metric) {
	for d, count := range c.dos.Droplets() {
		ch <- prometheus.MustNewConstMetric(
			c.Droplets,
			prometheus.GaugeValue,
			float64(count),
			d.region,
			d.size,
			d.status,
		)
	}
}

func (c *DigitalOceanCollector) collectFipsCounts(ch chan<- prometheus.Metric) {
	for fip, count := range c.dos.FloatingIPs() {
		ch <- prometheus.MustNewConstMetric(
			c.FloatingIPs,
			prometheus.GaugeValue,
			float64(count),
			fip.region,
			fip.status,
		)
	}
}

func (c *DigitalOceanCollector) collectLoadBalancerCounts(ch chan<- prometheus.Metric) {
	for lb, count := range c.dos.LoadBalancers() {
		ch <- prometheus.MustNewConstMetric(
			c.LoadBalancers,
			prometheus.GaugeValue,
			float64(count),
			lb.region,
			lb.status,
		)
	}
}

func (c *DigitalOceanCollector) collectTagCounts(ch chan<- prometheus.Metric) {
	for t, count := range c.dos.Tags() {
		ch <- prometheus.MustNewConstMetric(
			c.Tags,
			prometheus.GaugeValue,
			float64(count),
			t.name,
			t.resourceType,
		)
	}
}

func (c *DigitalOceanCollector) collectVolumeCounts(ch chan<- prometheus.Metric) {
	for v, count := range c.dos.Volumes() {
		ch <- prometheus.MustNewConstMetric(
			c.Volumes,
			prometheus.GaugeValue,
			float64(count),
			v.region,
			v.size,
			v.status,
		)
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *DigitalOceanCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.Droplets,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect sends the metric values for each metric pertaining to the DigitalOcean
// resources to the provided prometheus Metric channel.
func (c *DigitalOceanCollector) Collect(ch chan<- prometheus.Metric) {
	c.collect(ch)
}
