import react from "@vitejs/plugin-react";
import wails from "@wailsio/runtime/plugins/vite";
import path from "path";
import { defineConfig } from "vite";
import babel from "vite-plugin-babel";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    wails("./bindings"),
    react(),
    babel({
      babelConfig: {
        plugins: ["babel-plugin-react-compiler"],
      },
    }),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@bindings": path.resolve(__dirname, "./bindings"),
    },
  },
});
