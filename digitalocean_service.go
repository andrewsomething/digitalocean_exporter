package digitaloceanexporter

import (
	//	"log"
	"strconv"

	"github.com/digitalocean/godo"
)

type DigitalOceanService struct {
	C *godo.Client
}

type DropletCounter struct {
	status string
	region string
	size   string
}

type FlipCounter struct {
	status string
	region string
}

type VolumeCounter struct {
	status string
	region string
	size   string
}

var pageOpt = &godo.ListOptions{
	Page:    1,
	PerPage: 200,
}

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
