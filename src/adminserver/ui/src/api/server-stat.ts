import axios from "axios";

async function stat(): Promise<ServerStatResp> {
    let resp = await axios.get("/server/stat")
    return resp.data
}

async function timeline(serv: number, type: string): Promise<Record<string, TimeStat[]>> {
    let resp = await axios.get(`/server/${type}/timeline?server=${serv}`)
    return resp.data
}

async function overview(): Promise<any> {
    let resp = await axios.get('/server/overview')
    return resp.data
}

async function etcdStat(): Promise<EtcdStatus[]> {
    let resp = await axios.get('/server/etcdstat')
    return resp.data
}

async function config(serverId: string): Promise<string> {
    let resp = await axios.get(`/server/config?serverId=${serverId}`)
    return resp.data
}

export {
    stat,
    timeline,
    overview,
    etcdStat,
    config,
}