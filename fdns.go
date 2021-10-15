package fdns

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/klauspost/pgzip"
)

// Parser object allows parsing datasets looking for records related with a domain.
type Parser struct {
	Domains []string
	// Records stores reported records.
	Records []string
	// Substrings stores the reported substrings.
	Substrings []string
	// Workers is the numer of simultaneous goroutines the parser will use.
	Workers int
}

// Parse reads from the given io.Reader and reports results and errors.
func (p *Parser) Parse(ctx context.Context, r io.Reader, out chan<- string, errs chan<- error) {
	defer close(out)

	gz, err := pgzip.NewReader(r)
	if err != nil {
		errs <- err
		return
	}
	defer gz.Close()

	lines := make(chan []byte)
	done := make(chan struct{})
	finished := make(chan struct{}, p.Workers)

	for i := 0; i < p.Workers; i++ {
		go func() {
			var e entry

			for {
				select {
				case <-done:
					finished <- struct{}{}
					return
				case v := <-lines:
					if err := json.Unmarshal(v, &e); err != nil {
						errs <- fmt.Errorf("could not decode JSON object: %w", err)
						continue
					}

					if p.Substring(e) || (p.IsInterestingDomain(e) && p.IsInterestingRecord(e)) {
						out <- e.Name
					}
				}
			}
		}()
	}

	sc := bufio.NewScanner(gz)
	for sc.Scan() {
		select {
		case <-ctx.Done():
			goto ctxDone
		default: // avoid blocking.
		}

		lines <- append([]byte{}, sc.Bytes()...)
	}
	if err := sc.Err(); err != nil {
		errs <- fmt.Errorf("could not scan: %w", err)
		return
	}

ctxDone:
	close(done)
	for i := 0; i < p.Workers; i++ {
		<-finished
	}
}

// IsInterestingRecord reports if the entry contains an interesting record.
func (p *Parser) IsInterestingRecord(e entry) bool {
	for _, r := range p.Records {
		if r == e.Type {
			return true
		}
	}
	return false
}

// IsInterestingDomain reports if the domain is a subdomain of the provided domains.
func (p *Parser) IsInterestingDomain(e entry) bool {
	for _, d := range p.Domains {
		if strings.HasSuffix(e.Name, d) {
			return true
		}
	}
	return false
}

// Substring reports if the domain contains an interesting substring.
func (p *Parser) Substring(e entry) bool {
	for _, sb := range p.Substrings {
		if strings.Contains(e.Name, sb) {
			return true
		}
	}
	return false
}
