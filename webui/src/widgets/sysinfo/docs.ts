import type { WidgetDocs } from '@/lib/widgetDocs'
import { sysinfoMocks } from './docFixture'

// Fields the server-rendered viewer does not read yet (widget_contract.json
// lists them as editorOnly).
const editorOnly = ' Editor preview only: dashboards do not show it yet.'

const docs: WidgetDocs = {
    summary: 'Shows the health of the machine Dashi runs on: disk usage, memory and uptime, read from the host by the server.',
    params: [
        { name: 'disks', type: 'string[]', description: 'Mount points to show as usage bars, e.g. `/`. The settings list the partitions the server can see.' },
        { name: 'showMemory', type: 'boolean', default: 'false', description: 'Show memory usage.' },
        { name: 'showUptime', type: 'boolean', default: 'false', description: 'Show how long the host has been up.' },
        { name: 'showCpu', type: 'boolean', default: 'false', description: 'Show CPU usage and core count.' + editorOnly },
        { name: 'showSwap', type: 'boolean', default: 'false', description: 'Show swap usage (hidden when the host has no swap).' + editorOnly },
        { name: 'showNetwork', type: 'boolean', default: 'false', description: 'Show bytes sent and received per network interface.' + editorOnly },
        { name: 'showHostInfo', type: 'boolean', default: 'false', description: 'Show hostname, OS and CPU model.' + editorOnly },
        { name: 'showProcesses', type: 'boolean', default: 'false', description: 'Show the number of running processes.' + editorOnly },
        { name: 'tempSensors', type: 'string[]', description: 'Temperature sensor labels to show, e.g. `Tctl`.' + editorOnly },
        { name: 'horizontal', type: 'boolean', default: 'false', description: 'Lay the items out side by side instead of stacked.' + editorOnly },
    ],
    notes: ['Values come from the host running Dashi and refresh every minute.'],
    preview: {
        title: 'System Info',
        config: {
            disks: ['/', '/srv/media'],
            showMemory: true,
            showUptime: true,
            showCpu: false,
            showSwap: false,
            showNetwork: false,
            showHostInfo: false,
            showProcesses: false,
            tempSensors: [],
            horizontal: false,
        },
        settingsNote: 'The preview and the partition list use a sample host.',
    },
    mocks: sysinfoMocks,
}

export default docs
