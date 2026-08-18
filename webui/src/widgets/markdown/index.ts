import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const markdown: WidgetModule = {
    type: 'markdown',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Markdown',
    icon: 'ti-markdown',
    description: 'Render markdown content from a file',
}

export default markdown
