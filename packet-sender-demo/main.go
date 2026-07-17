package main

import (
	"fmt"

	"github.com/google/gopacket/pcap"
)

func main() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		fmt.Println("Cihaz okunamadı")
		return
	}
	device := devices[4].Name

	protocolSelection := "TCP"

	// kedi := " /\\_/\\\n" +
	// 	"         ( ^.^ )\n" +
	// 	"         > ^ <"

	payload := []byte("hello-world")
	payloadLength := len("payload")

	calculateUDPChecksumOption := true

	ethernet_header := []byte{
		0x54, 0x0d, 0xf9, 0xd0, 0x78, 0x8a, // Destination MAC (Router MAC adresi)
		0xa0, 0x36, 0xbc, 0x32, 0x04, 0xe9, // Source MAC (Kendi MAC adresin)
		0x08, 0x00, // Ethertype: IPv4
	}

	ip_header := []byte{
		0x45,       // Version & IHL
		0x00,       // Type of Service
		0x00, 0x00, // Total Length (Dinamik)
		0x12, 0x34, // Identification
		0x40, 0x00, // Flags & Fragment Offset
		0x80,       // TTL
		0x00,       // Protocol
		0x00, 0x00, // Header Checksum
		192, 168, 30, 110, // Source IP
		192, 168, 30, 110, // Destination IP
	}

	tcp_header := []byte{
		0xd4, 0x31, // Source Port (54321)
		0xd4, 0x32, // Destination Port (54322)
		0x00, 0x00, 0x00, 39, // Sequence Number
		0x00, 0x00, 0x00, 0x00, // Acknowledgment Number
		0x50,       // Data Offset + reserved
		0x18,       // Flags: PSH, ACK
		0xfa, 0xf0, // Window Size
		0x00, 0x00, // Checksum
		0x00, 0x00, // Urgent Pointer
	}

	udp_header := []byte{
		0xd4, 0x31, // 0-1: Source Port
		0xd4, 0x32, // 2-3: Destination Port
		0x00, 0x00, // 4-5: UDP Length
		0x00, 0x00, // 6-7: UDP Checksum
	}

	var l4_header []byte

	if protocolSelection == "TCP" {
		ip_header[9] = 0x06 // TCP Protokol Numarası 6
		totalIPLength := uint16(20 + 20 + payloadLength)
		ip_header[2] = byte(totalIPLength >> 8)
		ip_header[3] = byte(totalIPLength)
		l4_header = tcp_header
	} else {
		ip_header[9] = 0x11 // UDP Protokol Numarası (17)
		totalIPLength := uint16(20 + 8 + payloadLength)
		ip_header[2] = byte(totalIPLength >> 8)
		ip_header[3] = byte(totalIPLength)

		totalUDPLength := uint16(8 + payloadLength)
		udp_header[4] = byte(totalUDPLength >> 8)
		udp_header[5] = byte(totalUDPLength)

		if calculateUDPChecksumOption {
			pseudoHeader := make([]byte, 12)
			copy(pseudoHeader[0:4], ip_header[12:16]) // Source IP
			copy(pseudoHeader[4:8], ip_header[16:20]) // Dest IP
			pseudoHeader[8] = 0x00                    // Reserved (Boş)
			pseudoHeader[9] = 17                      // Protocol (UDP)
			pseudoHeader[10] = byte(totalUDPLength >> 8)
			pseudoHeader[11] = byte(totalUDPLength)

			checksumData := append(pseudoHeader, udp_header...)
			checksumData = append(checksumData, payload...)

			udpChecksum := calculateChecksum(checksumData)

			if udpChecksum == 0 {
				udpChecksum = 0xFFFF
			}

			udp_header[6] = byte(udpChecksum >> 8)
			udp_header[7] = byte(udpChecksum)
		} else {
			udp_header[6] = 0x00
			udp_header[7] = 0x00
		}

		l4_header = udp_header
	}

	ipChecksum := calculateChecksum(ip_header)
	ip_header[10] = byte(ipChecksum >> 8)
	ip_header[11] = byte(ipChecksum)

	finalPacket := append(ethernet_header, ip_header...)
	finalPacket = append(finalPacket, l4_header...)
	finalPacket = append(finalPacket, payload...)

	handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		fmt.Println("HATA: Kart açılamadı:", err)
		return
	}
	defer handle.Close()

	err = handle.WritePacketData(finalPacket)
	if err != nil {
		fmt.Println("[Sürücü Uyarısı]:", err)
	} else {
		fmt.Println("Paket başarıyla enjekte edildi! Wireshark'ta kontrol edebilirsin.")
	}
}

func calculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(data); i += 2 {
		word := uint16(data[i])<<8 | uint16(data[i+1])
		sum += uint32(word)
	}
	if len(data)%2 == 1 {
		sum += uint32(uint16(data[len(data)-1]) << 8)
	}
	for (sum >> 16) != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return ^uint16(sum)
}
