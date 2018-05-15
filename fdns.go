package fdns

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/a8m/djson"
	"github.com/klauspost/pgzip"
)

// Parser interface represents a FDNS dataset parser.
type Parser interface {
	// Parse parses the dataset and sends valid records for any subdomain of
	// the domain through the channel.
	Parse(ctx context.Context, r io.Reader, domain string, workers int, out chan<- string, errors chan<- error)
}

// NewParser returns a FDNS parser that reports entries for the given record.
func NewParser(record string) (Parser, error) {
	var p Parser
	switch record {
	case "A":
		p = a{}
	case "CNAME":
		p = cname{}
	case "NS":
		p = ns{}
	case "PTR":
		p = ptr{}
	default:
		return nil, errors.New("unsupported record type")
	}

	return p, nil
}

func parse(ctx context.Context, record string, domain string, workers int, r io.Reader, out chan<- string, errs chan<- error) {
	defer close(out)

	gz, err := pgzip.NewReader(r)
	if err != nil {
		errs <- err
		return
	}
	defer gz.Close()

	var wg sync.WaitGroup
	wg.Add(workers)

	done := make(chan struct{})
	chans := make([]chan []byte, workers)
	for i := 0; i < len(chans); i++ {
		chans[i] = make(chan []byte)
	}

	domain = fmt.Sprintf(".%s", domain)
	for _, ch := range chans {
		go func(c chan []byte) {
			defer wg.Done()

			select {
			case <-done:
				return
			default: // avoid blocking
			}

			for v := range c {
				res, err := djson.DecodeObject(v)
				if err != nil {
					errs <- err
					done <- struct{}{}
					return
				}

				if res["type"].(string) == record {
					if strings.HasSuffix(res["name"].(string), domain) {
						out <- res["value"].(string)
					}
				}
			}
		}(ch)
	}

	sc := bufio.NewScanner(gz)
	var current int
	for sc.Scan() {
		select {
		case <-ctx.Done():
			errs <- ctx.Err()
			done <- struct{}{}
			return
		default: // avoid blocking.
		}

		chans[current%workers] <- sc.Bytes()
		current++
	}

	if err := sc.Err(); err != nil {
		errs <- err
		done <- struct{}{}
		return
	}

	for _, c := range chans {
		close(c)
	}

	wg.Wait()
}

// a is a dataset parser that reports A records.
type a struct{}

// Parse parses the dataset and reports valid A records of any subdomain.
func (rec a) Parse(ctx context.Context, r io.Reader, domain string, workers int, out chan<- string, errors chan<- error) {
	parse(ctx, "a", domain, workers, r, out, errors)
}

// cname is a dataset parser that reports CNAME records.
type cname struct{}

// Parse parses the dataset and reports valid CNAME records of any subdomain.
func (rec cname) Parse(ctx context.Context, r io.Reader, domain string, workers int, out chan<- string, errors chan<- error) {
	parse(ctx, "cname", domain, workers, r, out, errors)
}

// ns is a dataset parser that reports NS records.
type ns struct{}

// Parse parses the dataset and reports valid NS records of any subdomain.
func (rec ns) Parse(ctx context.Context, r io.Reader, domain string, workers int, out chan<- string, errors chan<- error) {
	parse(ctx, "ns", domain, workers, r, out, errors)
}

// ptr is a dataset parser that reports PTR records.
type ptr struct{}

// Parse parses the dataset and reports valid PTR records of any subdomain.
func (rec ptr) Parse(ctx context.Context, r io.Reader, domain string, workers int, out chan<- string, errors chan<- error) {
	parse(ctx, "ptr", domain, workers, r, out, errors)
}
