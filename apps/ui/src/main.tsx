import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import "./styles.css";
import "reshaped/themes/slate/theme.css";

import { getRouter } from "./router";
import { Reshaped } from "reshaped";

const queryClient = new QueryClient({
  defaultOptions: {},
});

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("Root element #root not found");
}

createRoot(rootElement).render(
  <StrictMode>
    <Reshaped theme="slate">
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={getRouter()} />
      </QueryClientProvider>
    </Reshaped>
  </StrictMode>,
);
