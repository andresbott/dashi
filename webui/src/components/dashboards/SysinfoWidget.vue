<script setup lang="ts">
import { computed } from 'vue'
import { useSysinfo } from '@/composables/useSysinfo'
import type { Widget } from '@/types/dashboard'
import type { SysinfoWidgetConfig, DiskInfo, CPUTemp } from '@/types/sysinfo'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<SysinfoWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    return props.widget.config as unknown as SysinfoWidgetConfig
})

const { data: sysinfo, isLoading, isError } = useSysinfo()

const enabledDisks = computed<DiskInfo[]>(() => {
    if (!sysinfo.value || !config.value?.disks?.length) return []
    const enabled = new Set(config.value.disks)
    return sysinfo.value.disks.filter(d => enabled.has(d.mountpoint))
})

const enabledTemps = computed<CPUTemp[]>(() => {
    if (!sysinfo.value || !config.value?.tempSensors?.length) return []
    const enabled = new Set(config.value.tempSensors)
    return (sysinfo.value.cpuTemps ?? []).filter(t => enabled.has(t.label))
})

const hasContent = computed(() => {
    if (!config.value) return false
    const c = config.value
    return c.showMemory || c.showSwap || c.showUptime || c.showCpu
        || (c.tempSensors?.length ?? 0) > 0
        || c.showNetwork || c.showHostInfo || c.showProcesses
        || (c.disks?.length ?? 0) > 0
})

function humanBytes(b: number): string {
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let i = 0
    let val = b
    while (val >= 1024 && i < units.length - 1) {
        val /= 1024
        i++
    }
    return `${val.toFixed(1)} ${units[i]}`
}

function humanUptime(seconds: number): string {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    if (days > 0) return `${days}d ${hours}h ${minutes}m`
    if (hours > 0) return `${hours}h ${minutes}m`
    return `${minutes}m`
}

function barColor(pct: number): string {
    if (pct >= 90) return 'var(--p-red-500)'
    if (pct >= 75) return 'var(--p-orange-500)'
    return 'var(--p-green-500)'
}
</script>

<template>
    <div class="sysinfo-widget">
        <div v-if="!hasContent" class="sysinfo-empty">
            <i class="ti ti-server-cog" />
            <span>Configure in edit mode</span>
        </div>

        <div v-else-if="isLoading" class="sysinfo-empty">
            <i class="ti ti-loader-2 sysinfo-spinner" />
        </div>

        <div v-else-if="isError" class="sysinfo-empty">
            <i class="ti ti-server-off" />
            <span>Failed to load system info</span>
        </div>

        <div v-else-if="sysinfo" class="sysinfo-content" :class="{ 'sysinfo-horizontal': config?.horizontal }">
            <div v-for="disk in enabledDisks" :key="disk.mountpoint" class="sysinfo-item">
                <i class="ti ti-device-floppy sysinfo-icon" />
                <div class="sysinfo-detail">
                    <div class="sysinfo-header">
                        <span class="sysinfo-label">{{ disk.mountpoint }}</span>
                        <span class="sysinfo-value">{{ humanBytes(disk.free) }} free / {{ humanBytes(disk.total) }}</span>
                    </div>
                    <div class="sysinfo-bar">
                        <div
                            class="sysinfo-bar-fill"
                            :style="{ width: disk.usedPct.toFixed(0) + '%', backgroundColor: barColor(disk.usedPct) }"
                        />
                    </div>
                </div>
            </div>

            <div v-if="config?.showCpu" class="sysinfo-item">
                <i class="ti ti-cpu sysinfo-icon" />
                <div class="sysinfo-detail">
                    <div class="sysinfo-header">
                        <span class="sysinfo-label">CPU</span>
                        <span class="sysinfo-value">{{ sysinfo.cpuUsagePct.toFixed(0) }}% ({{ sysinfo.cpuCores }} cores)</span>
                    </div>
                    <div class="sysinfo-bar">
                        <div
                            class="sysinfo-bar-fill"
                            :style="{ width: sysinfo.cpuUsagePct.toFixed(0) + '%', backgroundColor: barColor(sysinfo.cpuUsagePct) }"
                        />
                    </div>
                </div>
            </div>

            <div v-if="config?.showMemory" class="sysinfo-item">
                <i class="ti ti-components sysinfo-icon" />
                <div class="sysinfo-detail">
                    <div class="sysinfo-header">
                        <span class="sysinfo-label">Memory</span>
                        <span class="sysinfo-value">{{ humanBytes(sysinfo.memUsed) }} / {{ humanBytes(sysinfo.memTotal) }}</span>
                    </div>
                    <div class="sysinfo-bar">
                        <div
                            class="sysinfo-bar-fill"
                            :style="{ width: sysinfo.memUsedPct.toFixed(0) + '%', backgroundColor: barColor(sysinfo.memUsedPct) }"
                        />
                    </div>
                </div>
            </div>

            <div v-if="config?.showSwap && sysinfo.swapTotal > 0" class="sysinfo-item">
                <i class="ti ti-switch-horizontal sysinfo-icon" />
                <div class="sysinfo-detail">
                    <div class="sysinfo-header">
                        <span class="sysinfo-label">Swap</span>
                        <span class="sysinfo-value">{{ humanBytes(sysinfo.swapUsed) }} / {{ humanBytes(sysinfo.swapTotal) }}</span>
                    </div>
                    <div class="sysinfo-bar">
                        <div
                            class="sysinfo-bar-fill"
                            :style="{ width: sysinfo.swapUsedPct.toFixed(0) + '%', backgroundColor: barColor(sysinfo.swapUsedPct) }"
                        />
                    </div>
                </div>
            </div>

            <div v-for="temp in enabledTemps" :key="temp.label" class="sysinfo-item sysinfo-text-item">
                <i class="ti ti-temperature sysinfo-icon" />
                <div class="sysinfo-detail">
                    <span class="sysinfo-label">{{ temp.label }}</span>
                    <span class="sysinfo-value">{{ temp.temperature.toFixed(0) }}°C</span>
                </div>
            </div>

            <div v-if="config?.showNetwork && sysinfo.netInterfaces?.length" class="sysinfo-item sysinfo-text-item">
                <i class="ti ti-network sysinfo-icon" />
                <div class="sysinfo-detail">
                    <span class="sysinfo-label">Network</span>
                    <span v-for="iface in sysinfo.netInterfaces" :key="iface.name" class="sysinfo-value">
                        {{ iface.name }}: ↑{{ humanBytes(iface.bytesSent) }} ↓{{ humanBytes(iface.bytesRecv) }}
                    </span>
                </div>
            </div>

            <div v-if="config?.showUptime" class="sysinfo-item sysinfo-text-item">
                <i class="ti ti-clock-up sysinfo-icon" />
                <div class="sysinfo-detail">
                    <span class="sysinfo-label">Uptime</span>
                    <span class="sysinfo-value">{{ humanUptime(sysinfo.uptimeSeconds) }}</span>
                </div>
            </div>

            <div v-if="config?.showProcesses" class="sysinfo-item sysinfo-text-item">
                <i class="ti ti-list-numbers sysinfo-icon" />
                <div class="sysinfo-detail">
                    <span class="sysinfo-label">Processes</span>
                    <span class="sysinfo-value">{{ sysinfo.numProcesses }}</span>
                </div>
            </div>

            <div v-if="config?.showHostInfo" class="sysinfo-item sysinfo-text-item">
                <i class="ti ti-server sysinfo-icon" />
                <div class="sysinfo-detail">
                    <span class="sysinfo-label">{{ sysinfo.hostname }}</span>
                    <span class="sysinfo-value">{{ sysinfo.os }}</span>
                    <span v-if="sysinfo.cpuModel" class="sysinfo-value">{{ sysinfo.cpuModel }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.sysinfo-widget {
    padding: 0.5rem;
    min-height: 60px;
}

.sysinfo-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.sysinfo-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.sysinfo-horizontal {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 1rem;
}

.sysinfo-horizontal .sysinfo-item {
    flex: 1;
    min-width: 120px;
}

.sysinfo-horizontal .sysinfo-icon {
    font-size: 1.5rem;
}

.sysinfo-item {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 0.375rem;
}

.sysinfo-icon {
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
    flex-shrink: 0;
}

.sysinfo-detail {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    flex: 1;
    min-width: 0;
}

.sysinfo-header {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.8rem;
}

.sysinfo-label {
    font-weight: 600;
    font-size: 0.8rem;
}

.sysinfo-value {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
}

.sysinfo-bar {
    height: 6px;
    background: var(--p-surface-200);
    border-radius: 3px;
    overflow: hidden;
}

.sysinfo-bar-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s ease;
}

.sysinfo-text-item .sysinfo-detail {
    flex-direction: row;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.375rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.sysinfo-spinner {
    animation: spin 1s linear infinite;
}
</style>
