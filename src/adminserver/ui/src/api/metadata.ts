import axios from "axios";

async function metadataPage(req: MetadataReq): Promise<Version[]> {
    let resp = await axios.get("/metadata/page", {
        params: req
    })
    return resp.data
}

async function versionPage(req: MetadataReq): Promise<Metadata[]> {
    let resp = await axios.get("/metadata/versions", {
        params: req
    })
    return resp.data
}

async function slotsDetail(): Promise<{ [key: string]: SlotsInfo }> {
    let resp = await axios.get("/metadata/slots_detail")
    return resp.data
}

async function startMigrate(req: MetaMigrateReq): Promise<any> {
    await axios.post("/metadata/migration", req)
}

export {
    metadataPage,
    versionPage,
    slotsDetail,
    startMigrate
}