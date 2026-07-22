import axios from 'axios';

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3001';

export async function checkDns(domain, type = 'A') {
  const { data } = await axios.get(`${API_BASE}/api/dnschecker`, {
    params: { domain, type },
  });
  return data;
}

export async function getFullRecords(domain) {
  const { data } = await axios.get(`${API_BASE}/api/dnschecker/full`, {
    params: { domain },
  });
  return data;
}