## fdns

[![pkg.go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/jimen0/fdns/v2)

Package **fdns** parses [Rapid7](https://www.rapid7.com/) [Forward DNS](https://github.com/rapid7/sonar/wiki/Forward-DNS) dataset in a concurrent way. The parser reports found entries (subdomains, IP addresses, records, etc) for the given record and domain.

Behaviour has changed since the project was created. Now `A` record reports DNS names instead of IP addresses.

**This project is deprecated and will be used to experiment with non-reflection APIs. Please, do not rely on it being backwards compatible. Docker images might be deleted at any time.**

### Build from source

```console
go install github.com/mailru/easyjson/...@v0.7.7
git clone --depth 1 https://github.com/jimen0/fdns.git
cd fdns
go generate
go install cmd/fdns/fdns.go
```

### Run with Docker

```console
docker container run --rm -it ghcr.io/jimen0/fdns/fdns:v2.0.2
```

### Usage

```console
➜  fdns
Usage of fdns:
      --domains strings      domains of which subdomains are discovered
      --file string          path of the dataset (can't be used with url)
      --goroutines int       number of goroutines (default 4)
      --records strings      records that will be parsed a|aaaa|cname|ns|ptr
      --substrings strings   substrings to match (ignores record types)
      --url string           URL of the dataset (can't be used with file)
      --verbose              enable verbose error messages
```

<a href="https://asciinema.org/a/QcyHYCj3z13hn34zoshNshO3x?autoplay=1"><img src="https://asciinema.org/a/QcyHYCj3z13hn34zoshNshO3x.png"/></a>


```console
docker container run \
  --rm \
  -it \
  fdns \
    --domains .yahoo.com,.github.com \
    --records a,aaaa,cname \
    --goroutines 4 \
    --url https://opendata.rapid7.com/sonar.fdns_v2/2020-07-24-1595549209-fdns_any.json.gz \
    --verbose
```

### Warnings

Please, remember that for old datasets, Rapid7 requires you to login before downloading them. This tool only aims to be used with the most recent release of them, therefore it doesn't support that logic. Feel free to submit a PR or fork the project if you need it.

### Test

Just run `go test -race -v ./...` inside of the project's root folder.

### Improvements

Submit a PR or open an issue. Just make sure that your PR passes `gofmt`, `golint` and `go vet`.
