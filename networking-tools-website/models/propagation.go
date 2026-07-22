package models

type GeoInfo struct {
	City    string  `json:"city"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

type DNSCheckResult struct {
	IP       string   `json:"ip"`
	City     string   `json:"city"`
	Country  string   `json:"country"`
	Lat      float64  `json:"lat"`
	Lon      float64  `json:"lon"`
	Resolved bool     `json:"resolved"`
	Type     string   `json:"type"`              // A, AAAA, MX vs.
	IPs      []string `json:"ips,omitempty"`     // Geriye dönük uyumluluk
	Records  []string `json:"records,omitempty"` // Tüm sonuç dizisi
}
