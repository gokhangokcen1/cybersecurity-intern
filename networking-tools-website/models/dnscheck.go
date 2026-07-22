package models

// Bir domain için toplanan tüm DNS kayıtlarının tam raporu
type DNSFullResult struct {
	Domain string `json:"domain"`

	A     []string `json:"a,omitempty"`
	AAAA  []string `json:"aaaa,omitempty"`
	CNAME string   `json:"cname,omitempty"`
	NS    []string `json:"ns,omitempty"`
	TXT   []string `json:"txt,omitempty"`
	PTR   []string `json:"ptr,omitempty"`

	MX     []MXRecord     `json:"mx,omitempty"`
	SRV    []SRVRecord    `json:"srv,omitempty"`
	SOA    *SOARecord     `json:"soa,omitempty"`
	DNSKEY []DNSKEYRecord `json:"dnskey,omitempty"`
	DS     []DSRecord     `json:"ds,omitempty"`
	CAA    []CAARecord    `json:"caa,omitempty"`
}

type MXRecord struct {
	Preference uint16 `json:"preference"`
	Host       string `json:"host"`
}

type SRVRecord struct {
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target"`
}

type SOARecord struct {
	NS      string `json:"ns"`
	Mbox    string `json:"mbox"`
	Serial  uint32 `json:"serial"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	Minimum uint32 `json:"minimum"`
}

type DNSKEYRecord struct {
	Flags     uint16 `json:"flags"`
	Protocol  uint8  `json:"protocol"`
	Algorithm uint8  `json:"algorithm"`
	PublicKey string `json:"public_key"`
}

type DSRecord struct {
	KeyTag     uint16 `json:"key_tag"`
	Algorithm  uint8  `json:"algorithm"`
	DigestType uint8  `json:"digest_type"`
	Digest     string `json:"digest"`
}

type CAARecord struct {
	Flag  uint8  `json:"flag"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}
