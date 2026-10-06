import { defineConfig } from "oxfmt";

export default defineConfig({
  printWidth: 100,
  singleQuote: false,
  trailingComma: "all",
  sortImports: true,
  sortPackageJson: true,
  ignorePatterns: ["src/routeTree.gen.ts"],
});
