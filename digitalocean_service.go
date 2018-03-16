package digitaloceanexporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/digitalocean/godo"
	"github.com/satori/go.uuid"
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
	status        string
	region        string
	size          string
	price_hourly  float64
	price_monthly float64
	tags          string
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
	refreshID       uuid.UUID

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
		b.logSearchRequest("Droplets", pageOpt, len(droplets), err)

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
	b.logLastError(err)

	for _, d := range droplets {
		c := DropletCounter{
			d.Status,
			d.Region.Slug,
			d.Size.Slug,
			d.Size.PriceHourly,
			d.Size.PriceMonthly,
			strings.Join(d.Tags, ","),
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
		b.logSearchRequest("FloatingIPs", pageOpt, len(fips), err)

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
	b.logLastError(err)

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
		b.logSearchRequest("LoadBalancers", pageOpt, len(lbs), err)

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
	b.logLastError(err)

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
		b.logSearchRequest("Tags", pageOpt, len(tags), err)

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
	b.logLastError(err)

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
		b.logSearchRequest("Volumes", volumeParams.ListOptions, len(volumes), err)

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
	b.logLastError(err)

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

func (b *DigitalOceanBuffer) refresh() {
	b.refreshID, _ = uuid.NewV4()
	log := logrus.WithField("refreshID", b.refreshID)

	log.Infoln("Starting DigitalOcean data refresh")
	startedAt := time.Now()

	b.prepareDroplets()
	b.prepareFloatingIPs()
	b.prepareLoadBalancers()
	b.prepareTags()
	b.prepareVolumes()

	defer func() {
		duration := time.Now().Sub(startedAt)
		log.WithField("duration", duration.String()).Infoln("Finished DigitalOcean data refresh")
	}()
}

func (b *DigitalOceanBuffer) watch() {
	b.refresh()
	for {
		select {
		case <-time.After(b.refreshInterval):
			b.refresh()
		}
	}
}

func (b *DigitalOceanBuffer) logSearchRequest(resource string, pageOpt *godo.ListOptions, elementsCount int, err error) {
	log := logrus.WithFields(logrus.Fields{
		"refreshID": b.refreshID,
		"page":      pageOpt.Page,
		"perPage":   pageOpt.PerPage,
		"found":     elementsCount,
	})

	message := fmt.Sprintf("Looking for %s", resource)
	if err == nil {
		log.Debugln(message)
	} else {
		log.WithField("error", err).Warningln(message)
	}
}

func (b *DigitalOceanBuffer) logLastError(err error) {
	if err != nil {
		logrus.WithField("refreshID", b.refreshID).WithError(err).Errorln("Error while requesting DigitalOcean")
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
