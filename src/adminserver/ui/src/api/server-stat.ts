import axios from "axios";

async function stat(): Promise<Map<string, ServerInfo>> {
    let resp = await axios.get("/server/stat")
    return resp.data
}

export {
    stat
}