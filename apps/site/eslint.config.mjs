// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    rules: {
      // prettier (lint-staged) owns formatting and prints void elements self-closed
      'vue/html-self-closing': 'off',
    },
  },
  {
    files: ['components/admin/**'],
    rules: {
      // admin form components edit a reactive draft object owned by the parent page
      'vue/no-mutating-props': 'off',
    },
  },
)
