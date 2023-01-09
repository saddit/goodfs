declare type OrderType = "create_time" | "update_time" | "name" | ""

declare interface Pageable {
    page: number
    pageSize: number
    total?: number
    orderBy?: OrderType
    desc?: boolean
}

declare interface MetadataReq extends Pageable {
    name: string
    version?: number
}

declare interface MetaMigrateReq {
    srcServerId: string
    destServerId: string
    slots: string[]
    slotsStr: string
}