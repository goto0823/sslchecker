package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"
)

type result struct {
	host     string
	daysLeft int
	err      error
}

func main() {
	hosts := []string{
		"example.com",
		"google.com",
		"does-not-exist.example",
		"also-does-not-exist.example",
		"yahoo.com",
	}

	var failedFetch bool
	var wg sync.WaitGroup
	now := time.Now()
	results := make([]result, len(hosts))

	for i, h := range hosts {
		wg.Go(func() {
			peerCert, err := fetchCert(h)
			if err != nil {
				results[i] = result{host: h, err: err}
				return
			}

			notAfter := peerCert.NotAfter

			daysLeft := daysUntil(notAfter, now)

			results[i] = result{host: h, daysLeft: daysLeft}
		})
	}

	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, r.err)
			failedFetch = true
			continue
		}
		fmt.Printf("%s 残り%d日\n", r.host, r.daysLeft)
	}

	if failedFetch {
		os.Exit(1)
	}
}

func fetchCert(host string) (*x509.Certificate, error) {
	const defaultPort = "443"
	addr := net.JoinHostPort(host, defaultPort)
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch cert %s: %w", host, err)
	}

	defer conn.Close()

	peerCerts := conn.ConnectionState().PeerCertificates
	if len(peerCerts) == 0 {
		return nil, fmt.Errorf("fetch cert %s: no cert returned", host)
	}

	return peerCerts[0], nil
}

// daysUntil は SSL 証明書の有効期限までの残り日数を返す。
// 端数は安全側（小さめ）に丸める。
//
//	0.5 => 0
//	-0.5 => -1
//
// 負の値はSSL証明書の期限が失効していることを意味している。
func daysUntil(notAfter, now time.Time) int {
	d := notAfter.Sub(now)

	return int(math.Floor(d.Hours() / 24))
}
