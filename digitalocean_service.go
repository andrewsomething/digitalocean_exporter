package digitaloceanexporter

import (
	"strconv"

	"github.com/digitalocean/godo"
)

// DigitalOceanService is a wrapper around godo.Client.
type DigitalOceanService struct {
	C *godo.Client
}

// DropletCounter is a struct holding information about a Droplet.
type DropletCounter struct {
	status string
	region string
	size   string
}

// FlipCounter is a struct holding information about a Floating IP.
type FlipCounter struct {
	status string
	region string
}

// VolumeCounter is a struct holding information about a Block Storage Volume.
type VolumeCounter struct {
	status string
	region string
	size   string
}

var pageOpt = &godo.ListOptions{
	Page:    1,
	PerPage: 200,
}

// Droplets retrieves a count of Droplets grouped by status, size, and region.
func (s *DigitalOceanService) Droplets() (map[DropletCounter]int, error) {
	droplets, _, err := s.C.Droplets.List(pageOpt)

	counters := make(map[DropletCounter]int)

	for _, d := range droplets {
		c := DropletCounter{
			d.Status,
			d.Region.Slug,
			d.Size.Slug,
		}
		counters[c]++
	}

	return counters, err
}

// FloatingIPs retrieves a count of Floating IPs grouped by status and region.
func (s *DigitalOceanService) FloatingIPs() (map[FlipCounter]int, error) {
	fips, _, err := s.C.FloatingIPs.List(pageOpt)

	counters := make(map[FlipCounter]int)

	for _, fip := range fips {
		var status string

		switch {
		case fip.Droplet == nil:
			status = "unassigned"
		default:
			status = "assigned"
		}
		c := FlipCounter{
			status,
			fip.Region.Slug,
		}
		counters[c]++
	}

	return counters, err
}

// Volumes retrieves a count of Volumes grouped by status, size, and region.
func (s *DigitalOceanService) Volumes() (map[VolumeCounter]int, error) {
	volumes, _, err := s.C.Storage.ListVolumes(pageOpt)

	counters := make(map[VolumeCounter]int)

	for _, v := range volumes {
		var status string

		switch {
		case len(v.DropletIDs) > 0:
			status = "attached"
		default:
			status = "unattached"
		}
		c := VolumeCounter{
			status,
			v.Region.Slug,
			strconv.FormatInt(v.SizeGigaBytes, 10),
		}
		counters[c]++
	}

	return counters, err
}
