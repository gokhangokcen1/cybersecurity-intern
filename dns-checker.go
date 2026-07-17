package main

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

const dnsServer = "8.8.8.8:53"

func main() {
	host := "arjeta.com.tr"

	fmt.Println("Domain:", host)

	// A
	fmt.Println("\nA Records:")
	ips, _ := net.LookupIP(host)

	hasIPv4 := false
	for _, ip := range ips {
		if ip.To4() != nil {
			fmt.Println(ip)
			hasIPv4 = true
		}
	}
	if !hasIPv4 {
		fmt.Println("Bulunamadı.")
	}

	// AAAA
	fmt.Println("\nAAAA Records:")
	hasIPv6 := false
	for _, ip := range ips {
		if ip.To4() == nil {
			fmt.Println(ip)
			hasIPv6 = true
		}
	}
	if !hasIPv6 {
		fmt.Println("Bulunamadı.")
	}

	// CNAME
	fmt.Println("\nCNAME Record:")
	cname, err := net.LookupCNAME(host)
	if err != nil {
		fmt.Println("Bulunamadı.")
	} else {
		fmt.Println(cname)
	}

	// MX
	fmt.Println("\nMX Records:")
	mxs, err := net.LookupMX(host)
	if err != nil || len(mxs) == 0 {
		fmt.Println("Bulunamadı.")
	} else {
		for _, mx := range mxs {
			fmt.Printf("%d %s\n", mx.Pref, mx.Host)
		}
	}

	// NS
	fmt.Println("\nNS Records:")
	nss, err := net.LookupNS(host)
	if err != nil || len(nss) == 0 {
		fmt.Println("Bulunamadı.")
	} else {
		for _, ns := range nss {
			fmt.Println(ns.Host)
		}
	}

	// TXT
	fmt.Println("\nTXT Records:")
	txts, err := net.LookupTXT(host)
	if err != nil || len(txts) == 0 {
		fmt.Println("Bulunamadı.")
	} else {
		for _, txt := range txts {
			fmt.Println(txt)
		}
	}

	// PTR
	fmt.Println("\nPTR Records:")
	foundPTR := false
	for _, ip := range ips {
		if ip.To4() == nil {
			continue
		}

		ptrs, err := net.LookupAddr(ip.String())
		if err == nil && len(ptrs) > 0 {
			for _, ptr := range ptrs {
				fmt.Println(ptr)
			}
			foundPTR = true
		}
		break
	}
	if !foundPTR {
		fmt.Println("Bulunamadı.")
	}

	// SRV
	fmt.Println("\nSRV Records:")
	_, srvs, err := net.LookupSRV("sip", "tcp", host)
	if err != nil || len(srvs) == 0 {
		fmt.Println("Bulunamadı.")
	} else {
		for _, srv := range srvs {
			fmt.Printf("%d %d %d %s\n",
				srv.Priority,
				srv.Weight,
				srv.Port,
				srv.Target)
		}
	}

	query(host, dns.TypeSOA)
	query(host, dns.TypeDNSKEY)
	query(host, dns.TypeDS)
}

func query(domain string, recordType uint16) {

	fmt.Printf("\n%s Records:\n", dns.TypeToString[recordType])

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), recordType)

	client := new(dns.Client)

	resp, _, err := client.Exchange(msg, dnsServer)
	if err != nil || len(resp.Answer) == 0 {
		fmt.Println("Bulunamadı.")
		return
	}

	for _, ans := range resp.Answer {

		switch v := ans.(type) {

		case *dns.CAA:
			fmt.Println("Flag :", v.Flag)
			fmt.Println("Tag  :", v.Tag)
			fmt.Println("Value:", v.Value)

		case *dns.SOA:
			fmt.Println("NS      :", v.Ns)
			fmt.Println("MBOX    :", v.Mbox)
			fmt.Println("SERIAL  :", v.Serial)
			fmt.Println("REFRESH :", v.Refresh)
			fmt.Println("RETRY   :", v.Retry)
			fmt.Println("EXPIRE  :", v.Expire)
			fmt.Println("MINIMUM :", v.Minttl)

		case *dns.DNSKEY:
			fmt.Println("Flags     :", v.Flags)
			fmt.Println("Protocol  :", v.Protocol)
			fmt.Println("Algorithm :", v.Algorithm)
			fmt.Println("PublicKey :", v.PublicKey)

		case *dns.DS:
			fmt.Println("KeyTag     :", v.KeyTag)
			fmt.Println("Algorithm  :", v.Algorithm)
			fmt.Println("DigestType :", v.DigestType)
			fmt.Println("Digest     :", v.Digest)

		default:
			fmt.Println(ans)
		}
	}
}
