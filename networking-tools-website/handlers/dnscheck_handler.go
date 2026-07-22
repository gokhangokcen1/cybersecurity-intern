package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/miekg/dns"

	"github.com/gokhangokcen1/subnet-backend/dnscheck"
	"github.com/gokhangokcen1/subnet-backend/models"
)

var resolverIPs = []string{
	// Kuzey Amerika / Global Anycast
	"8.8.8.8:53",        // Google Primary
	"8.8.4.4:53",        // Google Secondary
	"1.1.1.1:53",        // Cloudflare Primary
	"1.0.0.1:53",        // Cloudflare Secondary
	"9.9.9.9:53",        // Quad9
	"208.67.222.222:53", // OpenDNS
	"208.67.220.220:53", // OpenDNS
	"64.6.64.6:53",      // Neustar

	// Avrupa
	"84.200.69.80:53",  // DNS.WATCH (Almanya)
	"213.133.98.98:53", // Hetzner (Almanya)
	"77.88.8.8:53",     // Yandex (Rusya)
	"195.46.39.39:53",  // SafeDNS (İngiltere)

	// Asya / Pasifik
	"223.5.5.5:53",       // AliDNS (Çin)
	"180.76.76.76:53",    // Baidu (Çin)
	"168.126.63.1:53",    // KT (Güney Kore)
	"101.101.101.101:53", // TWNIC (Tayvan)

	// Güney Amerika
	"200.221.11.100:53", // NIC.br (Brezilya)
}

var (
	geoCache   = map[string]models.GeoInfo{}
	geoCacheMu sync.Mutex
)

func geolocateIP(ip string) (models.GeoInfo, error) {
	geoCacheMu.Lock()
	if cached, ok := geoCache[ip]; ok {
		geoCacheMu.Unlock()
		return cached, nil
	}
	geoCacheMu.Unlock()

	resp, err := http.Get("http://ip-api.com/json/" + ip + "?fields=city,country,lat,lon,status")
	if err != nil {
		return models.GeoInfo{}, err
	}
	defer resp.Body.Close()

	var raw struct {
		City    string  `json:"city"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
		Status  string  `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return models.GeoInfo{}, err
	}

	info := models.GeoInfo{City: raw.City, Country: raw.Country, Lat: raw.Lat, Lon: raw.Lon}

	geoCacheMu.Lock()
	geoCache[ip] = info
	geoCacheMu.Unlock()

	return info, nil
}

func stripPort(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// 12 Kayıt Türünün Tamamını Miekg/DNS Sabitleriyle Eşleme
func parseDNSType(recordType string) uint16 {
	switch strings.ToUpper(recordType) {
	case "AAAA":
		return dns.TypeAAAA
	case "CNAME":
		return dns.TypeCNAME
	case "MX":
		return dns.TypeMX
	case "NS":
		return dns.TypeNS
	case "PTR":
		return dns.TypePTR
	case "SRV":
		return dns.TypeSRV
	case "SOA":
		return dns.TypeSOA
	case "TXT":
		return dns.TypeTXT
	case "CAA":
		return dns.TypeCAA
	case "DS":
		return dns.TypeDS
	case "DNSKEY":
		return dns.TypeDNSKEY
	default:
		return dns.TypeA
	}
}

func queryOne(ctx context.Context, domain, addr, recordTypeStr string) models.DNSCheckResult {
	ip := stripPort(addr)
	reqType := strings.ToUpper(recordTypeStr)
	res := models.DNSCheckResult{
		IP:   ip,
		Type: reqType,
	}

	if geo, err := geolocateIP(ip); err == nil {
		res.City, res.Country, res.Lat, res.Lon = geo.City, geo.Country, geo.Lat, geo.Lon
	}

	dnsType := parseDNSType(reqType)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dnsType)

	client := new(dns.Client)
	client.Timeout = 3 * time.Second

	ch := make(chan *dns.Msg, 1)
	go func() {
		resp, _, err := client.Exchange(msg, addr)
		if err != nil {
			ch <- nil
			return
		}
		ch <- resp
	}()

	select {
	case <-ctx.Done():
		return res
	case resp := <-ch:
		if resp == nil || len(resp.Answer) == 0 {
			return res
		}

		for _, ans := range resp.Answer {
			switch r := ans.(type) {
			case *dns.A:
				res.Records = append(res.Records, r.A.String())
				res.IPs = append(res.IPs, r.A.String())
			case *dns.AAAA:
				res.Records = append(res.Records, r.AAAA.String())
				res.IPs = append(res.IPs, r.AAAA.String())
			case *dns.CNAME:
				res.Records = append(res.Records, r.Target)
			case *dns.MX:
				res.Records = append(res.Records, fmt.Sprintf("%d %s", r.Preference, r.Mx))
			case *dns.NS:
				res.Records = append(res.Records, r.Ns)
			case *dns.PTR:
				res.Records = append(res.Records, r.Ptr)
			case *dns.SRV:
				res.Records = append(res.Records, fmt.Sprintf("%d %d %d %s", r.Priority, r.Weight, r.Port, r.Target))
			case *dns.SOA:
				res.Records = append(res.Records, fmt.Sprintf("%s %s %d %d %d %d %d", r.Ns, r.Mbox, r.Serial, r.Refresh, r.Retry, r.Expire, r.Minttl))
			case *dns.TXT:
				res.Records = append(res.Records, strings.Join(r.Txt, " "))
			case *dns.CAA:
				res.Records = append(res.Records, fmt.Sprintf("%d %s \"%s\"", r.Flag, r.Tag, r.Value))
			case *dns.DS:
				res.Records = append(res.Records, fmt.Sprintf("%d %d %d %s", r.KeyTag, r.Algorithm, r.DigestType, r.Digest))
			case *dns.DNSKEY:
				res.Records = append(res.Records, fmt.Sprintf("%d %d %d %s", r.Flags, r.Protocol, r.Algorithm, r.PublicKey))
			}
		}

		res.Resolved = len(res.Records) > 0 || len(res.IPs) > 0
		return res
	}
}

func isValidDomain(d string) bool {
	if len(d) == 0 || len(d) > 253 {
		return false
	}
	for _, c := range d {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '-') {
			return false
		}
	}
	return true
}

func CheckDNSHandler(c fiber.Ctx) error {
	domain := c.Query("domain")
	recordType := c.Query("type", "A")

	if !isValidDomain(domain) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz domain"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	results := make([]models.DNSCheckResult, len(resolverIPs))

	for i, addr := range resolverIPs {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, addr string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = queryOne(ctx, domain, addr, recordType)
		}(i, addr)
	}

	wg.Wait()
	return c.JSON(results)
}

func GetFullRecordsHandler(c fiber.Ctx) error {
	domain := c.Query("domain")
	if !isValidDomain(domain) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz domain"})
	}

	result := dnscheck.CheckAllRecords(domain)
	return c.JSON(result)
}
