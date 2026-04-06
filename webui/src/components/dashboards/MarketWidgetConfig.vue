<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import type { MarketWidgetConfig } from '@/types/market'

const props = defineProps<{
    config: MarketWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: MarketWidgetConfig]
}>()

const rangeOptions = [
    { label: '1 Day', value: '1d' },
    { label: '5 Days', value: '5d' },
    { label: '1 Month', value: '1mo' },
    { label: '3 Months', value: '3mo' },
    { label: '6 Months', value: '6mo' },
    { label: '1 Year', value: '1y' },
]

const editSymbol = ref(props.config?.symbol ?? '')
const editRange = ref(props.config?.range ?? '1mo')
const editShowChart = ref(props.config?.showChart ?? true)
const editShowChange = ref(props.config?.showChange ?? true)

watch(() => props.config, (val) => {
    if (val) {
        editSymbol.value = val.symbol ?? ''
        editRange.value = val.range ?? '1mo'
        editShowChart.value = val.showChart ?? true
        editShowChange.value = val.showChange ?? true
    }
})

const emitUpdate = () => {
    emit('update:config', {
        symbol: editSymbol.value,
        range: editRange.value,
        showChart: editShowChart.value,
        showChange: editShowChange.value,
    })
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Symbol</label>
            <InputText v-model="editSymbol" placeholder="e.g. AAPL, BTC-USD" @input="emitUpdate" />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Range</label>
            <Select
                v-model="editRange"
                :options="rangeOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                @change="emitUpdate"
            />
        </div>
        <div class="flex flex-column gap-2">
            <label class="text-sm font-semibold">Display</label>
            <div class="flex align-items-center gap-2">
                <Checkbox v-model="editShowChart" :binary="true" inputId="marketShowChart" @update:modelValue="emitUpdate" />
                <label for="marketShowChart" class="text-sm">Show chart</label>
            </div>
            <div class="flex align-items-center gap-2">
                <Checkbox v-model="editShowChange" :binary="true" inputId="marketShowChange" @update:modelValue="emitUpdate" />
                <label for="marketShowChange" class="text-sm">Show price change</label>
            </div>
        </div>
    </div>
</template>
