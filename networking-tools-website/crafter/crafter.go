package crafter

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/gokhangokcen1/subnet-backend/models"
	"github.com/google/gopacket/pcap"
)

var (
	seqMutex   sync.Mutex
	lastSeqNum uint32 = 0
)

func GetAndIncrementSeq(payloadLen uint32) uint32 {
	seqMutex.Lock()
	defer seqMutex.Unlock()

	currentSeq := lastSeqNum
	if payloadLen > 0 {
		lastSeqNum += payloadLen
	} else {
		lastSeqNum += 1
	}
	return currentSeq
}

func SendCraftedPacket(req models.PacketCraftRequest) error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return fmt.Errorf("sistemdeki ağ kartları okunamadı: %w", err)
	}
	if len(devices) <= 4 {
		return fmt.Errorf("sistemde en az 5 adet ağ kartı bulunmalıdır (4. indeks bulunamadı)")
	}
	selectedDevice := devices[4].Name

	srcMacBytes, err := net.ParseMAC(req.SrcMAC)
	if err != nil {
		return fmt.Errorf("kaynak MAC adresi geçersiz: %w", err)
	}
	dstMacBytes, err := net.ParseMAC(req.DstMAC)
	if err != nil {
		return fmt.Errorf("hedef MAC adresi geçersiz: %w", err)
	}

	srcIP := net.ParseIP(req.SrcIP).To4()
	if srcIP == nil {
		return fmt.Errorf("kaynak IP geçersiz (sadece IPv4)")
	}
	dstIP := net.ParseIP(req.DstIP).To4()
	if dstIP == nil {
		return fmt.Errorf("hedef IP geçersiz (sadece IPv4)")
	}

	kediPayload := " /\\_/\\\n" +
		"         ( ^.^ )\n" +
		"         > ^ <"

	processedPayloadStr := req.Payload
	if strings.Contains(req.Payload, "kedi") {
		processedPayloadStr = strings.ReplaceAll(req.Payload, "kedi", kediPayload)
	}
	payloadBytes := []byte(processedPayloadStr)
	payloadLength := len(payloadBytes)

	ethernetHeader := append(dstMacBytes, srcMacBytes...)
	ethernetHeader = append(ethernetHeader, 0x08, 0x00) // EtherType: IPv4

	ipHeader := []byte{
		0x45,       // Version & IHL
		0x00,       // Type of Service
		0x00, 0x00, // Toplam Uzunluk (Aşağıda hesaplanacak)
		0x12, 0x34, // Identification
		0x40, 0x00, // Flags & Fragment Offset
		0x80,       // TTL
		0x00,       // Protocol
		0x00, 0x00, // Header Checksum
	}
	ipHeader = append(ipHeader, srcIP...)
	ipHeader = append(ipHeader, dstIP...)

	var l4Header []byte

	if strings.ToUpper(req.Protocol) == "TCP" {
		ipHeader[9] = 0x06 // IP Üzeri Protokol: TCP

		totalIPLength := uint16(20 + 20 + payloadLength)
		ipHeader[2] = byte(totalIPLength >> 8)
		ipHeader[3] = byte(totalIPLength)

		var flagByte byte = 0x00
		for _, flag := range req.TCPFlags {
			switch strings.ToUpper(flag) {
			case "FIN":
				flagByte |= 0x01
			case "SYN":
				flagByte |= 0x02
			case "RST":
				flagByte |= 0x04
			case "PSH":
				flagByte |= 0x08
			case "ACK":
				flagByte |= 0x10
			case "URG":
				flagByte |= 0x20
			}
		}
		if flagByte == 0 {
			flagByte = 0x18 // Varsayılan: PSH | ACK
		}

		seqNum := GetAndIncrementSeq(uint32(payloadLength))

		var ackNum uint32 = 0
		if (flagByte & 0x10) != 0 { // ACK biti aktifse
			ackNum = 5001
		}

		tcpHeader := []byte{
			byte(req.SrcPort >> 8), byte(req.SrcPort), // Src Port
			byte(req.DstPort >> 8), byte(req.DstPort), // Dst Port
			byte(seqNum >> 24), byte(seqNum >> 16), byte(seqNum >> 8), byte(seqNum), // Sequence
			byte(ackNum >> 24), byte(ackNum >> 16), byte(ackNum >> 8), byte(ackNum), // Acknowledgment
			0x50,       // Data Offset (20 byte)
			flagByte,   // Flags
			0xfa, 0xf0, // Window Size
			0x00, 0x00, // Checksum (Aşağıda hesaplanacak)
			0x00, 0x00, // Urgent Pointer
		}

		pseudoHeader := make([]byte, 12)
		copy(pseudoHeader[0:4], srcIP)
		copy(pseudoHeader[4:8], dstIP)
		pseudoHeader[8] = 0x00
		pseudoHeader[9] = 0x06
		totalTCPLength := uint16(20 + payloadLength)
		pseudoHeader[10] = byte(totalTCPLength >> 8)
		pseudoHeader[11] = byte(totalTCPLength)

		checksumData := append(pseudoHeader, tcpHeader...)
		checksumData = append(checksumData, payloadBytes...)
		tcpChecksum := calculateChecksum(checksumData)

		tcpHeader[16] = byte(tcpChecksum >> 8)
		tcpHeader[17] = byte(tcpChecksum)

		l4Header = tcpHeader

	} else {
		ipHeader[9] = 0x11 // IP Üzeri Protokol: UDP

		totalIPLength := uint16(20 + 8 + payloadLength)
		ipHeader[2] = byte(totalIPLength >> 8)
		ipHeader[3] = byte(totalIPLength)

		totalUDPLength := uint16(8 + payloadLength)

		udpHeader := []byte{
			byte(req.SrcPort >> 8), byte(req.SrcPort), // Src Port
			byte(req.DstPort >> 8), byte(req.DstPort), // Dst Port
			byte(totalUDPLength >> 8), byte(totalUDPLength), // UDP Length
			0x00, 0x00, // Checksum
		}

		pseudoHeader := make([]byte, 12)
		copy(pseudoHeader[0:4], srcIP)
		copy(pseudoHeader[4:8], dstIP)
		pseudoHeader[8] = 0x00
		pseudoHeader[9] = 17
		pseudoHeader[10] = byte(totalUDPLength >> 8)
		pseudoHeader[11] = byte(totalUDPLength)

		checksumData := append(pseudoHeader, udpHeader...)
		checksumData = append(checksumData, payloadBytes...)
		udpChecksum := calculateChecksum(checksumData)

		if udpChecksum == 0 {
			udpChecksum = 0xFFFF
		}

		udpHeader[6] = byte(udpChecksum >> 8)
		udpHeader[7] = byte(udpChecksum)

		l4Header = udpHeader
	}

	ipChecksum := calculateChecksum(ipHeader)
	ipHeader[10] = byte(ipChecksum >> 8)
	ipHeader[11] = byte(ipChecksum)

	finalPacket := append(ethernetHeader, ipHeader...)
	finalPacket = append(finalPacket, l4Header...)
	finalPacket = append(finalPacket, payloadBytes...)

	handle, err := pcap.OpenLive(selectedDevice, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("sabitlenmiş %s ağ arayüzü açılamadı: %w", selectedDevice, err)
	}
	defer handle.Close()

	err = handle.WritePacketData(finalPacket)
	if err != nil {
		return fmt.Errorf("paket enjekte edilemedi: %w", err)
	}

	return nil
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
