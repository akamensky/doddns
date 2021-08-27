package main

import (
	"doddns/provider"
	"doddns/utils"
	"github.com/akamensky/argparse"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	p := argparse.NewParser("doddns", "Use DitialOcean for Dynamic DNS")

	externalServicesArg := p.StringList("s", "service", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "External services used to detect public IP",
		Default: []string{
			"https://ifconfig.co/",
			"https://ifconfig.me/",
			"https://ifconfig.io/",
		},
	})

	hostnameArg := p.String("", "hostname", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Hostname name to set. Corresponding domain name must already be created",
		Default:  os.Getenv("DODDNS_HOSTNAME"),
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
		Help:     "Interval in minutes at which to check public IP",
		Default:  utils.GetEnvDefaultInt("DODDNS_CHECK_INTERVAL_MINUTES", 1),
	})

	ttlArg := p.Int("", "ttl", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "DNS record TTL to use",
		Default:  utils.GetEnvDefaultInt("DODDNS_RECORD_TTL", 300),
	})

	err := p.Parse(os.Args)
	if err != nil {
		log.Fatalln(err)
	}

	externalServices := *externalServicesArg
	hostname := *hostnameArg
	apiTokenFile := *apiTokenFileArg
	checkIntervalSeconds := *ipCheckIntervalSecondsArg
	ttl := *ttlArg

	if hostname == "" {
		log.Fatal("hostname required")
	}

	if apiTokenFile == "" {
		log.Fatalln("API token file required")
	}

	token, err := utils.ReadTokenFromFile(apiTokenFile)
	if err != nil {
		log.Fatalln(err)
	}

	doClient := provider.NewDigitalOceanProvider(token)
	// Get domain for hostname
	hostname, domain, err := doClient.GetDomainForHostname(hostname)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Found domain [%s], hostname [%s]", domain, hostname)

	// Verify only 1 A and 1 AAAA records for hostname
	var aRecord, aaaaRecord provider.Record
	recs, err := doClient.GetARecords(hostname, domain)
	if err != nil {
		log.Fatalln(err)
	}
	switch len(recs) {
	case 0:
		log.Printf("Found 0 A records for hostname [%s] in domain [%s]\n", hostname, domain)
	case 1:
		aRecord = recs[0]
		log.Printf("Found 1 A record for hostname [%s] in domain [%s] pointing to [%s]\n", hostname, domain, aRecord.Ip())
	default:
		log.Fatalf("found multiple A records for hostname [%s] in domain [%s]\n", hostname, domain)
	}

	recs, err = doClient.GetAAAARecords(hostname, domain)
	if err != nil {
		log.Fatalln(err)
	}
	switch len(recs) {
	case 0:
		log.Printf("Found 0 AAAA records for hostname [%s] in domain [%s]\n", hostname, domain)
	case 1:
		aaaaRecord = recs[0]
		log.Printf("Found 1 AAAA record for hostname [%s] in domain [%s] pointing to [%s]\n", hostname, domain, aaaaRecord.Ip())
	default:
		log.Fatalf("found multiple AAAA records for hostname [%s] in domain [%s]\n", hostname, domain)
	}

	for true {
		var ip net.IP
		for _, externalService := range externalServices {
			resp, err := http.Get(externalService)
			if err != nil || resp.StatusCode != 200 {
				log.Printf("Failed to get IP from %s\n", externalService)
				continue
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			ip, err = utils.ParseAndValidateIPAddress(strings.TrimSpace(string(b)))
			if err != nil {
				log.Fatalln(err)
			}
			break
		}
		if ip == nil {
			log.Fatalln("Could not obtain public IP address from any known service. Failing.")
		}

		if utils.IsIPv4(ip) {
			if aRecord == nil {
				aRecord, err = doClient.CreateARecord(hostname, domain, ip.String(), ttl)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("Created new A record [%s.%s] pointing to [%s] with TTL %d", hostname, domain, ip.String(), ttl)
			} else if aRecord.Ip() != ip.String() || aRecord.Ttl() != ttl {
				aRecord, err = doClient.UpdateARecord(aRecord, hostname, domain, ip.String(), ttl)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("Updated A record [%s.%s] pointing to [%s] with TTL %d", hostname, domain, ip.String(), ttl)
			}
		} else {
			if aaaaRecord == nil {
				aaaaRecord, err = doClient.CreateAAAARecord(hostname, domain, ip.String(), ttl)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("Created new AAAA record [%s.%s] pointing to [%s] with TTL %d", hostname, domain, ip.String(), ttl)
			} else if aaaaRecord.Ip() != ip.String() || aaaaRecord.Ttl() != ttl {
				aaaaRecord, err = doClient.UpdateAAAARecord(aaaaRecord, hostname, domain, ip.String(), ttl)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("Updated AAAA record [%s.%s] pointing to [%s] with TTL %d", hostname, domain, ip.String(), ttl)
			}
		}

		time.Sleep(time.Minute * time.Duration(checkIntervalSeconds))
	}
}
