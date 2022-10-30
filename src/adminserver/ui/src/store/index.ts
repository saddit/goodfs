import {defineStore} from "pinia";
import {Base64} from "js-base64";

export const useStore = defineStore('default', () => {
    const basicAuth = ref("")
    const locale = ref("en")

    watch(locale, (lang)=>{
        localStorage.setItem("locale", lang)
    })

    function setAuth(username: string, password: string) {
        basicAuth.value = Base64.encode(`${username}:${password}`)
    }

    function setLocale(lang: string) {
        locale.value = lang
    }

    return {basicAuth, locale, setAuth, setLocale}
}, {
    persist: {
        storage: localStorage,
        paths: ['locale']
    }
})