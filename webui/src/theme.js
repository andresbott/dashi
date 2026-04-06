import { definePreset } from '@primevue/themes'
import Lara from '@primevue/themes/lara'

const CustomTheme = definePreset(Lara, {
    semantic: {
        primary: {
            50: '#e8eef7',
            100: '#c5d4ea',
            200: '#9fb8dc',
            300: '#799cce',
            400: '#5c87c3',
            500: '#3465a4',
            600: '#2e5a93',
            700: '#274b7b',
            800: '#203d64',
            900: '#192f4d',
            950: '#0f1d30'
        }
    }
})

export default CustomTheme
