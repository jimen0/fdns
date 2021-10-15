package fdns

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/klauspost/pgzip"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		domain  string
		records []string
		workers int
		exp     []string
	}{
		{
			name:    "cname 1 worker",
			input:   `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}`,
			domain:  "foo.tld",
			records: []string{"cname"},
			workers: 1,
			exp:     []string{"reseauocoz.cluster007.foo.tld"},
		},
		{
			name: "cname 2 workers",
			input: `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster008.foo.tld", "type": "cname", "value": "cluster008.bar.tld"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster009.foo.tld", "type": "cname", "value": "cluster009.bar.tld"}`,
			domain:  "foo.tld",
			records: []string{"cname"},
			workers: 2,
			exp:     []string{"reseauocoz.cluster007.foo.tld", "reseauocoz.cluster008.foo.tld", "reseauocoz.cluster009.foo.tld"},
		},
		{
			name:    "a 1 worker",
			input:   `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "a", "value": "127.0.0.1"}`,
			domain:  "foo.tld",
			records: []string{"a"},
			workers: 1,
			exp:     []string{"reseauocoz.cluster007.foo.tld"},
		},
		{
			name: "a 3 workers 2 entries",
			input: `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "a", "value": "127.0.0.1"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster008.foo.tld", "type": "a", "value": "127.0.0.2"}`,
			domain:  "foo.tld",
			records: []string{"a"},
			workers: 3,
			exp:     []string{"reseauocoz.cluster007.foo.tld", "reseauocoz.cluster008.foo.tld"},
		},
		{
			name:    "ns 1 worker",
			input:   `{"timestamp": "1492468299","name": "ns.cdn.foo.tld", "type": "ns", "value": "ns.internal.foo.tld"}`,
			domain:  "foo.tld",
			records: []string{"ns"},
			workers: 1,
			exp:     []string{"ns.cdn.foo.tld"},
		},
		{
			name:    "aaaa 1 worker",
			input:   `{"timestamp": "1492468299","name": "1087074.ostk.bm2.corp.ne1.foo.tld", "type": "aaaa", "value": "2001:4998:efeb:202::5001"}`,
			domain:  "foo.tld",
			records: []string{"aaaa"},
			workers: 1,
			exp:     []string{"1087074.ostk.bm2.corp.ne1.foo.tld"},
		},
		{
			name:    "invalid JSON",
			input:   `{invalidJSON"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}`,
			domain:  "foo.tld",
			records: []string{"cname"},
			workers: 1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := make(chan string)
			errs := make(chan error)
			done := make(chan struct{})

			p := Parser{
				Domains: []string{tc.domain},
				Workers: tc.workers,
				Records: tc.records,
			}

			var res []string
			go func() {
				for c := range out {
					res = append(res, c)
				}
				done <- struct{}{}
			}()

			go p.Parse(context.Background(), bytes.NewReader(encode(t, tc.input)), out, errs)
			select {
			case err := <-errs:
				if tc.exp != nil {
					t.Fatal(err)
				}
			case <-done:
			}

			if !equal(t, tc.exp, res) {
				t.Fatalf("expected %q got %q", tc.exp, res)
			}
		})
	}
}

func TestIsInterestingRecord(t *testing.T) {
	tt := []struct {
		name string
		rec  string
		exp  bool
	}{
		{
			name: "interesting",
			rec:  "a",
			exp:  true,
		},
		{
			name: "boring",
			rec:  "ns",
		},
	}

	for _, tc := range tt {
		p := &Parser{Records: []string{"a"}}

		got := p.IsInterestingRecord(entry{Type: tc.rec})
		if tc.exp != got {
			t.Fatalf("%s: expected %v got %v", tc.name, tc.exp, got)
		}
	}
}

func TestIsInterestingDomain(t *testing.T) {
	tt := []struct {
		name   string
		domain string
		exp    bool
	}{
		{
			name:   "interesting",
			domain: "bazz.example.com",
			exp:    true,
		},
		{
			name:   "boring",
			domain: "bar.notexample.com",
		},
	}

	for _, tc := range tt {
		p := &Parser{Domains: []string{".example.com"}}

		got := p.IsInterestingDomain(entry{Name: tc.domain})
		if tc.exp != got {
			t.Fatalf("%s: expected %v got %v", tc.name, tc.exp, got)
		}
	}
}

func TestSubstring(t *testing.T) {
	tt := []struct {
		sb     string
		domain string
		exp    bool
	}{
		{
			sb:     "interesting",
			domain: "bazz.interestingexample.com",
			exp:    true,
		},
		{
			sb:     "boring",
			domain: "bar.notexample.com",
		},
	}

	for _, tc := range tt {
		p := &Parser{Substrings: []string{"interesting"}, Records: []string{"a"}}

		got := p.Substring(entry{Name: tc.domain, Type: "A", Value: "0.0.0.0"})
		if tc.exp != got {
			t.Fatalf("expected %v got %v for substring %q and domain %q", tc.exp, got, tc.sb, tc.domain)
		}
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
	gz := pgzip.NewWriter(&b)
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
