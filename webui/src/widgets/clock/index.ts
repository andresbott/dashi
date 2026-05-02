import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const clock: WidgetModule = {
    type: 'clock',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Clock',
    icon: 'ti-clock',
    description: 'Digital clock with date',
}

export default clock
