package provider

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"net/http"
	"strings"
)

func NewDigitalOceanProvider(token string) Provider {
	p := &doProvider{client: godo.NewFromToken(token)}
	return p
}

type doProvider struct {
	client *godo.Client
}

func (p *doProvider) GetDomainForHostname(hostname string) (string, string, error) {
	domainParts := strings.Split(hostname, ".")

	hostnameParts := make([]string, 0)

	return p.getDomainForHostname(domainParts, hostnameParts)
}

func (p *doProvider) getDomainForHostname(domainParts []string, hostnameParts []string) (string, string, error) {
	if len(domainParts) < 1 {
		return "", "", fmt.Errorf("domain for requested hostname was not found")
	}

	_, resp, err := p.client.Domains.Get(context.TODO(), strings.Join(domainParts, "."))
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return p.getDomainForHostname(domainParts[1:], append(hostnameParts, domainParts[:1]...))
		} else {
			return "", "", err
		}
	}

	return strings.Join(hostnameParts, "."), strings.Join(domainParts, "."), nil
}

func (p *doProvider) GetARecords(hostname, domain string) ([]Record, error) {
	records, _, err := p.client.Domains.RecordsByTypeAndName(context.TODO(), domain, "A", fmt.Sprintf("%s.%s", hostname, domain), nil)
	if err != nil {
		return nil, err
	}

	result := make([]Record, 0)

	for _, record := range records {
		result = append(result, &doRecord{record: record})
	}

	return result, nil
}

func (p *doProvider) GetAAAARecords(hostname, domain string) ([]Record, error) {
	records, _, err := p.client.Domains.RecordsByTypeAndName(context.TODO(), domain, "AAAA", fmt.Sprintf("%s.%s", hostname, domain), nil)
	if err != nil {
		return nil, err
	}

	result := make([]Record, 0)

	for _, record := range records {
		result = append(result, &doRecord{record: record})
	}

	return result, nil
}

func (p *doProvider) CreateARecord(hostname, domain, ip string, ttl int) (Record, error) {
	record, _, err := p.client.Domains.CreateRecord(context.TODO(), domain, &godo.DomainRecordEditRequest{
		Type: "A",
		Name: hostname,
		Data: ip,
		TTL:  ttl,
	})
	if err != nil {
		return nil, err
	}
	return &doRecord{record: *record}, nil
}

func (p *doProvider) CreateAAAARecord(hostname, domain, ip string, ttl int) (Record, error) {
	record, _, err := p.client.Domains.CreateRecord(context.TODO(), domain, &godo.DomainRecordEditRequest{
		Type: "AAAA",
		Name: hostname,
		Data: ip,
		TTL:  ttl,
	})
	if err != nil {
		return nil, err
	}
	return &doRecord{record: *record}, nil
}

func (p *doProvider) UpdateARecord(record Record, hostname, domain, ip string, ttl int) (Record, error) {
	rec, _, err := p.client.Domains.EditRecord(context.TODO(), domain, record.Id(), &godo.DomainRecordEditRequest{
		Type: "A",
		Name: hostname,
		Data: ip,
		TTL:  ttl,
	})
	if err != nil {
		return nil, err
	}
	return &doRecord{record: *rec}, nil
}

func (p *doProvider) UpdateAAAARecord(record Record, hostname, domain, ip string, ttl int) (Record, error) {
	rec, _, err := p.client.Domains.EditRecord(context.TODO(), domain, record.Id(), &godo.DomainRecordEditRequest{
		Type: "AAAA",
		Name: hostname,
		Data: ip,
		TTL:  ttl,
	})
	if err != nil {
		return nil, err
	}
	return &doRecord{record: *rec}, nil
}

type doRecord struct {
	record godo.DomainRecord
}

func (r *doRecord) Id() int {
	return r.record.ID
}

func (r *doRecord) Ip() string {
	return r.record.Data
}

func (r *doRecord) Hostname() string {
	return r.record.Name
}

func (r *doRecord) Ttl() int {
	return r.record.TTL
}
