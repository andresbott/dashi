import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const xkcd: WidgetModule = {
    type: 'xkcd',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'XKCD',
    icon: 'ti-pencil',
    description: 'XKCD comic strip',
}

export default xkcd
