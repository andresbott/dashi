import type { DocQueryMock } from '@/lib/widgetDocs'
import type { SystemInfo } from './types'

const GiB = 1024 ** 3

// A canned home server for the widget docs. The settings form lists these
// disks and sensors, so ticking them shows them in the preview.
export const sampleSysinfo: SystemInfo = {
    disks: [
        { mountpoint: '/', device: '/dev/nvme0n1p2', total: 476 * GiB, free: 201 * GiB, usedPct: 57.8, fstype: 'ext4' },
        { mountpoint: '/srv/media', device: '/dev/sda1', total: 3726 * GiB, free: 410 * GiB, usedPct: 89.0, fstype: 'xfs' },
        { mountpoint: '/boot', device: '/dev/nvme0n1p1', total: 0.5 * GiB, free: 0.04 * GiB, usedPct: 92.0, fstype: 'vfat' },
    ],
    memTotal: 32 * GiB,
    memUsed: 11.6 * GiB,
    memUsedPct: 36.3,
    swapTotal: 8 * GiB,
    swapUsed: 0.6 * GiB,
    swapUsedPct: 7.5,
    uptimeSeconds: 12 * 86400 + 5 * 3600 + 17 * 60,
    cpuUsagePct: 23,
    cpuModel: 'AMD Ryzen 7 5700G',
    cpuCores: 16,
    cpuTemps: [
        { label: 'Tctl', temperature: 48 },
        { label: 'nvme Composite', temperature: 39 },
    ],
    netInterfaces: [{ name: 'eth0', bytesSent: 182 * GiB, bytesRecv: 1290 * GiB }],
    hostname: 'homeserver',
    os: 'Debian GNU/Linux 13',
    kernelVersion: '6.12.0',
    loadAvg1: 0.42,
    loadAvg5: 0.38,
    loadAvg15: 0.31,
    numProcesses: 312,
}

export const sysinfoMocks: Record<string, DocQueryMock> = {
    sysinfo: () => sampleSysinfo,
}
