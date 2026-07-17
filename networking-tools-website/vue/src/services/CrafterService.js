import axios from "axios"

const API_URL = "http://localhost:3001/api"

export default {
  // Sunucuya özel paketi hazırlar ve enjekte etmesi için istek atar
  paketGonder(packetData) {
    return axios.post(`${API_URL}/packet-sender`, packetData)
  }
}