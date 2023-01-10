import axios from "axios";

async function login() {
    await axios.post("/login")
}

async function logout() {
    await axios.post("/logout")
}

export {
    login, logout
}