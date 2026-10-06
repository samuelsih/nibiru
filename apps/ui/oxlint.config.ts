import { defineConfig } from "oxlint";

export default defineConfig({
  plugins: ["react", "jsx-a11y", "import", "promise", "unicorn", "typescript", "oxc"],

  ignorePatterns: ["src/routeTree.gen.ts"],

  categories: {
    correctness: "error",
    suspicious: "warn",
    perf: "warn",
  },

  env: {
    browser: true,
    es2024: true,
  },

  rules: {
    // React
    "react/rules-of-hooks": "error",
    "react/exhaustive-deps": "warn",
    "react/only-export-components": ["warn", { allowConstantExport: true }],
    "react/button-has-type": "error",
    "react/jsx-boolean-value": ["warn", "never"],
    "react/jsx-curly-brace-presence": ["warn", { props: "never", children: "never" }],
    "react/jsx-fragments": ["warn", "syntax"],
    "react/jsx-key": "error",
    "react/jsx-no-constructed-context-values": "warn",
    "react/jsx-no-target-blank": "error",
    "react/no-array-index-key": "warn",
    "react/no-children-prop": "error",
    "react/no-danger": "warn",
    "react/no-direct-mutation-state": "error",
    "react/no-string-refs": "error",
    "react/no-this-in-sfc": "error",
    "react/no-unknown-property": "error",
    "react/no-unstable-nested-components": "error",
    "react/prefer-function-component": "warn",
    "react/react-in-jsx-scope": "off",
    "react/require-render-return": "error",
    "react/self-closing-comp": "warn",
    "react/void-dom-elements-no-children": "error",

    // React Compiler
    "react/error-boundaries": "error",
    "react/immutability": "error",
    "react/preserve-manual-memoization": "error",
    "react/purity": "error",
    "react/refs": "error",
    "react/set-state-in-effect": "warn",
    "react/set-state-in-render": "error",
    "react/static-components": "error",
    "react/unsupported-syntax": "error",
    "react/void-use-memo": "error",

    // JSX a11y
    "jsx-a11y/alt-text": "error",
    "jsx-a11y/anchor-has-content": "error",
    "jsx-a11y/anchor-is-valid": "error",
    "jsx-a11y/aria-activedescendant-has-tabindex": "error",
    "jsx-a11y/aria-props": "error",
    "jsx-a11y/aria-proptypes": "error",
    "jsx-a11y/aria-role": "error",
    "jsx-a11y/aria-unsupported-elements": "error",
    "jsx-a11y/click-events-have-key-events": "warn",
    "jsx-a11y/heading-has-content": "error",
    "jsx-a11y/html-has-lang": "error",
    "jsx-a11y/iframe-has-title": "error",
    "jsx-a11y/img-redundant-alt": "warn",
    "jsx-a11y/no-aria-hidden-on-focusable": "error",
    "jsx-a11y/no-autofocus": "warn",
    "jsx-a11y/no-distracting-elements": "error",
    "jsx-a11y/no-redundant-roles": "error",
    "jsx-a11y/role-has-required-aria-props": "error",
    "jsx-a11y/role-supports-aria-props": "error",
    "jsx-a11y/scope": "error",
    "jsx-a11y/tabindex-no-positive": "warn",

    // TypeScript
    "typescript/consistent-type-imports": "warn",
    "typescript/no-explicit-any": "warn",
    "typescript/no-namespace": "error",
    "typescript/no-non-null-assertion": "warn",

    // Imports
    "import/newline-after-import": "warn",
    "import/no-duplicates": "error",
    "import/no-self-import": "error",
    "import/no-unassigned-import": ["warn", { allow: ["**/*.css"] }],

    // General
    eqeqeq: ["error", "smart"],
    "no-console": ["warn", { allow: ["warn", "error"] }],
    "no-var": "error",
    "object-shorthand": "warn",
    "prefer-const": "error",
  },

  overrides: [
    {
      // TanStack Router route files export the `Route` object alongside the component.
      files: ["src/routes/**/*.tsx"],
      rules: {
        "react/only-export-components": "off",
      },
    },
  ],
});
