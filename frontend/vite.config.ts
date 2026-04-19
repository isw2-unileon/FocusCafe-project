import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";


export default defineConfig(({ mode }) =>{
  const env = loadEnv(mode, process.cwd(), '');
  const backendPort = env.BACKEND_PORT || "8080";
  
  return {
      plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      port: 5173,
      proxy: {
        "/api": {
          target: `http://localhost:${backendPort}`,
          changeOrigin: true,
        },
        "/health": {
          target: `http://localhost:${backendPort}`,
          changeOrigin: true,
        },
      },
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks: {
            "vendor-react": ["react", "react-dom"],
          },
        },
      },
    },
    esbuild: {
      drop: process.env.NODE_ENV === "production" ? ["console", "debugger"] : [],
    },
  }
});
