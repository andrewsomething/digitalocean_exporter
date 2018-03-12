FROM golang:1.8-alpine3.6

EXPOSE 9292

WORKDIR /go/src/github.com/andrewsomething/digitalocean_exporter

RUN apk add -U ca-certificates

COPY . .
RUN apk add --no-cache git && \
    go get -u -v github.com/golang/dep/cmd/dep && \
    dep ensure && \
    go get -v ./cmd/digitalocean_exporter && \
    which digitalocean_exporter && \
    go clean -v -i github.com/golang/dep/cmd/dep... && \
    apk del git

ENTRYPOINT ["/go/bin/digitalocean_exporter", "-listen", "0.0.0.0:9292"]
