import axios from "axios";
import {ApiError} from "@/api/error";

axios.defaults.baseURL = window.baseUrl
axios.defaults.timeout = 2500

// Add a request interceptor
axios.interceptors.request.use(function (config) {
    // Do something before request is sent
    config.headers.Authorization = `Basic ${useStore().basicAuth}`
    return config;
}, function (error) {
    // Do something with request error
    return Promise.reject(error);
});

// Add a response interceptor
axios.interceptors.response.use(function (response) {
    // Any status code that lie within the range of 2xx cause this function to trigger
    // Do something with response data
    return response;
}, function (error) {
    // Any status codes that falls outside the range of 2xx cause this function to trigger
    // Do something with response error
    if (axios.isAxiosError(error)) {
        console.error(error)
        if (!error.response) {
            return Promise.reject(new ApiError(error.status, "Network Error"))
        }
        return Promise.reject(new ApiError(error.status, error.response.data))
    }
    return Promise.reject(error);
});

export function getBaseUrl() {
    return axios.defaults.baseURL
}