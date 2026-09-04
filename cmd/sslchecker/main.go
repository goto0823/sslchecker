package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"os"
	"time"
)

func main() {
	domain := "example.com"
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

// daysUntilはssl期限の残り日を返す。
// 端数は安全に少なく倒すようにしている。
//
//	0.5 => 0
//	-0.5 => -1
//
// 負の値はすでに証明書の期限が失効を意味している。
func daysUntil(notAfter, now time.Time) int {
	d := notAfter.Sub(now)

	return int(math.Floor(d.Hours() / 24))
}
