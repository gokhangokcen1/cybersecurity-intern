<script setup>
import { ref } from 'vue'
import CrafterService from '../services/CrafterService'

const protokol = ref('TCP')
const kaynakMac = ref('a0:36:bc:32:04:e9')
const hedefMac = ref('ac:a7:f1:b9:07:ed')
const kaynakIp = ref('192.168.1.103')
const hedefIp = ref('192.168.1.103')
const kaynakPort = ref(54321)
const hedefPort = ref(54322)
const payload = ref('') // Payload alanı boş başlıyor

// TCP Flags
const flagSyn = ref(true)
const flagAck = ref(true)
const flagPsh = ref(true)
const flagFin = ref(false)
const flagRst = ref(false)
const flagUrg = ref(false)

const gonderiliyor = ref(false)
const hataMesaji = ref('')
const basariMesaji = ref('')

async function paketGonder() {
  hataMesaji.value = ''
  basariMesaji.value = ''
  gonderiliyor.value = true

  const seciliBayraklar = []
  if (flagSyn.value) seciliBayraklar.push('SYN')
  if (flagAck.value) seciliBayraklar.push('ACK')
  if (flagPsh.value) seciliBayraklar.push('PSH')
  if (flagFin.value) seciliBayraklar.push('FIN')
  if (flagRst.value) seciliBayraklar.push('RST')
  if (flagUrg.value) seciliBayraklar.push('URG')

  const parametreler = {
    protocol: protokol.value,
    srcMac: kaynakMac.value,
    dstMac: hedefMac.value,
    srcIp: kaynakIp.value,
    dstIp: hedefIp.value,
    srcPort: parseInt(kaynakPort.value),
    dstPort: parseInt(hedefPort.value),
    payload: payload.value,
    tcpFlags: seciliBayraklar
  }

  try {
    const response = await CrafterService.paketGonder(parametreler)
    if (response.data.success) {
      basariMesaji.value = response.data.message
    } else {
      hataMesaji.value = response.data.message
    }
  } catch (error) {
    hataMesaji.value = error.response?.data?.message || 'Bağlantı hatası: Paket enjekte edilemedi.'
  } finally {
    gonderiliyor.value = false
  }
}
</script>

<template>
  <div class="crafter-container">
    <div class="header-section">
      <h1>Paket Yollama</h1>
    </div>

    <!-- Hata / Başarı Bildirim Panelleri -->
    <div class="konsol-mesaj hata" v-if="hataMesaji">
      <span class="konsol-tag">[HATA]</span> {{ hataMesaji }}
    </div>
    <div class="konsol-mesaj basari" v-if="basariMesaji">
      <span class="konsol-tag">[BAŞARILI]</span> {{ basariMesaji }}
    </div>

    <div class="panel-layout">
      <!-- SOL PANEL: IP / PORT / MAC -->
      <div class="kontrol-paneli">
        <h3 class="panel-baslik">Ağ & Katman Yapılandırması</h3>
        
        <div class="input-alani">
          <label>Protokol Seçimi</label>
          <div class="secenek-grup">
            <label class="secenek">
              <input type="radio" value="TCP" v-model="protokol" /> TCP Protokolü
            </label>
            <label class="secenek">
              <input type="radio" value="UDP" v-model="protokol" /> UDP Protokolü
            </label>
          </div>
        </div>

        <div class="ciftli-satir">
          <div class="input-alani">
            <label>Kaynak MAC</label>
            <input type="text" v-model="kaynakMac" placeholder="a0:36:bc:32:04:e9" />
          </div>
          <div class="input-alani">
            <label>Hedef MAC</label>
            <input type="text" v-model="hedefMac" placeholder="ac:a7:f1:b9:07:ed" />
          </div>
        </div>

        <div class="ciftli-satir">
          <div class="input-alani">
            <label>Kaynak IP (IPv4)</label>
            <input type="text" v-model="kaynakIp" placeholder="192.168.1.103" />
          </div>
          <div class="input-alani">
            <label>Hedef IP (IPv4)</label>
            <input type="text" v-model="hedefIp" placeholder="192.168.1.103" />
          </div>
        </div>

        <div class="ciftli-satir">
          <div class="input-alani">
            <label>Kaynak Port</label>
            <input type="number" v-model="kaynakPort" />
          </div>
          <div class="input-alani">
            <label>Hedef Port</label>
            <input type="number" v-model="hedefPort" />
          </div>
        </div>
      </div>

      <!-- SAĞ PANEL: TCP FLAGS / PAYLOAD -->
      <div class="kontrol-paneli flex-dikey">
        <div>
          <h3 class="panel-baslik">Veri ve Bayrak Yapılandırması</h3>

          <!-- TCP Flags Seçenekleri (Sadece TCP Aktifken Gösterilir) -->
          <div class="input-alani" v-if="protokol === 'TCP'">
            <label>TCP Bayrakları (Flags)</label>
            <div class="flags-izgara">
              <label class="bayrak-kart"><input type="checkbox" v-model="flagSyn" /> SYN</label>
              <label class="bayrak-kart"><input type="checkbox" v-model="flagAck" /> ACK</label>
              <label class="bayrak-kart"><input type="checkbox" v-model="flagPsh" /> PSH</label>
              <label class="bayrak-kart"><input type="checkbox" v-model="flagFin" /> FIN</label>
              <label class="bayrak-kart"><input type="checkbox" v-model="flagRst" /> RST</label>
              <label class="bayrak-kart"><input type="checkbox" v-model="flagUrg" /> URG</label>
            </div>
          </div>

          <div class="input-alani">
            <label>Payload (Paket Mesaj Verisi)</label>
            <textarea v-model="payload" rows="6" placeholder="Ağa göndermek istediğiniz veriyi yazın..."></textarea>
          </div>
        </div>

        <button @click="paketGonder" :disabled="gonderiliyor" class="tetikleyici-btn">
          {{ gonderiliyor ? 'Paket gönderiliyor...' : 'Paket Gönder' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.crafter-container {
  font-family: monospace;
}
.header-section {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 25px;
}
.app-logo {
  width: 70px;
  height: 70px;
}
.badge-v3 {
  background: #4f46e5;
  color: #fff;
  font-size: 0.85rem;
  padding: 4px 10px;
  border-radius: 4px;
}
.panel-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}
@media (max-width: 768px) {
  .panel-layout {
    grid-template-columns: 1fr;
  }
}
.kontrol-paneli {
  background: #ffffff;
  border: 1px solid #ccc;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.03);
}
.flex-dikey {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}
.panel-baslik {
  margin-top: 0;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 10px;
  color: #1e293b;
}
.input-alani {
  margin-bottom: 16px;
}
.input-alani label {
  display: block;
  font-weight: bold;
  margin-bottom: 6px;
  font-size: 0.85rem;
  color: #475569;
}
.ciftli-satir {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
input[type="text"], input[type="number"], select, textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.9rem;
  box-sizing: border-box;
}
input:focus, select:focus, textarea:focus {
  border-color: #6366f1;
  outline: none;
}
.secenek-grup {
  display: flex;
  gap: 20px;
  padding: 5px 0;
}
.secenek {
  cursor: pointer;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 6px;
}
.flags-izgara {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.bayrak-kart {
  border: 1px solid #e2e8f0;
  padding: 8px;
  border-radius: 6px;
  background: #f8fafc;
  cursor: pointer;
  user-select: none;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: bold;
}
.tetikleyici-btn {
  background: #10b981;
  color: white;
  border: none;
  padding: 14px;
  border-radius: 6px;
  font-weight: bold;
  cursor: pointer;
  font-family: monospace;
  font-size: 1rem;
  transition: background 0.2s;
  margin-top: 15px;
}
.tetikleyici-btn:hover {
  background: #059669;
}
.tetikleyici-btn:disabled {
  background: #94a3b8;
  cursor: not-allowed;
}
.konsol-mesaj {
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 20px;
  font-size: 0.9rem;
}
.konsol-mesaj.hata {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #f87171;
}
.konsol-mesaj.basari {
  background: #d1fae5;
  color: #065f46;
  border: 1px solid #34d399;
}
.konsol-tag {
  font-weight: bold;
}
</style>