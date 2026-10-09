import { fileURLToPath } from "node:url";

import babel from "@rolldown/plugin-babel";
import babelPluginReactifx from "@samuelsih/babel-plugin-reactifx";
import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    tailwindcss(),
    tanstackRouter({ target: "react", autoCodeSplitting: true }),
    react(),
    babel({
      presets: [reactCompilerPreset()],
      plugins: [babelPluginReactifx()],
    }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
