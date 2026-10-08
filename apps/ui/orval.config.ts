import { defineConfig } from "orval";

export default defineConfig({
  nibiru: {
    input: "http://localhost:7000/docs.json",
    output: {
      target: "./src/client/nibiru.ts",
      schemas: "./src/client/model",
      client: "react-query",
    },
  },
});
