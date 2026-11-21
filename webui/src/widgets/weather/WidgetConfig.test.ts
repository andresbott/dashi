import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import Checkbox from 'primevue/checkbox'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import WidgetConfig from './WidgetConfig.vue'
import type { WeatherWidgetConfig } from './types'

function mountForm(config: Partial<WeatherWidgetConfig>) {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    return mount(WidgetConfig, {
        props: { config: { city: 'Zurich, CH', latitude: 47.37, longitude: 8.54, ...config } as WeatherWidgetConfig },
        global: { plugins: [PrimeVue, [VueQueryPlugin, { queryClient }]] },
    })
}

describe('weather WidgetConfig', () => {
    it('lays the form out as three sections: location + display, extra info, forecast + graph', () => {
        const wrapper = mountForm({ showHourly: true, showForecast: true, showGraph: true })
        const sections = wrapper.findAll('section').map(s => s.attributes('data-testid'))
        expect(sections).toEqual(['weather-config-location', 'weather-config-extra', 'weather-config-forecast'])

        const location = wrapper.find('[data-testid="weather-config-location"]').text()
        expect(location).toContain('Search city')
        expect(location).toContain('Zurich, CH')
        expect(location).toContain('Show current weather')

        const extra = wrapper.find('[data-testid="weather-config-extra"] .weather-config-extra-grid')
        expect(extra.findAllComponents(Checkbox)).toHaveLength(8)

        const forecast = wrapper.find('[data-testid="weather-config-forecast"]').text()
        expect(forecast).toContain('Show hourly forecast')
        expect(forecast).toContain('Show daily forecast')
        expect(forecast).toContain('Show forecast graph')
        expect(forecast).toContain('Temp color')
    })

    it('still emits the full config when a checkbox changes', async () => {
        const wrapper = mountForm({})
        await wrapper.find('#weatherUV').setValue(true)
        const emitted = wrapper.emitted('update:config')
        expect(emitted).toBeTruthy()
        expect(emitted!.at(-1)![0]).toMatchObject({ city: 'Zurich, CH', showUV: true, showCurrent: true })
    })
})
