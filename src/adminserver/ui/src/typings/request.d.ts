declare type OrderType = "create_time" | "updated_time" | "name"

declare interface Pageable {
    page: number
    pageSize: number
    OrderBy?: OrderType
    Desc?: boolean
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