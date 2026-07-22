<template>
  <div class="dns-checker-page">
    <h2>DNS Propagation Checker</h2>

    <!-- Form & Bütün Kayıt Tiplerini İçeren Seçim Alanı -->
    <form @submit.prevent="handleCheck" class="dns-form">
      <select v-model="selectedType" class="type-select">
        <option value="A">A</option>
        <option value="AAAA">AAAA</option>
        <option value="CNAME">CNAME</option>
        <option value="MX">MX</option>
        <option value="NS">NS</option>
        <option value="PTR">PTR</option>
        <option value="SRV">SRV</option>
        <option value="SOA">SOA</option>
        <option value="TXT">TXT</option>
        <option value="CAA">CAA</option>
        <option value="DS">DS</option>
        <option value="DNSKEY">DNSKEY</option>
      </select>

      <input v-model="domain" type="text" placeholder="example.com" required />
      
      <button type="submit" :disabled="loading">
        {{ loading ? 'Sorgulanıyor...' : 'Kontrol Et' }}
      </button>
    </form>

    <p v-if="error" class="error">{{ error }}</p>

    <!-- Harita Ekranı -->
    <DnsCheckerForm :results="propagationResults" />

    <!-- SUNUCU BAZLI CANLI KAYIT LİSTESİ -->
    <div v-if="propagationResults.length" class="propagation-list">
      <h3>Farklı Sunuculardan Dönen [{{ selectedType }}] Kayıtları</h3>
      <table>
        <thead>
          <tr>
            <th>Durum</th>
            <th>DNS Sunucu IP</th>
            <th>Konum</th>
            <th>Dönen {{ selectedType }} Yanıtı</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(res, index) in propagationResults" :key="index">
            <td>
              <span :class="['status-badge', res.resolved ? 'success' : 'fail']">
                {{ res.resolved ? '✅ OK' : '❌ Bulunamadı' }}
              </span>
            </td>
            <td><code>{{ res.ip }}</code></td>
            <td>{{ res.city ? `${res.city}, ${res.country}` : 'Bilinmiyor' }}</td>
            <td>
              <div v-if="res.records && res.records.length">
                <div v-for="(rec, i) in res.records" :key="i" class="record-tag">
                  {{ rec }}
                </div>
              </div>
              <div v-else-if="res.ips && res.ips.length">
                <div v-for="(ip, i) in res.ips" :key="i" class="record-tag">
                  {{ ip }}
                </div>
              </div>
              <span v-else class="no-record">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { checkDns, getFullRecords } from '@/services/DnsCheckerService';
import DnsCheckerForm from '@/components/DnsCheckerForm.vue';

const domain = ref('');
const selectedType = ref('A');
const propagationResults = ref([]);
const fullRecords = ref(null);
const loading = ref(false);
const error = ref('');

async function handleCheck() {
  if (!domain.value) return;
  loading.value = true;
  error.value = '';

  try {
    const [propagation, records] = await Promise.all([
      checkDns(domain.value, selectedType.value),
      getFullRecords(domain.value),
    ]);
    propagationResults.value = propagation;
    fullRecords.value = records;
  } catch (e) {
    error.value = e.message || 'DNS sorgusu başarısız.';
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.dns-form {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}
.type-select {
  padding: 8px 12px;
  font-weight: bold;
  border-radius: 6px;
  border: 1px solid #ccc;
  background-color: #f8fafc;
}
.propagation-list {
  margin-top: 25px;
}
table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}
th, td {
  border: 1px solid #e5e7eb;
  padding: 10px;
  text-align: left;
}
th {
  background-color: #f9fafb;
}
.status-badge.success { color: #16a34a; font-weight: bold; }
.status-badge.fail { color: #dc2626; font-weight: bold; }
.record-tag {
  display: block;
  background: #f1f5f9;
  color: #0f172a;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: monospace;
  margin-bottom: 4px;
  word-break: break-all;
}
.no-record { color: #9ca3af; }
.error { color: #ef4444; }
</style>