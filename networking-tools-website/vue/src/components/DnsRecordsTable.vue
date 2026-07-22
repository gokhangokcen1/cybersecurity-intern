<script setup>
import { ref } from 'vue'
import { getFullRecords } from '../services/DnsCheckerService'

const domain = ref('')
const sonuc = ref(null)
const hataMesaji = ref('')
const yukleniyor = ref(false)

async function kontrolEt() {
  hataMesaji.value = ''
  sonuc.value = null
  yukleniyor.value = true

  try {
    sonuc.value = await getFullRecords(domain.value)
  } catch (error) {
    if (error.response && error.response.data && error.response.data.error) {
      hataMesaji.value = error.response.data.error
    } else {
      hataMesaji.value = 'Kontrol basarisiz, backend calisiyor mu kontrol et'
    }
  } finally {
    yukleniyor.value = false
  }
}
</script>

<template>
  <div class="container">
    <h1>DNS Kayıt Raporu</h1>

    <form @submit.prevent="kontrolEt">
      <input v-model="domain" placeholder="Ornek: google.com" />
      <button type="submit" :disabled="!domain || yukleniyor">
        {{ yukleniyor ? 'Kontrol Ediliyor...' : 'Kontrol Et' }}
      </button>
    </form>

    <p v-if="hataMesaji" class="hata">{{ hataMesaji }}</p>

    <div v-if="sonuc" class="sonuc">
      <h2>A</h2>
      <table>
        <tbody>
          <tr v-if="!sonuc.a || !sonuc.a.length"><td>Bulunamadi</td></tr>
          <tr v-for="(ip, i) in sonuc.a" :key="i"><td>{{ ip }}</td></tr>
        </tbody>
      </table>

      <h2>AAAA</h2>
      <table>
        <tbody>
          <tr v-if="!sonuc.aaaa || !sonuc.aaaa.length"><td>Bulunamadi</td></tr>
          <tr v-for="(ip, i) in sonuc.aaaa" :key="i"><td>{{ ip }}</td></tr>
        </tbody>
      </table>

      <h2>CNAME</h2>
      <table>
        <tbody><tr><td>{{ sonuc.cname || 'Bulunamadi' }}</td></tr></tbody>
      </table>

      <h2>MX</h2>
      <table>
        <thead><tr><th>Oncelik</th><th>Host</th></tr></thead>
        <tbody>
          <tr v-if="!sonuc.mx || !sonuc.mx.length"><td colspan="2">Bulunamadi</td></tr>
          <tr v-for="(mx, i) in sonuc.mx" :key="i">
            <td>{{ mx.preference }}</td>
            <td>{{ mx.host }}</td>
          </tr>
        </tbody>
      </table>

      <h2>NS</h2>
      <table>
        <tbody>
          <tr v-if="!sonuc.ns || !sonuc.ns.length"><td>Bulunamadi</td></tr>
          <tr v-for="(ns, i) in sonuc.ns" :key="i"><td>{{ ns }}</td></tr>
        </tbody>
      </table>

      <h2>PTR</h2>
      <table>
        <tbody>
          <tr v-if="!sonuc.ptr || !sonuc.ptr.length"><td>Bulunamadi</td></tr>
          <tr v-for="(ptr, i) in sonuc.ptr" :key="i"><td>{{ ptr }}</td></tr>
        </tbody>
      </table>

      <h2>SRV</h2>
      <table>
        <thead><tr><th>Oncelik</th><th>Agirlik</th><th>Port</th><th>Hedef</th></tr></thead>
        <tbody>
          <tr v-if="!sonuc.srv || !sonuc.srv.length"><td colspan="4">Bulunamadi</td></tr>
          <tr v-for="(srv, i) in sonuc.srv" :key="i">
            <td>{{ srv.priority }}</td>
            <td>{{ srv.weight }}</td>
            <td>{{ srv.port }}</td>
            <td>{{ srv.target }}</td>
          </tr>
        </tbody>
      </table>

      <h2>SOA</h2>
      <table>
        <tbody v-if="sonuc.soa">
          <tr><td>NS</td><td>{{ sonuc.soa.ns }}</td></tr>
          <tr><td>Mbox</td><td>{{ sonuc.soa.mbox }}</td></tr>
          <tr><td>Serial</td><td>{{ sonuc.soa.serial }}</td></tr>
          <tr><td>Refresh</td><td>{{ sonuc.soa.refresh }}</td></tr>
          <tr><td>Retry</td><td>{{ sonuc.soa.retry }}</td></tr>
          <tr><td>Expire</td><td>{{ sonuc.soa.expire }}</td></tr>
          <tr><td>Minimum</td><td>{{ sonuc.soa.minimum }}</td></tr>
        </tbody>
        <tbody v-else><tr><td>Bulunamadi</td></tr></tbody>
      </table>

      <h2>TXT</h2>
      <table>
        <tbody>
          <tr v-if="!sonuc.txt || !sonuc.txt.length"><td>Bulunamadi</td></tr>
          <tr v-for="(txt, i) in sonuc.txt" :key="i"><td>{{ txt }}</td></tr>
        </tbody>
      </table>

      <h2>CAA</h2>
      <table>
        <thead><tr><th>Flag</th><th>Tag</th><th>Value</th></tr></thead>
        <tbody>
          <tr v-if="!sonuc.caa || !sonuc.caa.length"><td colspan="3">Bulunamadi</td></tr>
          <tr v-for="(caa, i) in sonuc.caa" :key="i">
            <td>{{ caa.flag }}</td>
            <td>{{ caa.tag }}</td>
            <td>{{ caa.value }}</td>
          </tr>
        </tbody>
      </table>

      <h2>DS</h2>
      <table>
        <thead><tr><th>Key Tag</th><th>Algorithm</th><th>Digest Type</th><th>Digest</th></tr></thead>
        <tbody>
          <tr v-if="!sonuc.ds || !sonuc.ds.length"><td colspan="4">Bulunamadi</td></tr>
          <tr v-for="(ds, i) in sonuc.ds" :key="i">
            <td>{{ ds.keyTag }}</td>
            <td>{{ ds.algorithm }}</td>
            <td>{{ ds.digestType }}</td>
            <td class="mono">{{ ds.digest }}</td>
          </tr>
        </tbody>
      </table>

      <h2>DNSKEY</h2>
      <table>
        <thead><tr><th>Flags</th><th>Protocol</th><th>Algorithm</th><th>Public Key</th></tr></thead>
        <tbody>
          <tr v-if="!sonuc.dnskey || !sonuc.dnskey.length"><td colspan="4">Bulunamadi</td></tr>
          <tr v-for="(key, i) in sonuc.dnskey" :key="i">
            <td>{{ key.flags }}</td>
            <td>{{ key.protocol }}</td>
            <td>{{ key.algorithm }}</td>
            <td class="mono">{{ key.publicKey }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.container { max-width: 700px; margin: 40px auto; font-family: sans-serif; }
form { display: flex; gap: 8px; margin-bottom: 16px; }
input, button { padding: 8px; }
.hata { color: red; }
table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
td, th { border: 1px solid #ccc; padding: 8px; text-align: left; word-break: break-all; }
.mono { font-family: monospace; font-size: 0.8rem; }
</style>