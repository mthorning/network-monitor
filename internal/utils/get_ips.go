package utils

import (
	"errors"
	"strings"
)

func GetIps(ipsString string) ([]string, error) {
	ips := strings.Split(ipsString, ",")
	var processedIps []string

	for i := range ips {
		ip := strings.TrimSpace(ips[i])

		if ip != "" {
			processedIps = append(processedIps, ip)
		}
	}

	if len(processedIps) == 0 {
		return nil, errors.New("No valid IP addresses supplied")
	}

	return processedIps, nil
}
