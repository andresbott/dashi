import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const image: WidgetModule = {
    type: 'image',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Image',
    icon: 'ti-photo',
    description: 'Display an uploaded image',
}

export default image
