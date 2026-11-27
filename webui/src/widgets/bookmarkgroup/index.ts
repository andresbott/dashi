import { defineAsyncComponent } from 'vue'
import type { WidgetModule } from '@/widgets/types'

const bookmarkgroup: WidgetModule = {
    type: 'bookmarkgroup',
    component: defineAsyncComponent(() => import('./Widget.vue')),
    configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
    label: 'Bookmark group',
    icon: 'ti-bookmarks',
    description: 'Several links sharing one layout',
    category: 'navigation',
}

export default bookmarkgroup
