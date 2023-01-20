declare interface PageResult<T> {
    list: T[]
    total: number
}

declare interface Version {
    hash: string
    size: number
    sequence: number
    ts: number
    storeStrategy: number
    dataShards: number
    parityShards: number
    shardSize: number
    locate: string[]
}

declare interface Bucket {
    name: string
    readonly: boolean
    versioning: boolean
    versionRemains: number
    compress: boolean
    storeStrategy: number
    dataShards: number
    parityShards: number
    createTime: number
    updateTime: number
    policies: string[]
}

declare interface Metadata {
    name: string
    createTime: number
    updateTime: number
}

declare interface DiskInfo {
    total: number
    free: number
    used: number
}

declare interface MemStat {
    all: number
    used: number
    free: number
    self: number
}

declare interface CpuStat {
    usedPercent: number
    logicalCount: number
    physicalCount: number
}

declare interface SystemInfo {
    diskInfo: DiskInfo
    memStatus: MemStat
    cpuStatus: CpuStat
}

declare interface TimeStat {
    time: string
    percent: number
}

declare interface ServerInfo {
    serverId: string
    httpAddr: string
    rpcAddr: string
    isMaster?: boolean
    sysInfo: SystemInfo
}

declare interface SlotsInfo {
    id: string
    serverId: string
    location: string
    checksum: string
    slots: string[]
}

declare interface SlotRange {
    start: number
    end: number
    identify: string
}