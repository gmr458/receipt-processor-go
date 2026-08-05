package main

import (
	"net"
	"net/http"
	"strings"
)

func clientIP(r *http.Request, trustedHeader string) (string, error) {
	if trustedHeader != "" {
		val := r.Header.Get(trustedHeader)
		if val != "" {
			parts := strings.Split(val, ",")
			last := strings.TrimSpace(parts[len(parts)-1])
			if ip := net.ParseIP(last); ip != nil {
				return ip.String(), nil
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return host, nil
}
