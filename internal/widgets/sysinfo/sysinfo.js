import { poll } from '/_dashi/assets/dashi.js'

// Mirrors humanBytes in image.go. Both sides must agree: Go prefills the
// value and JS overwrites it 30 seconds later.
export function humanBytes(b) {
    const kb = 1024, mb = 1024 * kb, gb = 1024 * mb, tb = 1024 * gb
    if (b >= tb) return `${(b / tb).toFixed(1)} TB`
    if (b >= gb) return `${(b / gb).toFixed(1)} GB`
    if (b >= mb) return `${(b / mb).toFixed(1)} MB`
    if (b >= kb) return `${(b / kb).toFixed(1)} KB`
    return `${b} B`
}

// Mirrors humanUptime in image.go.
export function humanUptime(seconds) {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    if (days > 0) return `${days}d ${hours}h ${minutes}m`
    if (hours > 0) return `${hours}h ${minutes}m`
    return `${minutes}m`
}

poll('.dashi-sysinfo', {
    url: '/api/v0/widgets/sysinfo',
    every: 30,
    apply: (el, info, set) => {
        for (const item of el.querySelectorAll('[data-disk]')) {
            const disk = (info.disks || []).find(d => d.mountpoint === item.dataset.disk)
            if (!disk) continue
            const fill = item.querySelector('.dashi-sysinfo__bar-fill')
            if (fill) fill.style.width = `${Math.round(disk.usedPct)}%`
            const value = item.querySelector('.dashi-sysinfo__value')
            if (value) value.textContent = `${humanBytes(disk.free)} free / ${humanBytes(disk.total)}`
        }
        set('[data-mem] .dashi-sysinfo__bar-fill', { style: { width: `${Math.round(info.memUsedPct)}%` } })
        set('[data-mem] .dashi-sysinfo__value', `${humanBytes(info.memUsed)} / ${humanBytes(info.memTotal)}`)
        set('[data-uptime] .dashi-sysinfo__value', humanUptime(info.uptimeSeconds))
    },
})
