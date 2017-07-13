package digitaloceanexporter

import (
	"context"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/digitalocean/godo"
)

const (
	DefaultRefreshInterval int = 60
)

// DigitalOceanService is a wrapper around godo.Client.
type DigitalOceanService struct {
	Buffer *DigitalOceanBuffer
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
func (s *DigitalOceanService) Droplets() map[DropletCounter]int {
	return s.Buffer.Droplets
}

// FloatingIPs retrieves a count of Floating IPs grouped by status and region.
func (s *DigitalOceanService) FloatingIPs() map[FlipCounter]int {
	return s.Buffer.FloatingIPs
}

// LoadBalancers retrieves a count of Load Balancers grouped by status and region.
func (s *DigitalOceanService) LoadBalancers() map[LoadBalancerCounter]int {
	return s.Buffer.LoadBalancers
}

// Tags retrieves a count of Tags grouped by name and resource type.
func (s *DigitalOceanService) Tags() map[TagCounter]int {
	return s.Buffer.Tags
}

// Volumes retrieves a count of Volumes grouped by status, size, and region.
func (s *DigitalOceanService) Volumes() map[VolumeCounter]int {
	return s.Buffer.Volumes
}

func NewDigitalOceanService(buffer *DigitalOceanBuffer) *DigitalOceanService {
	return &DigitalOceanService{
		Buffer: buffer,
	}
}

type DigitalOceanBuffer struct {
	client          *godo.Client
	refreshInterval time.Duration

	Droplets      map[DropletCounter]int
	FloatingIPs   map[FlipCounter]int
	LoadBalancers map[LoadBalancerCounter]int
	Tags          map[TagCounter]int
	Volumes       map[VolumeCounter]int
}

func (b *DigitalOceanBuffer) listDroplets() ([]godo.Droplet, error) {
	ctx := context.TODO()
	dropletList := []godo.Droplet{}
	pageOpt := newPageOpt()

	for {
		droplets, resp, err := b.client.Droplets.List(ctx, pageOpt)
		logSearchRequest("Droplets", pageOpt, len(droplets), err)

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

func (b *DigitalOceanBuffer) prepareDroplets() {
	counters := make(map[DropletCounter]int)

	droplets, err := b.listDroplets()
	logLastError(err)

	for _, d := range droplets {
		c := DropletCounter{
			d.Status,
			d.Region.Slug,
			d.Size.Slug,
		}
		counters[c]++
	}

	b.Droplets = counters
}

func (b *DigitalOceanBuffer) listFips() ([]godo.FloatingIP, error) {
	ctx := context.TODO()
	fipList := []godo.FloatingIP{}
	pageOpt := newPageOpt()

	for {
		fips, resp, err := b.client.FloatingIPs.List(ctx, pageOpt)
		logSearchRequest("FloatingIPs", pageOpt, len(fips), err)

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

func (b *DigitalOceanBuffer) prepareFloatingIPs() {
	counters := make(map[FlipCounter]int)

	floatingIPs, err := b.listFips()
	logLastError(err)

	for _, fip := range floatingIPs {
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

	b.FloatingIPs = counters
}

func (b *DigitalOceanBuffer) listLoadBalancers() ([]godo.LoadBalancer, error) {
	ctx := context.TODO()
	lbList := []godo.LoadBalancer{}
	pageOpt := newPageOpt()

	for {
		lbs, resp, err := b.client.LoadBalancers.List(ctx, pageOpt)
		logSearchRequest("LoadBalancers", pageOpt, len(lbs), err)

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

func (b *DigitalOceanBuffer) prepareLoadBalancers() {
	counters := make(map[LoadBalancerCounter]int)

	loadBallancers, err := b.listLoadBalancers()
	logLastError(err)

	for _, lb := range loadBallancers {
		c := LoadBalancerCounter{
			lb.Status,
			lb.Region.Slug,
		}
		counters[c]++
	}

	b.LoadBalancers = counters
}

func (b *DigitalOceanBuffer) listTags() ([]godo.Tag, error) {
	ctx := context.TODO()
	tagList := []godo.Tag{}
	pageOpt := newPageOpt()

	for {
		tags, resp, err := b.client.Tags.List(ctx, pageOpt)
		logSearchRequest("Tags", pageOpt, len(tags), err)

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

func (b *DigitalOceanBuffer) prepareTags() {
	counters := make(map[TagCounter]int)

	tags, err := b.listTags()
	logLastError(err)

	for _, t := range tags {
		// Note: Currently only Droplets may be tagged.
		// reflect.ValueOf(t.Resources).Elem().Type().Field(0).Name
		c := TagCounter{
			t.Name,
			"droplets",
		}
		counters[c] = counters[c] + t.Resources.Droplets.Count
	}

	b.Tags = counters
}

func (b *DigitalOceanBuffer) listVolumes() ([]godo.Volume, error) {
	ctx := context.TODO()
	volumeList := []godo.Volume{}
	volumeParams := &godo.ListVolumeParams{
		ListOptions: newPageOpt(),
	}

	for {
		volumes, resp, err := b.client.Storage.ListVolumes(ctx, volumeParams)
		logSearchRequest("Volumes", volumeParams.ListOptions, len(volumes), err)

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

func (b *DigitalOceanBuffer) prepareVolumes() {
	counters := make(map[VolumeCounter]int)

	volumes, err := b.listVolumes()
	logLastError(err)

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

	b.Volumes = counters
}

func (n *DigitalOceanBuffer) refresh() {
	logrus.Infoln("Refreshing DigitalOcean data")
	startedAt := time.Now()

	n.prepareDroplets()
	n.prepareFloatingIPs()
	n.prepareLoadBalancers()
	n.prepareTags()
	n.prepareVolumes()

	defer func() {
		duration := time.Now().Sub(startedAt)
		logrus.WithFields(logrus.Fields{
			"duration": duration.String(),
		}).Infoln("Finished DigitalOcean data refresh")
	}()
}

func (n *DigitalOceanBuffer) watch() {
	n.refresh()
	for {
		select {
		case <-time.After(n.refreshInterval):
			n.refresh()
		}
	}
}

func NewDigitalOceanBuffer(client *godo.Client, refreshInterval int) *DigitalOceanBuffer {
	interval := time.Duration(refreshInterval) * time.Second
	buffer := &DigitalOceanBuffer{
		client:          client,
		refreshInterval: interval,
	}

	go buffer.watch()

	return buffer
}

func logSearchRequest(resource string, pageOpt *godo.ListOptions, elementsCount int, err error) {
	logrus.Debugf(
		"Looking for %s: page=%d perPage=%d found=%d error=%v",
		resource,
		pageOpt.Page,
		pageOpt.PerPage,
		elementsCount,
		err,
	)
}

func logLastError(err error) {
	if err != nil {
		logrus.WithError(err).Errorln("Error while requesting DigitalOcean")
	}
}
