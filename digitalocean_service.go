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

// LoadBalancerCounter is a struct holding information about a Load Balancer.
type LoadBalancerCounter struct {
	status string
	region string
}

// TagCounter is a struct holding information about a Tag.
type TagCounter struct {
	name         string
	resourceType string
}

// VolumeCounter is a struct holding information about a Block Storage Volume.
type VolumeCounter struct {
	status string
	region string
	size   string
}

func newPageOpt() *godo.ListOptions {
	return &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}
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
	pageOpt := newPageOpt()

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
	pageOpt := newPageOpt()

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

// LoadBalancers retrieves a count of Load Balancers grouped by status and region.
func (s *DigitalOceanService) LoadBalancers() (map[LoadBalancerCounter]int, error) {
	lbs, err := listLoadBalancers(s)

	counters := make(map[LoadBalancerCounter]int)

	for _, lb := range lbs {
		c := LoadBalancerCounter{
			lb.Status,
			lb.Region.Slug,
		}
		counters[c]++
	}

	return counters, err
}

func listLoadBalancers(s *DigitalOceanService) ([]godo.LoadBalancer, error) {
	ctx := context.TODO()
	lbList := []godo.LoadBalancer{}
	pageOpt := newPageOpt()

	for {
		lbs, resp, err := s.C.LoadBalancers.List(ctx, pageOpt)

		if err != nil {
			return nil, err
		}

		for _, lb := range lbs {
			lbList = append(lbList, lb)
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

	return lbList, nil
}

// Tags retrieves a count of Tags grouped by name and resource type.
func (s *DigitalOceanService) Tags() (map[TagCounter]int, error) {
	tags, err := listTags(s)

	counters := make(map[TagCounter]int)

	for _, t := range tags {
		// Note: Currently only Droplets may be tagged.
		// reflect.ValueOf(t.Resources).Elem().Type().Field(0).Name
		c := TagCounter{
			t.Name,
			"droplets",
		}
		counters[c] = counters[c] + t.Resources.Droplets.Count
	}

	return counters, err
}

func listTags(s *DigitalOceanService) ([]godo.Tag, error) {
	ctx := context.TODO()
	tagList := []godo.Tag{}
	pageOpt := newPageOpt()

	for {
		tags, resp, err := s.C.Tags.List(ctx, pageOpt)

		if err != nil {
			return nil, err
		}

		for _, t := range tags {
			tagList = append(tagList, t)
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

	return tagList, nil
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
		ListOptions: newPageOpt(),
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
