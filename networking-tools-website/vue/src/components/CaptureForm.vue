<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import CaptureService from '../services/CaptureService'

const cihazlar = ref([])
const secilenCihaz = ref('')
const paketler = ref([])
const secilenIndex = ref(null)
const yakalaniyor = ref(false)
const hataMesaji = ref('')

const showTCP = ref(true)
const showUDP = ref(true)
const showARP = ref(true)
const showICMP = ref(true)
const showOther = ref(true)

let socket = null

const filtrelenmisPaketler = computed(() => {
  return paketler.value.filter((p) => {
    if (p.protocol === 'TCP') return showTCP.value
    if (p.protocol === 'UDP') return showUDP.value
    if (p.protocol === 'ARP') return showARP.value
    if (p.protocol === 'ICMP') return showICMP.value
    return showOther.value
  })
})

const secilenPaket = computed(() => {
  if (secilenIndex.value === null) return null
  return filtrelenmisPaketler.value[secilenIndex.value] || null
})

async function cihazlariGetir() {
  try {
    const response = await CaptureService.cihazlariGetir()
    cihazlar.value = response.data
  } catch (error) {
    hataMesaji.value = 'Cihaz listesi alinamadi'
  }
}

function websocketBaglan() {
  socket = new WebSocket(CaptureService.websocketUrl())
  socket.onmessage = (event) => {
    const info = JSON.parse(event.data)
    paketler.value.push(info)
  }
  socket.onerror = () => {
    hataMesaji.value = 'WebSocket hatasi'
  }
  socket.onclose = () => {
    yakalaniyor.value = false
  }
}

async function baslat() {
  hataMesaji.value = ''
  try {
    await CaptureService.baslat(secilenCihaz.value)
    yakalaniyor.value = true
    websocketBaglan()
  } catch (error) {
    hataMesaji.value = error.response?.data?.error || 'Capture baslatilamadi'
  }
}

async function durdur() {
  try {
    await CaptureService.durdur()
  } catch (error) {
    hataMesaji.value = 'Capture durdurulamadi'
  } finally {
    yakalaniyor.value = false
    if (socket) {
      socket.close()
      socket = null
    }
  }
}

function temizle() {
  paketler.value = []
  secilenIndex.value = null
}

function icmpTypeAdi(typeCode) {
  const type = (typeCode >> 8) & 0xff
  if (type === 8) return 'Echo Request (ping)'
  if (type === 0) return 'Echo Reply (pong)'
  if (type === 3) return 'Destination Unreachable'
  if (type === 11) return 'Time Exceeded'
  return `Type ${type}`
}

onMounted(() => {
  cihazlariGetir()
})

onBeforeUnmount(() => {
  if (socket) socket.close()
})
</script>

<template>
  <div class="capture-form">
    <div style="display: flex; align-items: center; justify-content: center; gap: 10px;">
      <!-- <img src="/wirehamsilogo.png" alt="WireHamsi Logo" style="width: 90px; height: 90px;" />-->
      <h1 style="margin-top: 20px;">WireHamsi</h1>
    </div>

    <div class="toolbar">
      <select v-model="secilenCihaz">
        <option disabled value="">Cihaz sec</option>
        <option v-for="d in cihazlar" :key="d.name" :value="d.name">
          {{ d.description || d.name }}
        </option>
      </select>

      <button v-if="!yakalaniyor" @click="baslat" :disabled="!secilenCihaz">Baslat</button>
      <button v-else @click="durdur">Durdur</button>

      <button @click="temizle" :disabled="paketler.length === 0">Temizle</button>
    </div>

    <div class="display-filter">
      <label><input type="checkbox" v-model="showTCP" /> TCP</label>
      <label><input type="checkbox" v-model="showUDP" /> UDP</label>
      <label><input type="checkbox" v-model="showARP" /> ARP</label>
      <label><input type="checkbox" v-model="showICMP" /> ICMP</label>
      <label><input type="checkbox" v-model="showOther" /> Diğer</label>
    </div>

    <p v-if="hataMesaji" class="hata">{{ hataMesaji }}</p>

    <div class="durum">
      Durum: {{ yakalaniyor ? 'Yakalaniyor...' : 'Durduruldu' }} —
      {{ filtrelenmisPaketler.length }} / {{ paketler.length }} paket gosteriliyor
    </div>

    <div class="ana-alan">
      <table class="paket-tablosu">
        <thead>
          <tr>
            <th>#</th>
            <th>Kaynak</th>
            <th>Hedef</th>
            <th>Protokol</th>
            <th>Port/Detay</th>
            <th>Uzunluk</th>
            <th>Flags</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(p, index) in filtrelenmisPaketler"
            :key="index"
            :class="{ secili: index === secilenIndex, tcp: p.protocol === 'TCP', udp: p.protocol === 'UDP', arp: p.protocol === 'ARP', icmp: p.protocol === 'ICMP' }"
            @click="secilenIndex = index"
          >
            <td>{{ index + 1 }}</td>
            <td>{{ p.protocol === 'ARP' ? p.srcMac : p.srcIp }}</td>
            <td>{{ p.protocol === 'ARP' ? p.dstMac : p.dstIp }}</td>
            <td>{{ p.protocol }}</td>
            <td v-if="p.protocol === 'ARP'">
              {{ p.arpOperation }}: {{ p.arpSenderIp }} → {{ p.arpTargetIp }}
            </td>
            <td v-else-if="p.protocol === 'ICMP'">
              {{ icmpTypeAdi(p.typeCode) }}
            </td>
            <td v-else>{{ p.srcPort }} → {{ p.dstPort }}</td>
            <td>{{ p.length }}</td>
            <td>{{ p.flags }}</td>
          </tr>
        </tbody>
      </table>

      <div v-if="secilenPaket" class="detay">
        <h3>Paket Detayi</h3>

        <template v-if="secilenPaket.protocol === 'ARP'">
          <p><strong>Islem:</strong> {{ secilenPaket.arpOperation }}</p>
          <p><strong>Sender MAC:</strong> {{ secilenPaket.srcMac }}</p>
          <p><strong>Sender IP:</strong> {{ secilenPaket.arpSenderIp }}</p>
          <p><strong>Target MAC:</strong> {{ secilenPaket.dstMac }}</p>
          <p><strong>Target IP:</strong> {{ secilenPaket.arpTargetIp }}</p>
        </template>

        <template v-else-if="secilenPaket.protocol === 'ICMP'">
          <p><strong>Kaynak IP:</strong> {{ secilenPaket.srcIp }}</p>
          <p><strong>Hedef IP:</strong> {{ secilenPaket.dstIp }}</p>
          <p><strong>Tip:</strong> {{ icmpTypeAdi(secilenPaket.typeCode) }}</p>
          <p><strong>Ham TypeCode:</strong> {{ secilenPaket.typeCode }}</p>
        </template>

        <template v-else>
          <p><strong>Kaynak IP:</strong> {{ secilenPaket.srcIp }}</p>
          <p><strong>Hedef IP:</strong> {{ secilenPaket.dstIp }}</p>
          <p><strong>Src Port:</strong> {{ secilenPaket.srcPort }}</p>
          <p><strong>Dst Port:</strong> {{ secilenPaket.dstPort }}</p>
          <p><strong>Flags:</strong> {{ secilenPaket.flags || '—' }}</p>
          <h4>Payload (Hex + ASCII)</h4>
          <pre class="hex-dump">{{ secilenPaket.hexDump || 'Payload yok' }}</pre>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.capture-form { font-family: monospace; }
.toolbar { display: flex; gap: 8px; margin-bottom: 8px; }
.display-filter { display: flex; gap: 16px; margin-bottom: 8px; }
.durum { font-size: 0.85rem; color: #555; margin-bottom: 8px; }
.hata { color: red; }
.ana-alan { display: flex; gap: 12px; }
.paket-tablosu { flex: 2; border-collapse: collapse; width: 100%; font-size: 0.8rem; }
.paket-tablosu th, .paket-tablosu td { border: 1px solid #ccc; padding: 4px 8px; text-align: left; }
.paket-tablosu tr.tcp { background: #eef5ff; }
.paket-tablosu tr.udp { background: #fff8e8; }
.paket-tablosu tr.arp { background: #f0fff0; }
.paket-tablosu tr.icmp { background: #ffe8f0; }
.paket-tablosu tr.secili { background: #ffe08a !important; }
.paket-tablosu tr { cursor: pointer; }
.detay { flex: 1; font-size: 0.85rem; }
.hex-dump { background: #1e1e1e; color: #d4d4d4; padding: 8px; border-radius: 4px; font-size: 0.75rem; overflow-x: auto; }
</style>
