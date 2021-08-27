package provider

type Provider interface {
	GetDomainForHostname(hostname string) (string, string, error)

	GetARecords(hostname, domain string) ([]Record, error)
	GetAAAARecords(hostname, domain string) ([]Record, error)

	CreateARecord(hostname, domain, ip string, ttl int) (Record, error)
	CreateAAAARecord(hostname, domain, ip string, ttl int) (Record, error)

	UpdateARecord(record Record, hostname, domain, ip string, ttl int) (Record, error)
	UpdateAAAARecord(record Record, hostname, domain, ip string, ttl int) (Record, error)
}

type Record interface {
	Id() int
	Ip() string
	Hostname() string
	Ttl() int
}
