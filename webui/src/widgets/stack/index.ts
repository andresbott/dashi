import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const stack: WidgetModule = {
    type: 'stack',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: null,
    label: 'Stack',
    icon: 'ti-layout-rows',
    description: 'Stack widgets vertically in a column',
}

export default stack
