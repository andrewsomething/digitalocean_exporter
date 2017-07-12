// Command digitalocean_exporter provides a Prometheus exporter for DigitalOcean.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/andrewsomething/digitalocean_exporter"
	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
)

const (
	version = "0.1-dev"
	agent   = "andrewsomething/digitalocean_exporter"
)

var (
	debug       = flag.Bool("debug", false, "Print debug logs")
	listenAddr  = flag.String("listen", "localhost:9292", "Listen address for DigitalOcean exporter")
	metricsPath = flag.String("metrics-path", "/metrics", "URL path for surfacing metrics")
	apiToken    = flag.String("token", "", "DigitalOcean API token (read-only)")
	versionFlag = flag.Bool("v", false, "Prints current digitalocean_exporter version")
)

// TokenSource holds an OAuth token.
type TokenSource struct {
	AccessToken string
}

// Token returns an OAuth token.
func (t *TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *apiToken == "" {
		log.Fatal("A DigitalOcean API token must be specified with '-token' flag")
	}

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ts := &TokenSource{AccessToken: *apiToken}
	oauthClient := oauth2.NewClient(oauth2.NoContext, ts)
	c := godo.NewClient(oauthClient)
	ua := []string{agent, version}
	c.UserAgent = strings.Join(ua, "/")

	newExporter := digitaloceanexporter.New(&digitaloceanexporter.DigitalOceanService{C: c})
	prometheus.MustRegister(newExporter)

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
