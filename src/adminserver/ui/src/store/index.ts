import {defineStore} from "pinia";
import {Base64} from "js-base64";

export const useStore = defineStore('default', () => {
    const basicAuth = ref("")
    const locale = ref("en")
    const tabClosed = ref(false)
    const user = ref("Anonymous")

    watch(locale, (lang)=>{
        localStorage.setItem("locale", lang)
    })

    function setAuth(username: string, password: string) {
        user.value = username
        basicAuth.value = Base64.encode(`${username}:${password}`)
    }

    function setLocale(lang: string) {
        locale.value = lang
    }

    function closeTab() {
        tabClosed.value = !tabClosed.value
    }

    return {basicAuth, locale, tabClosed, user, setAuth, setLocale, closeTab}
}, {
    persist: {
        storage: localStorage,
        paths: ['locale', 'tabClosed', 'user', 'basicAuth']
    }
})