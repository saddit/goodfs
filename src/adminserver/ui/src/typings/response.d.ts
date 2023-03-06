declare interface ErrResponse {
    message: string
    subMessage: string
}

declare interface ServerStatResp {
    apiServer: { [k: string]: ServerInfo }
    metaServer: { [k: string]: ServerInfo }
    dataServer: { [k: string]: ServerInfo }
}

declare interface StringAny extends Record<string, any> {}