// Command digitalocean_exporter provides a Prometheus exporter for DigitalOcean.
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/andrewsomething/digitalocean_exporter"
	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"
)

const (
	version = "0.1-dev"
	agent   = "andrewsomething/digitalocean_exporter"
)

var (
	debug           = flag.Bool("debug", false, "Print debug logs")
	listenAddr      = flag.String("listen", "localhost:9292", "Listen address for DigitalOcean exporter")
	metricsPath     = flag.String("metrics-path", "/metrics", "URL path for surfacing metrics")
	apiToken        = flag.String("token", "", "DigitalOcean API token (read-only)")
	refreshInterval = flag.Int("refresh-interval", digitaloceanexporter.DefaultRefreshInterval, "Interval (in seconds) between subsequent requests against DigitalOcean API")
	versionFlag     = flag.Bool("v", false, "Prints current digitalocean_exporter version")
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

type Handler struct {
	metricsHandler    http.Handler
	metricsPathRegexp *regexp.Regexp
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if h.metricsHandler == nil {
		logrus.Fatalln("metricsHandler is not set")
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ID := uuid.NewV4()

	logrus.WithFields(logrus.Fields{
		"requestID":  ID,
		"host":       host,
		"method":     r.Method,
		"requestURI": r.RequestURI,
		"protocol":   r.Proto,
		"userAgent":  r.UserAgent(),
	}).Infoln("Started request")

	startedAt := time.Now()
	defer func() {
		duration := time.Now().Sub(startedAt)
		logrus.WithFields(logrus.Fields{
			"requestID": ID,
			"duration":  duration.String(),
		}).Infoln("Finished request")
	}()

	if h.metricsPathRegexp.MatchString(r.RequestURI) {
		h.metricsHandler.ServeHTTP(rw, r)
	} else {
		rw.WriteHeader(404)
		rw.Write([]byte(`<html>
		<head><title>DigitalOcean Exporter</title></head>
		<body>
		<h1>DigitalOcean Exporter</h1>
		<p><a href='` + *metricsPath + `'>Metrics</a></p>
		</body>
		</html>`))
	}
}

func newHandler(metricsPath string) *Handler {
	return &Handler{
		metricsHandler:    prometheus.Handler(),
		metricsPathRegexp: regexp.MustCompile(fmt.Sprintf("^%s$", metricsPath)),
	}
}

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *apiToken == "" {
		token := os.Getenv("DIGITALOCEAN_TOKEN")
		if token == "" {
			logrus.Fatalln("A DigitalOcean API token must be specified with '-token' flag or with DIGITALOCEAN_TOKEN environment variable")
		}
		*apiToken = token
	}

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ts := &TokenSource{AccessToken: *apiToken}
	oauthClient := oauth2.NewClient(oauth2.NoContext, ts)
	c := godo.NewClient(oauthClient)
	ua := []string{agent, version}
	c.UserAgent = strings.Join(ua, "/")

	digitalOceanBuffer := digitaloceanexporter.NewDigitalOceanBuffer(c, *refreshInterval)
	digitalOceanService := digitaloceanexporter.NewDigitalOceanService(digitalOceanBuffer)
	newExporter := digitaloceanexporter.New(digitalOceanService)
	prometheus.MustRegister(newExporter)

	logrus.Printf("Starting DigitalOcean exporter on %q", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, newHandler(*metricsPath)); err != nil {
		logrus.Fatalf("Cannot start DigitalOcean exporter: %s", err)
	}
}
