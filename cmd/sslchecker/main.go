package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
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
	var timeout time.Duration
	var concurrency int
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "接続タイムアウト")
	flag.IntVar(&concurrency, "concurrency", 10, "同時接続数")

	flag.Parse()

	hosts := flag.Args()
	if len(hosts) == 0 {
		fmt.Fprintln(os.Stderr, "ホストを1つ以上入力してください")
		flag.Usage()
		os.Exit(2)
	}

	if concurrency < 1 {
		fmt.Fprintln(os.Stderr, "同時接続数は1以上選択してください")
		flag.Usage()
		os.Exit(2)
	}

	var failedFetch bool
	var wg sync.WaitGroup
	now := time.Now()
	results := make([]result, len(hosts))
	sem := make(chan struct{}, concurrency)

	for i, h := range hosts {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			peerCert, err := fetchCert(h, timeout)
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

func fetchCert(host string, timeout time.Duration) (*x509.Certificate, error) {
	const defaultPort = "443"
	addr := net.JoinHostPort(host, defaultPort)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var d tls.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("fetch cert %s: %w", host, err)
	}

	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	peerCerts := tlsConn.ConnectionState().PeerCertificates
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
