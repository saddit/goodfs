import axios from "axios";

async function stat(): Promise<Map<string, ServerInfo>> {
    let resp = await axios.get("/server/stat")
    return resp.data
}

async function timeline(serv: number, type: string): Promise<Map<string, TimeStat>> {
    let resp = await axios.get(`/server/${type}/timeline?server=${serv}`)
    return resp.data
}

export {
    stat,
    timeline
}