import { globalIgnores } from 'eslint/config'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import pluginVue from 'eslint-plugin-vue'
import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'

// To allow more languages other than `ts` in `.vue` files, uncomment the following lines:
// import { configureVueProject } from '@vue/eslint-config-typescript'
// configureVueProject({ scriptLangs: ['ts', 'tsx'] })
// More info at https://github.com/vuejs/eslint-config-typescript/#advanced-setup

export default defineConfigWithVueTs(
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,mts,tsx,vue}'],
  },

  globalIgnores(['**/dist/**', '**/dist-ssr/**', '**/coverage/**']),

  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  skipFormatting,

  {
    rules: {
      'no-unused-vars': 'off',
      'vue/no-unused-vars': 'off',
      'vue/one-component-per-file': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
    },
  }
)

// import eslint from '@eslint/js'
// import typescriptEslint from 'typescript-eslint'
// import eslintPluginVue from 'eslint-plugin-vue'
// import globals from 'globals'
//
// export default typescriptEslint.config(
//   { ignores: ['*.d.ts', '**/coverage', '**/dist'] },
//   {
//     extends: [
//       eslint.configs.recommended,
//       ...typescriptEslint.configs.recommended,
//       ...eslintPluginVue.configs['flat/recommended'],
//     ],
//     files: ['**/*.{ts,vue}'],
//     languageOptions: {
//       ecmaVersion: 'latest',
//       sourceType: 'module',
//       globals: globals.browser,
//       parserOptions: {
//         parser: typescriptEslint.parser,
//       },
//     },
//     rules: {
//       'no-unused-vars': 'off',
//       'vue/no-unused-vars': 'off',
//       'vue/one-component-per-file': 'off',
//       '@typescript-eslint/no-unused-vars': 'off',
//     },
//   }
// )
