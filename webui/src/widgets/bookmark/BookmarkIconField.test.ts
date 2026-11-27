import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import Select from 'primevue/select'
import BookmarkIconField from './BookmarkIconField.vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'

vi.mock('@/composables/useDashboards', () => ({
    useDashboardAssets: () => ({ data: ref([]) }),
}))

function sourceOptions(icon: string, dashboardId?: string) {
    const wrapper = mount(BookmarkIconField, {
        props: { modelValue: icon },
        global: {
            plugins: [PrimeVue],
            stubs: { IconSelector: true },
            provide: dashboardId === undefined ? {} : { [DASHBOARD_ID as symbol]: ref(dashboardId) },
        },
    })
    const options = wrapper.findAllComponents(Select)[0].props('options') as { value: string }[]
    return options.map(o => o.value)
}

describe('BookmarkIconField', () => {
    beforeEach(() => {
        vi.stubGlobal('matchMedia', () => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() }))
    })
    afterEach(() => vi.unstubAllGlobals())

    it('offers dashboard assets inside a dashboard', () => {
        expect(sourceOptions('ti-bookmark', 'dash1')).toEqual(['tabler', 'selfhst', 'dashboard'])
    })

    it('hides dashboard assets outside a dashboard', () => {
        expect(sourceOptions('ti-bookmark')).toEqual(['tabler', 'selfhst'])
    })

    it('keeps the option for a value that already uses it', () => {
        expect(sourceOptions('dashboard:logo.png')).toEqual(['tabler', 'selfhst', 'dashboard'])
    })
})
