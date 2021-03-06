package digitaloceanexporter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestDroplets(t *testing.T) {
	var dropletTests = []struct {
		resp     string
		expected map[DropletCounter]int
	}{
		{`{"droplets": [
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc3"}},
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc3"}}]}`,
			map[DropletCounter]int{DropletCounter{status: "active", size: "1gb", region: "nyc3", price_hourly: 0.014880, price_monthly: 5.0}: 2}},
		{`{"droplets": [
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc3"}},
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc3"}},
        {"status":"active", "size":{"slug":"2gb", "price_hourly": 0.029760, "price_monthly": 20.0}, "region":{"slug":"nyc3"}}]}`,
			map[DropletCounter]int{DropletCounter{status: "active", size: "1gb", region: "nyc3", price_hourly: 0.014880, price_monthly: 5.0}: 2,
				DropletCounter{status: "active", size: "2gb", region: "nyc3", price_hourly: 0.029760, price_monthly: 20.0}: 1}},
		{`{"droplets": [
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc3"}},
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc3"}},
        {"status":"active", "size":{"slug":"1gb", "price_hourly": 0.014880, "price_monthly": 5.0}, "region":{"slug":"nyc2"}},
        {"status":"active", "size":{"slug":"2gb", "price_hourly": 0.029760, "price_monthly": 20.0}, "region":{"slug":"nyc3"}}]}`,
			map[DropletCounter]int{DropletCounter{status: "active", size: "1gb", region: "nyc3", price_hourly: 0.014880, price_monthly: 5.0}: 2,
				DropletCounter{status: "active", size: "1gb", region: "nyc2", price_hourly: 0.014880, price_monthly: 5.0}:  1,
				DropletCounter{status: "active", size: "2gb", region: "nyc3", price_hourly: 0.029760, price_monthly: 20.0}: 1}},
	}

	for _, tt := range dropletTests {
		apiServer(t, "/v2/droplets", tt.resp, func() {
			dob := getDOBuffer()
			dob.prepareDroplets()
			dos := NewDigitalOceanService(dob)
			assert.Equal(t, tt.expected, dos.Droplets(), "they should be equal")
		})
	}
}

func TestFloatingIPs(t *testing.T) {
	var fipTests = []struct {
		resp     string
		expected map[FlipCounter]int
	}{
		{`{"floating_ips": [
        {"droplet":{"id": 1}, "region":{"slug":"nyc3"}},
        {"droplet":{"id": 2}, "region":{"slug":"nyc3"}}]}`,
			map[FlipCounter]int{FlipCounter{status: "assigned", region: "nyc3"}: 2}},
		{`{"floating_ips": [
        {"droplet":{"id": 1}, "region":{"slug":"nyc3"}},
        {"droplet":{"id": 2}, "region":{"slug":"nyc3"}},
        {"droplet": null, "region":{"slug":"nyc3"}}]}`,
			map[FlipCounter]int{FlipCounter{status: "assigned", region: "nyc3"}: 2,
				FlipCounter{status: "unassigned", region: "nyc3"}: 1}},
		{`{"floating_ips": [
        {"droplet":{"id": 1}, "region":{"slug":"nyc3"}},
        {"droplet":{"id": 2}, "region":{"slug":"nyc3"}},
        {"droplet": null, "region":{"slug":"nyc3"}},
        {"droplet": null, "region":{"slug":"nyc2"}}]}`,
			map[FlipCounter]int{FlipCounter{status: "assigned", region: "nyc3"}: 2,
				FlipCounter{status: "unassigned", region: "nyc3"}: 1,
				FlipCounter{status: "unassigned", region: "nyc2"}: 1}},
	}

	for _, tt := range fipTests {
		apiServer(t, "/v2/floating_ips", tt.resp, func() {
			dob := getDOBuffer()
			dob.prepareFloatingIPs()
			dos := NewDigitalOceanService(dob)
			assert.Equal(t, tt.expected, dos.FloatingIPs(), "they should be equal")
		})
	}
}

func TestLoadBalancers(t *testing.T) {
	var lbTests = []struct {
		resp     string
		expected map[LoadBalancerCounter]int
	}{
		{`{"load_balancers": [
        {"id": "abc", "region":{"slug":"nyc3"}, "status": "active"},
        {"id": "xyz", "region":{"slug":"nyc3"}, "status": "active"}]}`,
			map[LoadBalancerCounter]int{LoadBalancerCounter{status: "active", region: "nyc3"}: 2}},
		{`{"load_balancers": [
        {"id": "abc", "region":{"slug":"nyc3"}, "status": "active"},
        {"id": "xyz", "region":{"slug":"nyc3"}, "status": "active"},
        {"droplet": null, "region":{"slug":"nyc3"}, "status": "new"}]}`,
			map[LoadBalancerCounter]int{LoadBalancerCounter{status: "active", region: "nyc3"}: 2,
				LoadBalancerCounter{status: "new", region: "nyc3"}: 1}},
		{`{"load_balancers": [
        {"id": "abc", "region":{"slug":"nyc3"}, "status": "active"},
        {"id": "xyz", "region":{"slug":"nyc3"}, "status": "active"},
        {"id": "efg", "region":{"slug":"nyc3"}, "status": "new"},
        {"id": "hijk", "region":{"slug":"nyc2"}, "status": "error"}]}`,
			map[LoadBalancerCounter]int{LoadBalancerCounter{status: "active", region: "nyc3"}: 2,
				LoadBalancerCounter{status: "new", region: "nyc3"}:   1,
				LoadBalancerCounter{status: "error", region: "nyc2"}: 1}},
	}

	for _, tt := range lbTests {
		apiServer(t, "/v2/load_balancers", tt.resp, func() {
			dob := getDOBuffer()
			dob.prepareLoadBalancers()
			dos := NewDigitalOceanService(dob)
			assert.Equal(t, tt.expected, dos.LoadBalancers(), "they should be equal")
		})
	}
}

func TestTags(t *testing.T) {
	var tagTests = []struct {
		resp     string
		expected map[TagCounter]int
	}{
		{`{"tags": [
        {"name": "foo", "resources": {"droplets": {"count": 3}}}]}`,
			map[TagCounter]int{TagCounter{name: "foo", resourceType: "droplets"}: 3}},
		{`{"tags": [
        {"name": "foo", "resources": {"droplets": {"count": 1}}},
        {"name": "bar", "resources": {"droplets": {"count": 2}}}]}`,
			map[TagCounter]int{TagCounter{name: "foo", resourceType: "droplets"}: 1,
				TagCounter{name: "bar", resourceType: "droplets"}: 2}},
	}

	for _, tt := range tagTests {
		apiServer(t, "/v2/tags", tt.resp, func() {
			dob := getDOBuffer()
			dob.prepareTags()
			dos := NewDigitalOceanService(dob)
			assert.Equal(t, tt.expected, dos.Tags(), "they should be equal")
		})
	}
}

func TestVolumes(t *testing.T) {
	var volumeTests = []struct {
		resp     string
		expected map[VolumeCounter]int
	}{
		{`{"volumes": [
        {"droplet_ids":[1], "size_gigabytes":100, "region":{"slug":"nyc3"}},
        {"droplet_ids":[2], "size_gigabytes":100, "region":{"slug":"nyc3"}}]}`,
			map[VolumeCounter]int{VolumeCounter{status: "attached", size: "100", region: "nyc3"}: 2}},
		{`{"volumes": [
        {"droplet_ids":[1], "size_gigabytes":100, "region":{"slug":"nyc3"}},
        {"droplet_ids":[2], "size_gigabytes":100, "region":{"slug":"nyc3"}},
        {"droplet_ids":[], "size_gigabytes":500, "region":{"slug":"nyc3"}}]}`,
			map[VolumeCounter]int{VolumeCounter{status: "attached", size: "100", region: "nyc3"}: 2,
				VolumeCounter{status: "unattached", size: "500", region: "nyc3"}: 1}},
		{`{"volumes": [
        {"droplet_ids":[1], "size_gigabytes":100, "region":{"slug":"nyc3"}},
        {"droplet_ids":[2], "size_gigabytes":100, "region":{"slug":"nyc3"}},
        {"droplet_ids":[], "size_gigabytes":100, "region":{"slug":"nyc2"}},
        {"droplet_ids":[], "size_gigabytes":500, "region":{"slug":"nyc3"}}]}`,
			map[VolumeCounter]int{VolumeCounter{status: "attached", size: "100", region: "nyc3"}: 2,
				VolumeCounter{status: "unattached", size: "100", region: "nyc2"}: 1,
				VolumeCounter{status: "unattached", size: "500", region: "nyc3"}: 1}},
	}

	for _, tt := range volumeTests {
		apiServer(t, "/v2/volumes", tt.resp, func() {
			dob := getDOBuffer()
			dob.prepareVolumes()
			dos := NewDigitalOceanService(dob)
			assert.Equal(t, tt.expected, dos.Volumes(), "they should be equal")
		})
	}
}

var GodoBase *url.URL

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

func getDOBuffer() *DigitalOceanBuffer {
	ts := &TokenSource{AccessToken: "fake-testing-token"}
	oauthClient := oauth2.NewClient(oauth2.NoContext, ts)
	c := godo.NewClient(oauthClient)
	c.BaseURL = GodoBase

	dob := &DigitalOceanBuffer{
		client: c,
	}

	return dob
}

func apiServer(t testing.TB, path string, resp string, test func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("Wrong URL: %v", r.URL.String())
			return
		}
		w.WriteHeader(200)
		fmt.Fprintln(w, resp)
	}))

	u, err := url.Parse(server.URL)
	if err != nil {
		panic(err)
	}
	GodoBase = u

	defer server.Close()
	test()
}
