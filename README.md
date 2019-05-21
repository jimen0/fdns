# fdns

![Go Report Card](https://goreportcard.com/badge/github.com/jimen0/fdns)
[![Documentation](https://godoc.org/github.com/jimen0/fdns?status.svg)](http://godoc.org/github.com/jimen0/fdns)


Package **fdns** parses [Rapid7](https://www.rapid7.com/) [Forward DNS](https://github.com/rapid7/sonar/wiki/Forward-DNS) dataset in a concurrent way. The parser reports found entries (subdomains, IP addresses, records, etc) for the given record and domain.

Behaviour has changed since the project was created. Now `A` record reports DNS names instead of IP addresses.

## Install

```console
go get -u github.com/jimen0/fdns/cmd/fdns
cd $GOPATH/src/github.com/jimen0/fdns/cmd/fdns
go install
```

## Run with Docker

```console
git clone https://github.com/jimen0/fdns.git
cd fdns
docker build -t fdns .
# Make sure to add your arguments
docker run --rm fdns
```

## Usage

```console
➜  ~ fdns
  -domain string
    	domain of which subdomains are discovered
  -file string
    	path of the dataset (can't be used with url)
  -record string
    	record that will be parsed A|CNAME|NS|PTR
  -t int
    	number of goroutines (default 50)
  -url string
    	URL of the dataset (can't be used with file)
  -verbose
    	enable verbose error messages
```

[![asciicast](https://asciinema.org/a/CCLBwLNuX9YJCuavQnk2bgMXd.svg)](https://asciinema.org/a/CCLBwLNuX9YJCuavQnk2bgMXd?speed=5)

```console
docker run --rm -it fdns -domain yahoo.com -record A -t 4 -url https://opendata.rapid7.com/sonar.fdns_v2/2019-04-27-1556328751-fdns_any.json.gz
```

## Test

Just run `go test -race -v github.com/jimen0/fdns`

## Improvements

Send a PR or open an issue. Just make sure that your PR passes `gofmt`, `golint` and `govet`.
