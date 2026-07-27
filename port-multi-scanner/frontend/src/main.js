import { createApp } from 'vue/dist/vue.esm-bundler.js'
import './style.css'

const commonPorts = [20, 21, 22, 23, 25, 53, 80, 110, 143, 443, 445, 3306, 3389, 5432, 5900, 8080, 8443]

createApp({
  data: () => ({
    devices: [], results: [], status: {}, error: '', notice: '', loading: true,
    mode: 'common', portsText: commonPorts.join(', '), startPort: 1, endPort: 1024, timeoutMs: 250, workers: 4000,
    newDevice: '', newCustomer: '', newDealer: '', newIP: '', refreshTimer: null
  }),
  computed: {
    running () { return !!this.status.running },
    targetCount () { return this.devices.reduce((count, device) => count + (device.ip_addresses || []).length, 0) },
    ports () { return [...new Set((this.portsText.match(/\d+/g) || []).map(Number).filter(port => port > 0 && port <= 65535))].sort((a, b) => a - b) },
    selectedPorts () { return this.mode === 'all' ? 65535 : this.mode === 'range' ? Math.max(0, this.endPort - this.startPort + 1) : this.ports.length },
    totalChecks () { return this.selectedPorts * this.targetCount },
    progress () { return this.status.total ? Math.round(this.status.completed * 100 / this.status.total) : 0 },
    groupedResults () {
      const groups = {}
      this.results.forEach(result => {
        if (!groups[result.ip]) groups[result.ip] = { ip: result.ip, device: result.deviceName, customer: result.customerName, ports: [] }
        groups[result.ip].ports.push(result.port)
      })
      return Object.values(groups).map(group => ({ ...group, ports: [...new Set(group.ports)].sort((a, b) => a - b) })).sort((a, b) => a.ip.localeCompare(b.ip))
    }
  },
  async mounted () { await this.refresh(); this.listenEvents(); if (this.running) this.startLiveRefresh() },
  methods: {
    async refresh () {
      try {
        const [feedResponse, dashboardResponse] = await Promise.all([fetch('/api/feed'), fetch('/api/dashboard')])
        if (!feedResponse.ok || !dashboardResponse.ok) throw new Error('Backend yanit vermedi.')
        const feed = await feedResponse.json(); const dashboard = await dashboardResponse.json()
        this.devices = feed.devices || []; this.results = dashboard.results || []; this.status = (dashboard.statuses || [])[0] || {}
        this.error = ''
      } catch (error) { this.error = 'Backend calismiyor. Backend icin: go run ./backend' }
      finally { this.loading = false }
    },
    listenEvents () {
      const events = new EventSource('/api/events')
      events.addEventListener('result', event => { const item = JSON.parse(event.data).result; if (item) this.results.unshift(item) })
      events.addEventListener('progress', event => { const data = JSON.parse(event.data); this.status = { ...this.status, running: true, completed: data.completed, total: data.total, currentPort: data.port } })
      events.addEventListener('finished', () => { this.status.running = false; this.stopLiveRefresh(); this.refresh() })
      events.addEventListener('cancelled', () => { this.status.running = false; this.stopLiveRefresh(); this.refresh() })
    },
    startLiveRefresh () {
      this.stopLiveRefresh()
      this.refreshTimer = window.setInterval(async () => {
        await this.refresh()
        if (!this.running) this.stopLiveRefresh()
      }, 700)
    },
    stopLiveRefresh () {
      if (this.refreshTimer) window.clearInterval(this.refreshTimer)
      this.refreshTimer = null
    },
    addIP () {
      const ip = this.newIP.trim(); const deviceName = this.newDevice.trim() || 'Yeni cihaz'
      if (!ip) { this.error = 'Eklenecek IP adresini girin.'; return }
      let device = this.devices.find(item => item.device_name === deviceName)
      if (!device) { device = { device_name: deviceName, customer_name: this.newCustomer.trim(), dealer_name: this.newDealer.trim(), ip_addresses: [] }; this.devices.push(device) }
      if (device.ip_addresses.includes(ip)) { this.error = 'Bu IP zaten listede.'; return }
      device.ip_addresses.push(ip); this.saveFeed(); this.newIP = ''
    },
    removeIP (deviceIndex, ipIndex) {
      const device = this.devices[deviceIndex]; device.ip_addresses.splice(ipIndex, 1)
      if (!device.ip_addresses.length) this.devices.splice(deviceIndex, 1)
      this.saveFeed()
    },
    async saveFeed () {
      this.error = ''; this.notice = ''
      const response = await fetch('/api/feed', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ devices: this.devices }) })
      const data = await response.json()
      if (!response.ok) { this.error = data.error || 'IP listesi kaydedilemedi.'; await this.refresh(); return }
      this.notice = data.message
    },
    scanRequest () {
      const base = { timeoutMs: Number(this.timeoutMs), workers: Number(this.workers) }
      if (this.mode === 'all') return { ...base, startPort: 1, endPort: 65535 }
      if (this.mode === 'range') return { ...base, startPort: Number(this.startPort), endPort: Number(this.endPort) }
      return { ...base, ports: this.ports }
    },
    async startScan () {
      this.error = ''; this.notice = ''
      if (!this.targetCount || !this.selectedPorts) { this.error = 'En az bir IP ve port secin.'; return }
      const response = await fetch('/api/scan', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.scanRequest()) })
      const data = await response.json()
      if (!response.ok) { this.error = data.error || 'Tarama baslatilamadi.'; return }
      this.results = []
      this.status = { ...this.status, running: true, completed: 0, total: this.totalChecks }
      this.startLiveRefresh()
      await this.refresh()
    },
    async stopScan () { await fetch('/api/scan', { method: 'DELETE' }); this.stopLiveRefresh(); await this.refresh() }
  },
  template: `
  <main class="shell">
    <header><div><div class="brand"><span>◎</span> Port Scanner</div><p>Port bazli, kontrollu batch tarama</p></div><div class="source">{{ targetCount }} hedef IP</div></header>
    <div v-if="error" class="alert error">{{ error }}</div><div v-if="notice" class="alert success">{{ notice }}</div>
    <section class="grid">
      <article class="card targets"><div class="title"><span>01</span><h2>Hedef IP listesi</h2><em>{{ targetCount }} IP</em></div>
        <div v-if="loading" class="empty">Liste yukleniyor…</div>
        <div class="device" v-for="(device, deviceIndex) in devices" :key="device.device_name"><b>{{ device.device_name }}</b><small>{{ device.customer_name || 'Musteri yok' }} · {{ device.dealer_name || 'Bayi yok' }}</small><div class="ip-row" v-for="(ip, ipIndex) in device.ip_addresses" :key="ip"><code>{{ ip }}</code><button class="remove" :disabled="running" @click="removeIP(deviceIndex, ipIndex)">Kaldir</button></div></div>
        <form class="add-ip" @submit.prevent="addIP"><input v-model="newDevice" placeholder="Cihaz adi (opsiyonel)"><input v-model="newIP" placeholder="IP adresi" required><input v-model="newCustomer" placeholder="Musteri (opsiyonel)"><input v-model="newDealer" placeholder="Bayi (opsiyonel)"><button :disabled="running">IP ekle</button></form>
      </article>
      <article class="card controls"><div class="title"><span>02</span><h2>Port tarama</h2><em>{{ selectedPorts }} port</em></div>
        <label class="radio"><input v-model="mode" type="radio" value="common"><b>Yaygin portlar</b><small>{{ ports.length }} servis portu</small></label>
        <label class="radio"><input v-model="mode" type="radio" value="range"><b>Port araligi</b><small>Baslangic ve bitisi secin</small></label><div v-if="mode === 'range'" class="range"><input v-model.number="startPort" min="1" max="65535" type="number"><span>—</span><input v-model.number="endPort" min="1" max="65535" type="number"></div>
        <label class="radio"><input v-model="mode" type="radio" value="all"><b>Tum portlar</b><small>1 – 65535</small></label>
        <label class="radio"><input v-model="mode" type="radio" value="custom"><b>Belirli portlar</b><small>Virgulle ayirin</small></label><input v-if="mode === 'custom'" v-model="portsText" class="input" placeholder="22, 80, 443">
        <div class="scan-settings"><label>Baglanti zaman asimi (ms)<input v-model.number="timeoutMs" min="25" max="30000" type="number"></label><label>Eszamanli goroutine<input v-model.number="workers" min="1" max="5000" type="number"></label></div>
        <p class="batch-note">Batch boyutu goroutine ve hedef IP sayisindan otomatik hesaplanir. Batch bitince sonraki port grubuna gecilir.</p>
        <button class="primary" :disabled="running" @click="startScan">Portlari tara · {{ totalChecks.toLocaleString('tr-TR') }} istek</button><button v-if="running" class="stop" @click="stopScan">Taramayi durdur</button>
      </article>
      <article class="card progress"><div class="title"><span>03</span><h2>Tarama durumu</h2><em>{{ running ? 'Devam ediyor' : 'Hazir' }}</em></div><strong>%{{ progress }}</strong><div class="bar"><i :style="{ width: progress + '%' }"></i></div><p>Port {{ status.currentPort || '—' }} · {{ status.completed || 0 }} / {{ status.total || 0 }} istek</p><small>{{ status.workerLimit || workers }} goroutine · Batch: {{ status.portsPerBatch || Math.max(1, Math.floor(workers / Math.max(targetCount, 1))) }} port × {{ targetCount }} IP</small></article>
    </section>
    <section class="card results"><div class="title"><span>04</span><h2>Acik portlar</h2><em>{{ results.length }} bulgu</em></div><div v-if="!groupedResults.length" class="empty">Tarama sonuclari burada gorunecek.</div><div class="result-grid"><article v-for="item in groupedResults" :key="item.ip" class="result-item"><code>{{ item.ip }}</code><b>{{ item.device || 'Cihaz bilgisi yok' }}</b><small>{{ item.customer || '—' }}</small><div class="port-tags"><span v-for="port in item.ports" :key="port">{{ port }}</span></div></article></div></section>
  </main>`
}).mount('#app')
