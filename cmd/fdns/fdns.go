package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/spf13/pflag"

	"github.com/jimen0/fdns/v2"
)

func main() {
	goroutines := flag.Int("goroutines", 4, "number of goroutines")
	file := flag.String("file", "", "path of the dataset (can't be used with url)")
	url := flag.String("url", "", "URL of the dataset (can't be used with file)")
	records := flag.StringSlice("records", []string{}, "records that will be parsed a|aaaa|cname|ns|ptr")
	domains := flag.StringSlice("domains", []string{}, "domains of which subdomains are discovered")
	substrings := flag.StringSlice("substrings", []string{}, "substrings to match (ignores record types)")
	verbose := flag.Bool("verbose", false, "enable verbose error messages")

	flag.Parse()

	if (*file != "" && *url != "") || (*file == "" && *url == "") {
		flag.Usage()
		return
	}

	var l logger = silentLogger{}
	if *verbose {
		l = verboseLogger{}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var r io.ReadCloser
	if *url != "" {
		client := http.DefaultClient

		req, err := http.NewRequest(http.MethodGet, *url, nil)
		if err != nil {
			log.Fatalf("could not create request: %v", err)
		}
		req = req.WithContext(ctx)
		req.Header.Set("user-agent", "fdns github.com/jimen0/fdns")

		// Remove the bodyclose's linter false positive here.
		//
		//nolint:bodyclose
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("could not request %q: %v", req.URL.String(), err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("got status code %d from %q but expected %d", resp.StatusCode, req.URL.String(), http.StatusOK)
		}

		r = resp.Body
	} else {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatalf("could not open file %s: %v", *file, err)
		}
		r = f
	}
	defer r.Close()

	parser := &fdns.Parser{
		Domains:    *domains,
		Workers:    *goroutines,
		Records:    *records,
		Substrings: *substrings,
	}
	out := make(chan string)
	errs := make(chan error)
	done := make(chan struct{})

	go parser.Parse(ctx, r, out, errs)
	go func() {
		for err := range errs {
			l.Logf(os.Stderr, "could not parse: %v", err)
		}
	}()
	go func() {
		for c := range out {
			fmt.Println(c)
		}
		done <- struct{}{}
	}()

	<-done
}

type logger interface {
	Logf(w io.Writer, format string, a ...interface{})
}

type silentLogger struct{}

func (s silentLogger) Logf(w io.Writer, format string, a ...interface{}) {}

type verboseLogger struct{}

func (v verboseLogger) Logf(w io.Writer, format string, a ...interface{}) {
	fmt.Fprintf(w, format, a...)
}
