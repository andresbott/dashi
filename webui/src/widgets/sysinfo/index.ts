import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const sysinfo: WidgetModule = {
    type: 'sysinfo',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'System Info',
    icon: 'ti-server-cog',
    description: 'Host disk, memory, and uptime',
}

export default sysinfo
