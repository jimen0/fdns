package fdns

import (
	"bytes"
	"compress/gzip"
	"context"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		domain  string
		f       ParseFunc
		workers int
		exp     []string
	}{
		{
			name:    "cname 1 worker",
			input:   `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}`,
			domain:  "foo.tld",
			f:       CNAME,
			workers: 1,
			exp:     []string{"reseauocoz.cluster007.foo.tld"},
		},
		{
			name: "cname 2 workers",
			input: `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster008.foo.tld", "type": "cname", "value": "cluster008.bar.tld"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster009.foo.tld", "type": "cname", "value": "cluster009.bar.tld"}`,
			domain:  "foo.tld",
			f:       CNAME,
			workers: 2,
			exp:     []string{"reseauocoz.cluster007.foo.tld", "reseauocoz.cluster008.foo.tld", "reseauocoz.cluster009.foo.tld"},
		},
		{
			name:    "a 1 worker",
			input:   `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "a", "value": "127.0.0.1"}`,
			domain:  "foo.tld",
			f:       A,
			workers: 1,
			exp:     []string{"reseauocoz.cluster007.foo.tld"},
		},
		{
			name: "a 3 workers 2 entries",
			input: `{"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "a", "value": "127.0.0.1"}
			{"timestamp": "1492468299","name": "reseauocoz.cluster008.foo.tld", "type": "a", "value": "127.0.0.2"}`,
			domain:  "foo.tld",
			f:       A,
			workers: 3,
			exp:     []string{"reseauocoz.cluster007.foo.tld", "reseauocoz.cluster008.foo.tld"},
		},
		{
			name:    "ns 1 worker",
			input:   `{"timestamp": "1492468299","name": "ns.cdn.foo.tld", "type": "ns", "value": "ns.internal.foo.tld"}`,
			domain:  "foo.tld",
			f:       NS,
			workers: 1,
			exp:     []string{"ns.cdn.foo.tld"},
		},
		{
			name:    "invalid JSON",
			input:   `{invalidJSON"timestamp": "1492468299","name": "reseauocoz.cluster007.foo.tld", "type": "cname", "value": "cluster007.bar.tld"}`,
			domain:  "foo.tld",
			f:       CNAME,
			workers: 1,
			exp:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := make(chan string)
			errs := make(chan error)
			done := make(chan struct{})

			p := NewParser(tc.domain, tc.workers, tc.f)

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
