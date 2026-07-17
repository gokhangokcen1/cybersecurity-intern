package models

// PacketCraftRequest kullanıcının arayüzden gönderdiği ham paket parametreleri
type PacketCraftRequest struct {
	Protocol string   `json:"protocol"` // "TCP" veya "UDP"
	SrcMAC   string   `json:"srcMac"`
	DstMAC   string   `json:"dstMac"`
	SrcIP    string   `json:"srcIp"`
	DstIP    string   `json:"dstIp"`
	SrcPort  uint16   `json:"srcPort"`
	DstPort  uint16   `json:"dstPort"`
	Payload  string   `json:"payload"`
	TCPFlags []string `json:"tcpFlags"` // ["SYN", "ACK", "PSH", vs.]
}

// PacketCraftResponse işlem sonucunu döneceğimiz yapı
type PacketCraftResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
