<script setup lang="ts">
import { ref, watch } from 'vue'
import Checkbox from 'primevue/checkbox'
import { useSysinfo } from '@/composables/useSysinfo'
import type { SysinfoWidgetConfig } from '@/types/sysinfo'

const props = defineProps<{
    config: SysinfoWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: SysinfoWidgetConfig]
}>()

const { data: sysinfo, isLoading } = useSysinfo()

const showMemory = ref(props.config?.showMemory ?? false)
const showSwap = ref(props.config?.showSwap ?? false)
const showUptime = ref(props.config?.showUptime ?? false)
const showCpu = ref(props.config?.showCpu ?? false)
const selectedTempSensors = ref<string[]>(props.config?.tempSensors ?? [])
const showNetwork = ref(props.config?.showNetwork ?? false)
const showHostInfo = ref(props.config?.showHostInfo ?? false)
const showProcesses = ref(props.config?.showProcesses ?? false)
const horizontal = ref(props.config?.horizontal ?? false)
const selectedDisks = ref<string[]>(props.config?.disks ?? [])

watch(() => props.config, (val) => {
    if (val) {
        showMemory.value = val.showMemory ?? false
        showSwap.value = val.showSwap ?? false
        showUptime.value = val.showUptime ?? false
        showCpu.value = val.showCpu ?? false
        selectedTempSensors.value = val.tempSensors ?? []
        showNetwork.value = val.showNetwork ?? false
        showHostInfo.value = val.showHostInfo ?? false
        showProcesses.value = val.showProcesses ?? false
        horizontal.value = val.horizontal ?? false
        selectedDisks.value = val.disks ?? []
    }
})

const emitUpdate = () => {
    emit('update:config', {
        showMemory: showMemory.value,
        showSwap: showSwap.value,
        showUptime: showUptime.value,
        showCpu: showCpu.value,
        tempSensors: selectedTempSensors.value,
        showNetwork: showNetwork.value,
        showHostInfo: showHostInfo.value,
        showProcesses: showProcesses.value,
        horizontal: horizontal.value,
        disks: selectedDisks.value,
    })
}

const toggleDisk = (mountpoint: string) => {
    const idx = selectedDisks.value.indexOf(mountpoint)
    if (idx >= 0) {
        selectedDisks.value.splice(idx, 1)
    } else {
        selectedDisks.value.push(mountpoint)
    }
    emitUpdate()
}

const isDiskSelected = (mountpoint: string): boolean => {
    return selectedDisks.value.includes(mountpoint)
}

const toggleTempSensor = (label: string) => {
    const idx = selectedTempSensors.value.indexOf(label)
    if (idx >= 0) {
        selectedTempSensors.value.splice(idx, 1)
    } else {
        selectedTempSensors.value.push(label)
    }
    emitUpdate()
}

const isTempSensorSelected = (label: string): boolean => {
    return selectedTempSensors.value.includes(label)
}
</script>

<template>
    <div class="sysinfo-config">
        <div class="flex flex-column gap-3">
            <label class="text-sm font-semibold">Layout</label>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="horizontal" :binary="true" inputId="sysHorizontal" @update:modelValue="emitUpdate" />
                <label for="sysHorizontal" class="text-sm">Horizontal layout</label>
            </div>
        </div>

        <div class="flex flex-column gap-3 mt-3">
            <label class="text-sm font-semibold">Health</label>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showCpu" :binary="true" inputId="sysCpu" @update:modelValue="emitUpdate" />
                <label for="sysCpu" class="text-sm">CPU usage</label>
            </div>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showMemory" :binary="true" inputId="sysMemory" @update:modelValue="emitUpdate" />
                <label for="sysMemory" class="text-sm">Memory usage</label>
            </div>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showSwap" :binary="true" inputId="sysSwap" @update:modelValue="emitUpdate" />
                <label for="sysSwap" class="text-sm">Swap usage</label>
            </div>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showNetwork" :binary="true" inputId="sysNetwork" @update:modelValue="emitUpdate" />
                <label for="sysNetwork" class="text-sm">Network throughput</label>
            </div>
        </div>

        <div class="flex flex-column gap-3 mt-3">
            <label class="text-sm font-semibold">Info</label>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showHostInfo" :binary="true" inputId="sysHostInfo" @update:modelValue="emitUpdate" />
                <label for="sysHostInfo" class="text-sm">Hostname / OS / CPU model</label>
            </div>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showUptime" :binary="true" inputId="sysUptime" @update:modelValue="emitUpdate" />
                <label for="sysUptime" class="text-sm">Uptime</label>
            </div>

            <div class="flex align-items-center gap-2">
                <Checkbox v-model="showProcesses" :binary="true" inputId="sysProcesses" @update:modelValue="emitUpdate" />
                <label for="sysProcesses" class="text-sm">Process count</label>
            </div>
        </div>

        <div class="flex flex-column gap-3 mt-3">
            <label class="text-sm font-semibold">Disk partitions</label>

            <div v-if="isLoading" class="text-sm" style="color: var(--p-text-muted-color)">
                <i class="ti ti-loader-2 sysinfo-config-spinner" /> Loading partitions...
            </div>

            <template v-else-if="sysinfo">
                <div v-for="disk in sysinfo.disks" :key="disk.mountpoint" class="flex align-items-center gap-2">
                    <Checkbox
                        :modelValue="isDiskSelected(disk.mountpoint)"
                        :binary="true"
                        :inputId="'disk-' + disk.mountpoint"
                        @update:modelValue="toggleDisk(disk.mountpoint)"
                    />
                    <label :for="'disk-' + disk.mountpoint" class="text-sm">
                        {{ disk.mountpoint }}
                        <span style="color: var(--p-text-muted-color)"> ({{ disk.device }}, {{ disk.fstype }})</span>
                    </label>
                </div>
            </template>

            <div v-else class="text-sm" style="color: var(--p-text-muted-color)">
                Could not load disk information
            </div>
        </div>

        <div class="flex flex-column gap-3 mt-3">
            <label class="text-sm font-semibold">Temperature sensors</label>

            <div v-if="isLoading" class="text-sm" style="color: var(--p-text-muted-color)">
                <i class="ti ti-loader-2 sysinfo-config-spinner" /> Loading sensors...
            </div>

            <template v-else-if="sysinfo?.cpuTemps?.length">
                <div v-for="temp in sysinfo.cpuTemps" :key="temp.label" class="flex align-items-center gap-2">
                    <Checkbox
                        :modelValue="isTempSensorSelected(temp.label)"
                        :binary="true"
                        :inputId="'temp-' + temp.label"
                        @update:modelValue="toggleTempSensor(temp.label)"
                    />
                    <label :for="'temp-' + temp.label" class="text-sm">
                        {{ temp.label }}
                        <span style="color: var(--p-text-muted-color)"> ({{ temp.temperature.toFixed(0) }}°C)</span>
                    </label>
                </div>
            </template>

            <div v-else class="text-sm" style="color: var(--p-text-muted-color)">
                No temperature sensors available
            </div>
        </div>
    </div>
</template>

<style scoped>
@keyframes spin {
    to { transform: rotate(360deg); }
}

.sysinfo-config-spinner {
    animation: spin 1s linear infinite;
}
</style>
