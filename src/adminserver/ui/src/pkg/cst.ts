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