package digitaloceanexporter

import (
	"context"
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
	droplets, err := listDroplets(s)

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

func listDroplets(s *DigitalOceanService) ([]godo.Droplet, error) {
	ctx := context.TODO()
	dropletList := []godo.Droplet{}

	for {
		droplets, resp, err := s.C.Droplets.List(ctx, pageOpt)

		if err != nil {
			return nil, err
		}

		for _, d := range droplets {
			dropletList = append(dropletList, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		pageOpt.Page = page + 1
	}

	return dropletList, nil
}

// FloatingIPs retrieves a count of Floating IPs grouped by status and region.
func (s *DigitalOceanService) FloatingIPs() (map[FlipCounter]int, error) {
	fips, err := listFips(s)

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

func listFips(s *DigitalOceanService) ([]godo.FloatingIP, error) {
	ctx := context.TODO()
	fipList := []godo.FloatingIP{}

	for {
		fips, resp, err := s.C.FloatingIPs.List(ctx, pageOpt)

		if err != nil {
			return nil, err
		}

		for _, f := range fips {
			fipList = append(fipList, f)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		pageOpt.Page = page + 1
	}

	return fipList, nil
}

// Volumes retrieves a count of Volumes grouped by status, size, and region.
func (s *DigitalOceanService) Volumes() (map[VolumeCounter]int, error) {
	volumes, err := listVolumes(s)

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

func listVolumes(s *DigitalOceanService) ([]godo.Volume, error) {
	ctx := context.TODO()
	volumeList := []godo.Volume{}
	volumeParams := &godo.ListVolumeParams{
		ListOptions: pageOpt,
	}

	for {
		volumes, resp, err := s.C.Storage.ListVolumes(ctx, volumeParams)

		if err != nil {
			return nil, err
		}

		for _, v := range volumes {
			volumeList = append(volumeList, v)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		volumeParams.ListOptions.Page = page + 1
	}

	return volumeList, nil
}
