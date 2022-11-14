import "./config"; // before import 'vue', import 'config' to init variable from backend
import {createApp} from 'vue'
import './tailwind.css'
import App from './App.vue'
import {createRouter, createWebHistory} from 'vue-router/auto'
import {createHead} from '@vueuse/head'
import {createI18n} from "vue-i18n";
import {createPinia} from "pinia";
import piniaPluginPersistedState from 'pinia-plugin-persistedstate'
import messages from '@intlify/vite-plugin-vue-i18n/messages'
import Toast from "vue-toastification";
import "vue-toastification/dist/index.css";
import icon from './font-awesome'

const app = createApp(App)
const head = createHead()

const pinia = createPinia().use(piniaPluginPersistedState)

const router = createRouter({
    history: createWebHistory(),
})

const i18n = createI18n({
    locale: localStorage.getItem("locale") || "en",
    messages
})

// see src/typings/vue-cus.d.ts
app.config.globalProperties.$utils = pkg.utils
app.config.globalProperties.$cst = pkg.cst

app.use(Toast, pkg.cst.notify)
app.use(icon)
app.use(router)
app.use(head)
app.use(i18n)
app.use(pinia)
app.mount(document.body)
