import type { NextConfig } from "next";

const apiOrigin = process.env.CONTROL_API_ORIGIN ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  poweredByHeader: false,
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiOrigin}/:path*` }];
  },
};

export default nextConfig;
