<script setup lang="ts">
import { computed } from 'vue'
import { useMarket } from '@/composables/useMarket'
import type { MarketWidgetConfig } from '@/types/market'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{
    widget: Widget
}>()

const config = computed<MarketWidgetConfig | null>(() => {
    if (!props.widget.config) return null
    return props.widget.config as unknown as MarketWidgetConfig
})

const symbol = computed(() => config.value?.symbol)
const rangeId = computed(() => config.value?.range ?? '1mo')
const showChart = computed(() => config.value?.showChart ?? true)
const showChange = computed(() => config.value?.showChange ?? true)

const symbolRef = computed(() => symbol.value)
const rangeRef = computed(() => rangeId.value)

const { data: market, isLoading, isError } = useMarket(symbolRef, rangeRef)

const rangeLabels: Record<string, string> = {
    '1d': '1 Day', '5d': '5 Days', '1mo': '1 Month',
    '3mo': '3 Months', '6mo': '6 Months', '1y': '1 Year',
}

const changeColor = computed(() => {
    if (!market.value) return undefined
    return market.value.quote.change >= 0 ? '#22c55e' : '#ef4444'
})

const changePrefix = computed(() => {
    if (!market.value) return ''
    return market.value.quote.change >= 0 ? '+' : ''
})

const formatPrice = (price: number) => price.toFixed(2)

const svgPath = computed(() => {
    if (!market.value || market.value.points.length < 2) return ''
    const points = market.value.points
    const closes = points.map(p => p.close)
    const min = Math.min(...closes)
    const max = Math.max(...closes)
    const range = max - min || 1
    const padding = range * 0.1
    const pMin = min - padding
    const pRange = max - min + padding * 2

    const w = 400
    const h = 120
    const coords = points.map((p, i) => ({
        x: (i / (points.length - 1)) * w,
        y: h - ((p.close - pMin) / pRange) * h,
    }))

    let d = `M ${coords[0].x},${coords[0].y}`
    for (let i = 1; i < coords.length; i++) {
        d += ` L ${coords[i].x},${coords[i].y}`
    }
    return d
})

const svgFill = computed(() => {
    if (!svgPath.value) return ''
    const w = 400
    const h = 120
    return `${svgPath.value} L ${w},${h} L 0,${h} Z`
})
</script>

<template>
    <div class="market-widget">
        <div v-if="!config?.symbol" class="market-empty">
            <i class="ti ti-chart-line" />
            <span>Set a symbol in edit mode</span>
        </div>

        <div v-else-if="isLoading" class="market-empty">
            <i class="ti ti-loader-2 market-spinner" />
        </div>

        <div v-else-if="isError" class="market-empty">
            <i class="ti ti-chart-off" />
            <span>Failed to load market data</span>
        </div>

        <div v-else-if="market">
            <div class="market-header">
                <span class="market-symbol">{{ market.quote.symbol }}</span>
                <span class="market-name">{{ market.quote.name }}</span>
            </div>
            <div class="market-price-row">
                <span class="market-price">{{ formatPrice(market.quote.price) }}</span>
                <span v-if="showChange" class="market-change" :style="{ color: changeColor }">
                    {{ changePrefix }}{{ formatPrice(market.quote.change) }}
                    ({{ changePrefix }}{{ market.quote.changePercent.toFixed(2) }}%)
                </span>
            </div>
            <div v-if="showChart && svgPath" class="market-chart">
                <svg viewBox="0 0 400 120" preserveAspectRatio="none" class="market-svg">
                    <defs>
                        <linearGradient :id="'mg-' + widget.id" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" :stop-color="changeColor" stop-opacity="0.3" />
                            <stop offset="100%" :stop-color="changeColor" stop-opacity="0" />
                        </linearGradient>
                    </defs>
                    <path :d="svgFill" :fill="'url(#mg-' + widget.id + ')'" />
                    <path :d="svgPath" fill="none" :stroke="changeColor" stroke-width="2" vector-effect="non-scaling-stroke" />
                </svg>
            </div>
            <div class="market-footer">
                <span class="market-currency">{{ market.quote.currency }}</span>
                <span class="market-range">{{ rangeLabels[rangeId] }}</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.market-widget {
    padding: 0.5rem;
    min-height: 60px;
}

.market-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60px;
    gap: 0.5rem;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.market-empty .ti {
    font-size: 1.5rem;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.market-spinner {
    animation: spin 1s linear infinite;
}

.market-header {
    margin-bottom: 0.25rem;
}

.market-symbol {
    font-weight: 700;
    font-size: 1.1em;
    margin-right: 0.5rem;
}

.market-name {
    font-size: 0.8em;
    color: var(--p-text-muted-color);
}

.market-price-row {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
}

.market-price {
    font-size: 1.5em;
    font-weight: 700;
}

.market-change {
    font-size: 0.9em;
    font-weight: 600;
}

.market-chart {
    margin: 0.25rem 0;
}

.market-svg {
    width: 100%;
    height: 80px;
}

.market-footer {
    display: flex;
    justify-content: space-between;
    font-size: 0.75em;
    color: var(--p-text-muted-color);
}
</style>
