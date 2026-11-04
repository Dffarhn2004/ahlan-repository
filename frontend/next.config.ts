import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	output: "standalone",
  /* config options here */
  allowedDevOrigins: ["enormous-mint-tomcat.ngrok-free.app"],
  // Proxy API calls through the frontend origin so one tunnel serves both
  // (same-origin: no CORS, session cookie stays first-party).
  async rewrites() {
    const backend = process.env.BACKEND_URL ?? "http://localhost:8080";
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
  cacheComponents: true,
  partialPrefetching: true,
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
