import { defineConfig } from "orval";

export default defineConfig({
  nibiru: {
    input: {
      target: "http://localhost:7000/docs.json",
    },
    output: {
      mode: "tags-split",
      target: "./src/client/nibiru.ts",
      schemas: "./src/client/model",
      client: "fetch",
      baseUrl: process.env.API_BASE_URL || "http://localhost:7000/api",
    },
  },
});
