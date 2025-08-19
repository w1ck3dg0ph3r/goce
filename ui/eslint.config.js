import eslint from '@eslint/js'
import typescriptEslint from 'typescript-eslint'
import eslintPluginVue from 'eslint-plugin-vue'
import globals from 'globals'

export default typescriptEslint.config(
  { ignores: ['*.d.ts', '**/coverage', '**/dist'] },
  {
    extends: [
      eslint.configs.recommended,
      ...typescriptEslint.configs.recommended,
      ...eslintPluginVue.configs['flat/recommended'],
    ],
    files: ['**/*.{ts,vue}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: globals.browser,
      parserOptions: {
        parser: typescriptEslint.parser,
      },
    },
    rules: {
      'no-unused-vars': 'off',
      'vue/no-unused-vars': 'off',
      'vue/one-component-per-file': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
    },
  }
)
