import {defineStore} from "pinia";
import {Base64} from "js-base64";

export const useStore = defineStore('default', () => {
    const basicAuth = ref("")
    const locale = ref("en")
    const selectedServerTab = ref(0)
    const tabClosed = ref(false)
    const user = ref("Anonymous")
    const serverInfo = ref<ServerStatResp>({apiServer: {}, metaServer: {}, dataServer: {}})

    watch(locale, (lang) => {
        localStorage.setItem("locale", lang)
    })

    function clearAuth() {
        basicAuth.value = ""
        user.value = ""
    }

    function setAuth(username: string, password: string) {
        user.value = username
        if (username == "" && password == "") {
            basicAuth.value = ""
            return
        }
        basicAuth.value = Base64.encode(`${username}:${password}`)
    }

    function setLocale(lang: string) {
        locale.value = lang
    }

    function setServerInfo(data: ServerStatResp) {
        serverInfo.value = data
    }

    function closeTab() {
        tabClosed.value = !tabClosed.value
    }

    function setSelectedServerTab(n: number) {
        selectedServerTab.value = n
    }

    return {
        basicAuth,
        locale,
        tabClosed,
        serverStat: serverInfo,
        user,
        selectedServerTab,
        setSelectedServerTab,
        setServerInfo,
        setAuth,
        setLocale,
        closeTab,
        clearAuth
    }
}, {
    persist: {
        storage: localStorage,
        paths: ['locale', 'tabClosed', 'user', 'basicAuth', 'selectedServerTab']
    }
})