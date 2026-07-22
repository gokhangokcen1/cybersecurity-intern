package dnscheck

import (
	"strings"
	"time"

	"github.com/miekg/dns"
)

func GetAuthoritativeIPs(host string) ([]string, error) {
	// 1) Domain'in yetkili sunucularını (NS kaydı) öğren
	nsAnswers := exchangeQuery(host, dns.TypeNS)
	var nsHost string
	for _, ans := range nsAnswers {
		if ns, ok := ans.(*dns.NS); ok {
			nsHost = strings.TrimSuffix(ns.Ns, ".")
			break
		}
	}
	if nsHost == "" {
		return nil, nil
	}

	// 2) O nameserver'ın kendi IP adresini bul (ona bağlanabilmek için)
	nsIPAnswers := exchangeQuery(nsHost, dns.TypeA)
	var nsIP string
	for _, ans := range nsIPAnswers {
		if a, ok := ans.(*dns.A); ok {
			nsIP = a.A.String()
			break
		}
	}
	if nsIP == "" {
		return nil, nil
	}

	// 3) Nameserver'a DİREKT sor — RecursionDesired=false diyerek
	// "cache'lenmiş cevap değil, sen zaten yetkilisin, kendi bildiğini ver" diyoruz
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	msg.RecursionDesired = false

	client := new(dns.Client)
	client.Timeout = 3 * time.Second

	resp, _, err := client.Exchange(msg, nsIP+":53")
	if err != nil || resp == nil {
		return nil, err
	}

	var ips []string
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	return ips, nil
}
