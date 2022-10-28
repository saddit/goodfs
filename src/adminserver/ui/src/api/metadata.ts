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

export {
    metadataPage,
    versionPage
}