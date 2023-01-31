import axios from "axios";

async function addBucket(b: Bucket) {
    await axios.post("/metadata/create_bucket", b)
}

async function updateBucket(b: Bucket) {
    await axios.put("/metadata/update_bucket", b)
}

async function removeBucket(name: string) {
    await axios.delete(`/metadata/delete_bucket?name=${name}`)
}

async function bucketPage(req: BucketReq): Promise<PageResult<Bucket>> {
    let resp = await axios.get("/metadata/buckets", {
        params: req
    })
    let total = resp.headers["x-total-count"] || "0"
    return {
        total: parseInt(total),
        list: resp.data
    }
}

async function metadataPage(req: MetadataReq): Promise<PageResult<Metadata>> {
    let resp = await axios.get("/metadata/page", {
        params: req
    })
    let total = resp.headers["x-total-count"] || "0"
    return {
        total: parseInt(total),
        list: resp.data
    }
}

async function versionPage(req: MetadataReq): Promise<PageResult<Version>> {
    let resp = await axios.get("/metadata/versions", {
        params: req
    })
    let total = resp.headers["x-total-count"] || "0"
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
    bucketPage,
    addBucket,
    updateBucket,
    removeBucket,
    metadataPage,
    versionPage,
    slotsDetail,
    startMigrate,
    getPeers,
    joinLeader,
    leaveCluster
}