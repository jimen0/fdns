package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jimen0/fdns"
)

func main() {
	routines := flag.Int("t", 50, "number of goroutines")
	file := flag.String("file", "", "path of the dataset (can't be used with url)")
	url := flag.String("url", "", "URL of the dataset (can't be used with file)")
	record := flag.String("record", "", "record that will be parsed A|CNAME|NS|PTR")
	domain := flag.String("domain", "", "domain of which subdomains are discovered")

	flag.Parse()

	if (*file != "" && *url != "") || (*file == "" && *url == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c, os.Interrupt)

	go func() {
		for {
			select {
			case <-c:
				log.Println("Received SIGINT")
				cancel()
			case <-ctx.Done():
				return
			}
		}
	}()

	var r io.Reader
	if *url != "" {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   15 * time.Second,
				ResponseHeaderTimeout: 16 * time.Second,
				MaxIdleConnsPerHost:   15,
				DisableKeepAlives:     false,
				MaxIdleConns:          100,
				IdleConnTimeout:       16 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		req, err := http.NewRequest("GET", *url, nil)
		if err != nil {
			log.Fatalf("could not create request: %v", err)
		}
		req = req.WithContext(ctx)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("could not request: %v", err)
		}
		defer resp.Body.Close()
		r = resp.Body
	} else {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatalf("could not open file %s: %v", *file, err)
		}
		r = f
	}

	parser, err := fdns.NewParser(*record)
	if err != nil {
		log.Fatal(err)
	}

	out := make(chan string)
	errs := make(chan error)
	done := make(chan struct{})

	go parser.Parse(ctx, r, *domain, *routines, out, errs)
	go func() {
		for c := range out {
			fmt.Println(c)
		}
		done <- struct{}{}
	}()

	select {
	case err := <-errs:
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	case <-done:
	}
}
