import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const bookmark: WidgetModule = {
    type: 'bookmark',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Bookmark',
    icon: 'ti-bookmark',
    description: 'Link to an external website',
}

export default bookmark
