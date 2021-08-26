package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	knownPublicIPv4 net.IP
	knownPublicIPv6 net.IP
)

func main() {
	p := argparse.NewParser("doddns", "Use DitialOcean for Dynamic DNS")

	externalServicesArg := p.StringList("s", "service", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "External services used to detect public IP",
		Default: []string{
			"http://ifconfig.co/",
			"http://ifconfig.me/",
			"http://ifconfig.io/",
		},
	})

	domainNameArg := p.String("d", "domain", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Domain name to set. A record for said name must already be created.",
		Default:  os.Getenv("DODDNS_DOMAIN"),
	})

	apiTokenFileArg := p.String("f", "token-file", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Location of a file where the DigitalOcean API token is stored",
		Default:  os.Getenv("DODDNS_API_TOKEN_FILE"),
	})

	ipCheckIntervalSecondsArg := p.Int("i", "check-interval", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Interval in seconds at which to check public IP",
		Default:  900,
	})

	err := p.Parse(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	externalServices := *externalServicesArg
	domainName := *domainNameArg
	apiTokenFile := *apiTokenFileArg
	checkIntervalSeconds := *ipCheckIntervalSecondsArg

	if domainName == "" {
		log.Fatal("domain name required")
	}

	if apiTokenFile == "" {
		log.Fatal("API token file required")
	}

	for true {
		var ip net.IP
		for _, externalService := range externalServices {
			resp, err := http.Get(externalService)
			if err != nil || resp.StatusCode != 200 {
				log.Printf("Failed to get IP from %s", externalService)
				continue
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			ip, err = parseAndValidateIPAddress(strings.TrimSpace(string(b)))
			if err != nil {
				log.Fatal(err)
			}
			break
		}
		if ip == nil {
			log.Fatal("Could not obtain public IP address from any known service. Failing.")
		}

		// If new IP then update and say so
		if isNewIp(ip) {
			log.Println("Found new public IP:", ip.String())
			storeIp(ip)
		}

		time.Sleep(time.Second * time.Duration(checkIntervalSeconds))
	}
}

func parseAndValidateIPAddress(s string) (net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("bad IP %s", s)
	}

	// TODO: Actually validate if it is good IP

	return ip, nil
}

func isIPv4(ip net.IP) bool {
	return ip.To4() != nil
}

func storeIp(ip net.IP) {
	if isIPv4(ip) {
		knownPublicIPv4 = ip
	} else {
		knownPublicIPv6 = ip
	}
}

func isNewIp(ip net.IP) bool {
	if isIPv4(ip) {
		if knownPublicIPv4.Equal(ip) {
			return false
		} else {
			return true
		}
	} else {
		if knownPublicIPv6.Equal(ip) {
			return false
		} else {
			return true
		}
	}
}
