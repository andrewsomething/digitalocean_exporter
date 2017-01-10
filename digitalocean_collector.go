package digitalocean_exporter

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// A DigitalOceanSource is an interface which can retrieve information about a
// DigitalOcean account. It is implemented by *digitalocean_exporter.DigitalOceanService.
type DigitalOceanSource interface {
	Droplets() (map[DropletCounter]int, error)
	Volumes() (map[VolumeCounter]int, error)
}

// A DigitalOceanCollector is a Prometheus collector for metrics regarding
// DigitalOcean.
type DigitalOceanCollector struct {
	Droplets *prometheus.Desc
	Volumes  *prometheus.Desc

	dos DigitalOceanSource
}

// Verify that DigitalOceanCollector implements the prometheus.Collector interface.
var _ prometheus.Collector = &DigitalOceanCollector{}

// NewDigitalOceanCollector creates a new DigitalOceanCollector which collects
// metrics about a DgitialOcean account.
func NewDigitalOceanCollector(dos DigitalOceanSource) *DigitalOceanCollector {
	return &DigitalOceanCollector{
		Droplets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "droplets", "count"),
			"Number of Droplets by region, size, and status.",
			[]string{"region", "size", "status"},
			nil,
		),
		Volumes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volumes", "count"),
			"Number of Volumes by region, size, and status.",
			[]string{"region", "size", "status"},
			nil,
		),

		dos: dos,
	}
}

// collect begins a metrics collection task for all metrics related to
// a DigitalOcean account.
func (c *DigitalOceanCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if count, err := c.collectDropletCounts(ch); err != nil {
		return count, err
	}
	if count, err := c.collectVolumeCounts(ch); err != nil {
		return count, err
	}
	// if desc, err := c.collectActiveDownloads(ch); err != nil {
	//     return desc, err
	// }

	return nil, nil
}

func (c *DigitalOceanCollector) collectDropletCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	droplets, err := c.dos.Droplets()
	if err != nil {
		return c.Droplets, err
	}

	for d, count := range droplets {
		ch <- prometheus.MustNewConstMetric(
			c.Droplets,
			prometheus.GaugeValue,
			float64(count),
			d.region,
			d.size,
			d.status,
		)
	}

	return nil, nil
}

func (c *DigitalOceanCollector) collectVolumeCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	volumes, err := c.dos.Volumes()
	if err != nil {
		return c.Volumes, err
	}

	for v, count := range volumes {
		ch <- prometheus.MustNewConstMetric(
			c.Volumes,
			prometheus.GaugeValue,
			float64(count),
			v.region,
			v.size,
			v.status,
		)
	}

	return nil, nil
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

// Collect sends the metric values for each metric pertaining to the rTorrent
// downloads to the provided prometheus Metric channel.
func (c *DigitalOceanCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] failed collecting DigitalOcean metric %v: %v", desc, err)
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}
}
