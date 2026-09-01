package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"time"
)

func main() {

	domain := "exampleeeeeeeeee.com"
	port := "443"

	host := net.JoinHostPort(domain, port)

	peerCert, err := fetchCert(host)

	if err != nil {
		fmt.Println(err)
		return
	}

	notAfter := peerCert.NotAfter

	now := time.Now()

	notAfterDuration := notAfter.Sub(now)

	notAfterDaysLeft := int(notAfterDuration.Hours() / 24)

	fmt.Printf("%s 残り%d日\n", domain, notAfterDaysLeft)

}

func fetchCert(h string) (*x509.Certificate, error) {

	conn, err := tls.Dial("tcp", h, nil)

	if err != nil {
		return nil, fmt.Errorf("fetch cert %s: %w", h, err)
	}

	defer conn.Close()

	peerCerts := conn.ConnectionState().PeerCertificates

	if len(peerCerts) == 0 {
		return nil, errors.New("no certificate returned")
	}

	return peerCerts[0], nil
}
