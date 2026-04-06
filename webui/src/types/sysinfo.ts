export interface DiskInfo {
    mountpoint: string
    device: string
    total: number
    free: number
    usedPct: number
    fstype: string
}

export interface NetInterface {
    name: string
    bytesSent: number
    bytesRecv: number
}

export interface CPUTemp {
    label: string
    temperature: number
}

export interface SystemInfo {
    disks: DiskInfo[]
    memTotal: number
    memUsed: number
    memUsedPct: number
    swapTotal: number
    swapUsed: number
    swapUsedPct: number
    uptimeSeconds: number
    cpuUsagePct: number
    cpuModel: string
    cpuCores: number
    cpuTemps: CPUTemp[]
    netInterfaces: NetInterface[]
    hostname: string
    os: string
    kernelVersion: string
    loadAvg1: number
    loadAvg5: number
    loadAvg15: number
    numProcesses: number
}

export interface SysinfoWidgetConfig {
    showMemory: boolean
    showSwap: boolean
    showUptime: boolean
    showCpu: boolean
    tempSensors: string[]
    showNetwork: boolean
    showHostInfo: boolean
    showProcesses: boolean
    horizontal: boolean
    disks: string[]
}
