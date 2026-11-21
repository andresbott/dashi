import pluginVue from 'eslint-plugin-vue'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'

// ESLint 9 flat config. Prettier owns formatting (skipFormatting turns the
// stylistic rules off), so this only reports real problems.
export default defineConfigWithVueTs(
    {
        name: 'app/files-to-lint',
        files: ['**/*.{js,mjs,ts,mts,tsx,vue}'],
    },
    {
        name: 'app/files-to-ignore',
        ignores: ['dist/**', 'coverage/**', 'node_modules/**', 'viewer/**'],
    },
    pluginVue.configs['flat/essential'],
    vueTsConfigs.recommended,
    skipFormatting,
    {
        name: 'app/rules',
        rules: {
            // Single-word component names are the project's convention for
            // widget folders (Widget.vue, WidgetConfig.vue).
            'vue/multi-word-component-names': 'off',
            // Most SFCs are TypeScript, but a few admin views are still plain
            // JS `<script setup>`; they are linted as JS rather than forced to migrate.
            'vue/block-lang': ['error', { script: { lang: 'ts', allowNoLang: true } }],
            // `_`-prefixed and rest-sibling unused bindings are intentional.
            '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_', ignoreRestSiblings: true }],
        },
    },
)
