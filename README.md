# fdns

![Go Report Card](https://goreportcard.com/badge/github.com/jimen0/fdns)
[![Documentation](https://godoc.org/github.com/jimen0/fdns?status.svg)](http://godoc.org/github.com/jimen0/fdns)


Package **fdns** parses [Rapid7](https://www.rapid7.com/) [Forward DNS](https://github.com/rapid7/sonar/wiki/Forward-DNS) dataset in a concurrent way. The parser reports found entries (subdomains, IP addresses, records, etc) for the given record and domain.

## Install

```sh
go get -u github.com/jimen0/fdns
cd $GOPATH/src/github.com/jimen0/fdns
go build cmd/fdns.go
```

## Usage

```
➜  cmd git:(master) ✗ ./fdns
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

[![asciicast](https://asciinema.org/a/lE3p8BLDcCOk5uOaRRDhbzDVY.png)](https://asciinema.org/a/lE3p8BLDcCOk5uOaRRDhbzDVY)

```
./fdns -t 50 -domain yahoo.com -file $HOME/2018-02-18-1518940801-fdns_any.json.gz -record A
```

## Test

Just run `go test -race -v github.com/jimen0/fdns`

## Improvements

Send a PR or open an issue. Just make sure that your PR passes `gofmt`, `golint` and `govet`.
