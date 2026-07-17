import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../pages/HomePage.vue'
import SubnettingPage from '../pages/SubnettingPage.vue'
import PortCheckerPage from '../pages/PortCheckerPage.vue'
import IPScannerPage from '../pages/IPScannerPage.vue'
import SSLCheckPage from '@/pages/SSLCheckPage.vue'
import CapturePage from '@/pages/CapturePage.vue'
import CrafterPage from '@/pages/CrafterPage.vue'


const routes = [
  { path: '/', component: HomePage },
  { path: '/subnetting', component: SubnettingPage },
  { path: '/portchecker', component: PortCheckerPage },
  { path: '/ipscanner', component: IPScannerPage},
  { path: '/sslcheck', component: SSLCheckPage},
  { path: '/wirehamsi', component: CapturePage },
  { path: '/packet-sender', component: CrafterPage },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
