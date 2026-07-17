<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <nav class="site-nav">
    <RouterLink to="/">Ana Sayfa</RouterLink>
    <RouterLink to="/subnetting">IP Subnetting</RouterLink>
    <RouterLink to="/portchecker">Port Checker</RouterLink>
    <RouterLink to="/ipscanner">IP Scanner</RouterLink>
    <RouterLink to="/sslcheck">SSL Checker</RouterLink>
    <RouterLink to="/wirehamsi">wirehamsi</RouterLink>
    <RouterLink to="/packet-sender">Packet Sender</RouterLink>
  </nav>

  <RouterView />
</template>

<style>
.site-nav { display: flex; min-height: 57px; align-items: center; gap: 6px; padding: 0 24px; overflow-x: auto; background: #0c1929; border-bottom: 1px solid #20394f; }
.site-nav a { padding: 7px 8px; border-radius: 5px; color: #9db3c2; font-size: 12px; text-decoration: none; white-space: nowrap; }
.site-nav a.router-link-active { background: #19364a; color: #e4f7fa; }.site-nav .wirehamsi-link { margin-left: auto; color: #49d7c4; font-weight: 800; }
</style>
