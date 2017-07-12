dependencies:
	go get github.com/digitalocean/godo
	go get github.com/prometheus/client_golang/prometheus
	go get github.com/Sirupsen/logrus
	go get golang.org/x/oauth2

build: CGO_ENABLED := 0
build: dependencies
	cd cmd/digitalocean_exporter && go build -o ../../digitalocean_exporter
