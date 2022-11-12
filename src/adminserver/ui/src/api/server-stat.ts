import axios from "axios";

async function stat(): Promise<ServerStatResp> {
    let resp = await axios.get("/server/stat")
    return resp.data
}

async function timeline(serv: number, type: string): Promise<{[key: string]: TimeStat[]}> {
    let resp = await axios.get(`/server/${type}/timeline?server=${serv}`)
    return resp.data
}

export {
    stat,
    timeline
}