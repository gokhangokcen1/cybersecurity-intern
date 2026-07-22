<template>
  <div ref="mapContainer" class="dns-map"></div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

const props = defineProps({
  results: { type: Array, default: () => [] },
});

const mapContainer = ref(null);
let map = null;
let markers = [];

function clearMarkers() {
  markers.forEach((m) => map.removeLayer(m));
  markers = [];
}

function renderMarkers(results) {
  if (!map) return;
  clearMarkers();

  results.forEach((item) => {
    if (!item.lat && !item.lon) return;

    const isSuccess = item.resolved;
    const color = isSuccess ? '#22c55e' : '#ef4444';

    const icon = L.divIcon({
      className: '',
      html: `<div style="
        background:${color};
        width:14px; height:14px; border-radius:50%;
        border:2px solid white; box-shadow:0 0 3px rgba(0,0,0,0.4);
      "></div>`,
      iconSize: [14, 14],
    });

    const marker = L.marker([item.lat, item.lon], { icon }).addTo(map);

    const answers = item.records?.length ? item.records : item.ips;
    const answersText = answers?.length ? answers.join(', ') : 'Cevap yok';

    marker.bindPopup(`
      <b>${item.city || 'Bilinmiyor'}, ${item.country || ''}</b><br>
      <b>IP:</b> ${item.ip}<br>
      <b>Tip:</b> ${item.type || 'A'}<br>
      <b>Sonuç:</b> ${isSuccess ? '✅ Başarılı' : '❌ Yanıtsız'}<br>
      <b>Cevap:</b> ${answersText}
    `);
    markers.push(marker);
  });
}

onMounted(() => {
  map = L.map(mapContainer.value).setView([20, 0], 2);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap',
  }).addTo(map);
  renderMarkers(props.results);
});

watch(() => props.results, (newResults) => renderMarkers(newResults));
</script>

<style scoped>
.dns-map { height: 500px; width: 1400px; border-radius: 8px; overflow: hidden; margin-left: 100px; }
</style>