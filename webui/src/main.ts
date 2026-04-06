import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import { queryClient } from '@/lib/queryClient'
import App from './App.vue'
// @ts-expect-error theme.js has no type declarations
import CustomTheme from '@/theme.js'

import 'primeflex/primeflex.css'
import '@tabler/icons-webfont/dist/tabler-icons-300.min.css'
import '@/assets/style.scss'

import PrimeVue from 'primevue/config'

const app = createApp(App)

// Detect and apply system color scheme preference
const applyTheme = () => {
    const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const htmlElement = document.documentElement

    if (darkModeMediaQuery.matches) {
        htmlElement.classList.add('dark-mode')
    } else {
        htmlElement.classList.remove('dark-mode')
    }
}

applyTheme()
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme)

app.use(PrimeVue, {
    ripple: true,
    theme: {
        preset: CustomTheme,
        options: {
            darkModeSelector: '.dark-mode',
            cssLayer: false
        }
    }
})

// Tooltip directive
import Tooltip from 'primevue/tooltip'
app.directive('tooltip', Tooltip)

// Toast service
import ToastService from 'primevue/toastservice'
app.use(ToastService)

// Router
import router from './router'
app.use(router)

// Vue Query
app.use(VueQueryPlugin, {
    queryClient
})

app.mount('#app')
