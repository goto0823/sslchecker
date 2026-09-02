package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {

	domain := "exampleeeeeeeeee.com"
	port := "443"

	host := net.JoinHostPort(domain, port)

	peerCert, err := fetchCert(host)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	notAfter := peerCert.NotAfter
	now := time.Now()

	daysLeft := daysUntil(notAfter, now)

	fmt.Printf("%s 残り%d日\n", domain, daysLeft)

}

func fetchCert(h string) (*x509.Certificate, error) {

	conn, err := tls.Dial("tcp", h, nil)

	if err != nil {
		return nil, fmt.Errorf("fetch cert %s: %w", h, err)
	}

	defer conn.Close()

	peerCerts := conn.ConnectionState().PeerCertificates

	if len(peerCerts) == 0 {
		return nil, fmt.Errorf("fetch cert %s: no cert returned", h)
	}

	return peerCerts[0], nil
}

func daysUntil(notAfter, now time.Time) int {

	d := notAfter.Sub(now)

	return int(d.Hours() / 24)
}
