import type {PluginOptions } from "vue-toastification/dist/types/types";
import {POSITION} from "vue-toastification";

export const notify: PluginOptions = {
    position: POSITION.TOP_RIGHT,
    timeout: 2000,
    closeOnClick: true,
    pauseOnHover: true,
    draggable: false,
    showCloseButtonOnHover: false,
    hideProgressBar: true,
    closeButton: "button",
    icon: true,
    rtl: false,
    transition: "Vue-Toastification__bounce",
    maxToasts: 20,
    newestOnTop: true
}

export const statTypeMem = "mem"
export const statTypeCpu = "cpu"
export const apiServerNo = 0
export const metaServerNo = 1
export const dataServerNo = 2