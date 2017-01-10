// Command digitalocean_exporter provides a Prometheus exporter for DigitalOcean.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/andrewsomething/digitalocean_exporter"
	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
)

var (
	listenAddr  = flag.String("listen", "localhost:9292", "Listen address for DigitalOcean exporter")
	metricsPath = flag.String("metrics-path", "/metrics", "URL path for surfacing metrics")
	apiToken    = flag.String("token", "", "DigitalOcean API token (read-only)")
)

// tokenSource holds an OAuth token.
type TokenSource struct {
	AccessToken string
}

// token returns an OAuth token.
func (t *TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

func main() {
	flag.Parse()

	if *apiToken == "" {
		log.Fatal("A DigitalOcean API token must be specified with '-token' flag")
	}

	ts := &TokenSource{AccessToken: *apiToken}
	oauthClient := oauth2.NewClient(oauth2.NoContext, ts)
	c := godo.NewClient(oauthClient)

	prometheus.MustRegister(digitalocean_exporter.New(&digitalocean_exporter.DigitalOceanService{c}))

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>DigitalOcean Exporter</title></head>
		<body>
		<h1>DigitalOcean Exporter</h1>
		<p><a href='` + *metricsPath + `'>Metrics</a></p>
		</body>
		</html>`))
	})

	log.Printf("starting DigitalOcean exporter on %q", *listenAddr)

	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("cannot start DigitalOcean exporter: %s", err)
	}
}
