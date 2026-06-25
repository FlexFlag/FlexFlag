/** @type {import('next').NextConfig} */
const nextConfig = {
  // 'standalone' is for Docker; skip it on Vercel (uses its own output format)
  output: process.env.VERCEL ? undefined : 'standalone',
  async rewrites() {
    // Use environment variable for API URL, fallback to localhost for local dev
    const apiUrl = process.env.INTERNAL_API_URL || 'http://localhost:8080';
    return [
      {
        source: '/api/v1/:path*',
        destination: `${apiUrl}/api/v1/:path*`,
      },
    ]
  },
}

module.exports = nextConfig