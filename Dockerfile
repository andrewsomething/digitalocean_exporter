FROM golang:1.8-alpine3.6

EXPOSE 9292

WORKDIR /go/src/github.com/andrewsomething/digitalocean_exporter

RUN apk add -U ca-certificates

COPY . .
RUN apk add --no-cache git make; \
    make build; \
    cp digitalocean_exporter /usr/local/bin/; \
    apk del git make

ENTRYPOINT ["/usr/local/bin/digitalocean_exporter", "-listen", "0.0.0.0:9292"]
