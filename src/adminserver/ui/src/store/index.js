import {defineStore} from "pinia";
import {Base64} from "js-base64";

export const useStore = defineStore('default', () => {
    const basicAuth = ref("")
    function setAuth(username, password) {
        basicAuth.value = Base64.encode(`${username}:${password}`)
    }

    return { basicAuth, setAuth }
})