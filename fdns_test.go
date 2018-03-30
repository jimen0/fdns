package fdns

import (
	"bytes"
	"compress/gzip"
	"context"
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	tt := []struct {
		name   string
		domain string
		record string
		exp    Parser
	}{
		{
			name:   "cname",
			domain: "foo.tld",
			record: "CNAME",
			exp:    cname{},
		},
		{
			name:   "a",
			domain: "foo.tld",
			record: "A",
			exp:    a{},
		},
		{
			name:   "ns",
			domain: "foo.tld",
			record: "NS",
			exp:    ns{},
		},
		{
			name:   "nil",
			domain: "foo.tld",
			record: "notfound",
			exp:    nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewParser(tc.record)
			if err != nil && tc.exp != nil {
				t.Fatalf("failed to create parser: %v", err)
			}
			if out != tc.exp {
				t.Fatalf("expected %v got %v", tc.exp, out)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tt := []struct {
		name    string
		parser  Parser
		ctx     context.Context
		domain  string
		in      string
		workers int
		exp     []string
	}{
		{
			name:    "cname 1 worker",
			parser:  cname{},
			ctx:     context.Background(),
			in:      `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}`,
			domain:  "foo.tld",
			workers: 1,
			exp:     []string{"cluster007.bar.tld"},
		},
		{
			name:   "cname 2 workers",
			parser: cname{},
			ctx:    context.Background(),
			in: `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster008.foo.tld", "type": "cname", "value": "cluster008.bar.tld"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster009.foo.tld", "type": "cname", "value": "cluster009.bar.tld"}`,
			domain:  "foo.tld",
			workers: 2,
			exp:     []string{"cluster007.bar.tld", "cluster008.bar.tld", "cluster009.bar.tld"},
		},
		{
			name:    "a 1 worker",
			parser:  a{},
			ctx:     context.Background(),
			in:      `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "a", "value": "127.0.0.1"}`,
			domain:  "foo.tld",
			workers: 1,
			exp:     []string{"127.0.0.1"},
		},
		{
			name:   "a 3 workers 2 entries",
			parser: a{},
			ctx:    context.Background(),
			in: `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "a", "value": "127.0.0.1"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster008.foo.tld", "type": "a", "value": "127.0.0.2"}`,
			domain:  "foo.tld",
			workers: 3,
			exp:     []string{"127.0.0.1", "127.0.0.2"},
		},
		{
			name:    "ns 1 worker",
			parser:  ns{},
			ctx:     context.Background(),
			in:      `{"timestamp": "1492468299","name": "ns.cdn.foo.tld", "type": "ns", "value": "ns.internal.foo.tld"}`,
			domain:  "foo.tld",
			workers: 1,
			exp:     []string{"ns.internal.foo.tld"},
		},
		{
			name:    "invalid JSON",
			parser:  cname{},
			ctx:     context.Background(),
			in:      `{invalidJSON"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}`,
			domain:  "foo.tld",
			workers: 1,
			exp:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := make(chan string)
			errs := make(chan error)
			done := make(chan struct{})

			go tc.parser.Parse(tc.ctx, bytes.NewReader(encode(t, tc.in)), tc.domain, tc.workers, out, errs)

			var res []string
			go func() {
				for c := range out {
					res = append(res, c)
				}
				done <- struct{}{}
			}()

			select {
			case err := <-errs:
				if tc.exp != nil {
					t.Fatal(err)
				}
			case <-done:
			}

			if !equal(t, tc.exp, res) {
				t.Fatalf("expected %v got %v", tc.exp, res)
			}
		})
	}
}

func equal(t *testing.T, a, b []string) bool {
	t.Helper()
	m1 := make(map[string]struct{}, len(a))
	for _, v := range a {
		m1[v] = struct{}{}
	}

	m2 := make(map[string]struct{}, len(b))
	for _, v := range b {
		m2[v] = struct{}{}
	}

	return reflect.DeepEqual(m1, m2)
}

func encode(t *testing.T, s string) []byte {
	t.Helper()

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}

	if err := gz.Flush(); err != nil {
		t.Fatal(err)
	}

	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}
