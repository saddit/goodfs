import axios from "axios";

async function metadataPage(req: MetadataReq): Promise<PageResult<Metadata>> {
    let resp = await axios.get("/metadata/page", {
        params: req
    })
    let total = resp.headers["X-Total-Count"] || "0"
    return {
        total: parseInt(total),
        list: resp.data
    }
}

async function versionPage(req: MetadataReq): Promise<PageResult<Version>> {
    let resp = await axios.get("/metadata/versions", {
        params: req
    })
    let total = resp.headers["X-Total-Count"] || "0"
    return {
        total: parseInt(total),
        list: resp.data
    }
}

async function slotsDetail(): Promise<{ [key: string]: SlotsInfo }> {
    let resp = await axios.get("/metadata/slots_detail")
    return resp.data
}

async function startMigrate(req: MetaMigrateReq): Promise<any> {
    await axios.post("/metadata/migration", req)
}

async function getPeers(servId: string): Promise<ServerInfo[]> {
    let resp = await axios.get(`/metadata/peers?serverId=${servId}`)
    return resp.data
}

async function joinLeader(leaderId: string, servId: string) {
    await axios.post("/metadata/join_leader", {
        masterId: leaderId,
        serverId: servId
    })
}

async function leaveCluster(servId: string) {
    await axios.post("/metadata/leave_cluster", {
        serverId: servId
    })
}

export {
    metadataPage,
    versionPage,
    slotsDetail,
    startMigrate,
    getPeers,
    joinLeader,
    leaveCluster
}