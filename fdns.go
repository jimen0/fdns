package fdns

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrWrongType is the error returned when the parsed entry contains an invalid
// record type.
var ErrWrongType = errors.New("incorrect record type")

// ParseFunc defines how a parsing function must work.
type ParseFunc func(e entry) (string, error)

// A reports DNS A records for the given domain.
func A(e entry) (string, error) {
	if e.Type != "a" {
		return "", ErrWrongType
	}
	return e.Name, nil
}

// CNAME reports DNS CNAME records for the given domain.
func CNAME(e entry) (string, error) {
	if e.Type != "cname" {
		return "", ErrWrongType
	}
	return e.Name, nil
}

// NS reports DNS NS records for the given domain.
func NS(e entry) (string, error) {
	if e.Type != "ns" {
		return "", ErrWrongType
	}
	return e.Name, nil
}

// PTR reports DNS PTR records for the given domain.
func PTR(e entry) (string, error) {
	if e.Type != "ptr" {
		return "", ErrWrongType
	}
	return e.Name, nil
}

// Parser object allows parsing datasets looking for records related with a domain.
type Parser struct {
	domain string
	// parse defines how the parser looks for results.
	parse ParseFunc
	// workers is the numer of simultaneous goroutines the parser will use.
	workers int
}

// Parse reads from the given io.Reader and reports results and errors.
func (p *Parser) Parse(ctx context.Context, r io.Reader, out chan<- string, errs chan<- error) {
	defer close(out)

	gz, err := gzip.NewReader(r)
	if err != nil {
		errs <- err
		return
	}
	defer gz.Close()

	lines := make(chan []byte)
	done := make(chan struct{})
	finished := make(chan struct{}, p.workers)

	domain := fmt.Sprintf(".%s", p.domain)
	for i := 0; i < p.workers; i++ {
		go func() {
			var e entry

			for {
				select {
				case <-done:
					finished <- struct{}{}
					return
				case v := <-lines:
					if err := json.Unmarshal(v, &e); err != nil {
						errs <- fmt.Errorf("could not decode JSON object: %v", err)
						continue
					}

					if !strings.HasSuffix(e.Name, domain) {
						continue
					}

					rec, err := p.parse(e)
					if err == ErrWrongType {
						// it's not the interesting record type.
						continue
					}
					if err != nil {
						errs <- fmt.Errorf("could not parse object: %v", err)
						continue
					}
					out <- rec
				}
			}
		}()
	}

	sc := bufio.NewScanner(gz)
	var current int
	for sc.Scan() {
		select {
		case <-ctx.Done():
			break
		default: // avoid blocking.
		}

		lines <- append([]byte{}, sc.Bytes()...)
		current++
	}

	if err := sc.Err(); err != nil {
		errs <- fmt.Errorf("could not scan: %v", err)
		return
	}
	close(done)

	for i := 0; i < p.workers; i++ {
		<-finished
	}
}

// NewParser returns a FDNS parser that reports entries for the given record.
func NewParser(domain string, workers int, f ParseFunc) *Parser {
	return &Parser{domain: domain, parse: f, workers: workers}
}
