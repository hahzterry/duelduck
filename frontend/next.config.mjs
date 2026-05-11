/** @type {import('next').NextConfig} */

const IS_STAGE = process.env.NEXT_PUBLIC_DUEL_DUCK_API?.includes('stage');

const securityHeaders = [
  {
    key: 'X-DNS-Prefetch-Control',
    value: 'on',
  },
  ...(!IS_STAGE
    ? [
        {
          key: 'Strict-Transport-Security',
          value: 'max-age=31536000; includeSubDomains; preload',
        },
      ]
    : []),
  {
    key: 'X-XSS-Protection',
    value: '1; mode=block',
  },
  {
    key: 'X-Content-Type-Options',
    value: 'nosniff',
  },
  {
    key: 'Referrer-Policy',
    value: 'origin-when-cross-origin',
  },
  { key: "Access-Control-Allow-Headers", value: "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version, x-stage-ac" },
  // {
  //   key: 'X-Frame-Options',
  //   value: 'SAMEORIGIN',
  // },
  {
    key: 'Content-Security-Policy',
    value: [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.googletagmanager.com https://*.google-analytics.com https://cdn.jsdelivr.net https://vercel.live https://*.cloudflareinsights.com https://challenges.cloudflare.com https://*.google.com https://unpkg.com https://*.walletconnect.com https://*.web3modal.com https://*.contentsquare.net https://static.ads-twitter.com",
      "worker-src 'self' blob:",
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
      "font-src 'self' https://fonts.gstatic.com data:",
      "img-src 'self' data: https: blob:",
      "media-src 'self' https: blob:",
      "connect-src 'self' https://*.duelduck.com https://*.google-analytics.com https://vercel.live https://*.web3modal.org https://*.walletconnect.org https://*.walletconnect.com https://*.helius-rpc.com https://*.extrnode.com https://rpc.whitechain.io https://*.googleapis.com https://*.jup.ag https://unpkg.com https://cdn.jsdelivr.net wss: https:",
      "frame-src 'self' https://www.youtube.com https://player.vimeo.com https://vercel.live https://app.supademo.com https://challenges.cloudflare.com https://*.firebaseapp.com",
      "object-src 'none'",
      "base-uri 'self'",
      "form-action 'self'",
      "frame-ancestors *",
      'upgrade-insecure-requests',
    ].join('; '),
  },
  {
    key: 'Permissions-Policy',
    value: 'camera=(), microphone=(), geolocation=(), interest-cohort=()',
  },
];

const nextConfig = {
  trailingSlash: false,
  skipTrailingSlashRedirect: true,
  skipProxyUrlNormalize: true,
  output: 'standalone',
  distDir: './dist',
  sassOptions: {
    additionalData: `
    @use 'sass:math';
    @use 'sass:map';
    @use "./src/styles/constants" as *;
    @use "./src/styles/colors" as *;
    @use "./src/styles/mixins" as *;
  `,
  },
  images: {
    unoptimized: true,
    disableStaticImages: true,
  },
  experimental: {
    optimizePackageImports: [
      'firebase',
      'firebase/auth',
      'firebase/firestore',
      'firebase/messaging',
      'ethers',
      'wagmi',
      '@reown/appkit',
      '@reown/appkit-adapter-solana',
      '@reown/appkit-adapter-wagmi',
      '@solana/web3.js',
      '@solana/spl-token',
      '@solana/wallet-adapter-wallets',
      'react-use',
      'react-datepicker',
      'swiper',
      'react-loading-skeleton',
      'react-transition-group',
      'classnames',
      '@tanstack/react-virtual',
    ],
  },
  compiler: {
    removeConsole: !IS_STAGE ? { exclude: ['error'] } : false,
  },
  reactCompiler: true,
  reactStrictMode: false,
  async headers() {
    return [
      {
        source: '/:path*',
        headers: securityHeaders,
      },
      {
        source: '/video/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
        ],
      },
      {
        source: '/mentionboard',
        headers: [
          ...securityHeaders,
          {
            key: 'Cache-Control',
            value: 'public, s-maxage=60, stale-while-revalidate=300',
          },
        ],
      },
    ];
  },
  async redirects() {
    return [
      {
        source: '/mentionboard',
        has: [
          {
            type: 'header',
            key: 'x-forwarded-proto',
            value: 'http',
          },
        ],
        destination: 'https://duelduck.com/mentionboard',
        permanent: true,
      },
      // SEO tech-debt — /duel/{uuid} is the old route, now at /duels/{uuid}.
      // Preserve link equity from 238 indexed URLs.
      { source: '/duel', destination: '/duels', permanent: true },
      { source: '/duel/:path*', destination: '/duels/:path*', permanent: true },
      // SEO tech-debt — plural /tournaments/{slug} is not a live route; the
      // canonical route is singular /tournament/{slug}. 301 so Google stops
      // hitting the 404 (only nested paths — /tournaments itself is a
      // live listing page). Deeper asset paths end up 301→404; the
      // middleware layers on top of that with a real 410 for asset
      // extensions so crawl budget is freed faster.
      {
        source: '/tournaments/:path+',
        destination: '/tournament/:path+',
        permanent: true,
      },
      // SEO tech-debt — /leaderboard/page/:num pages were removed from HTML
      // but still crawled. Collapse them back to /leaderboard.
      {
        source: '/leaderboard/page/:num',
        destination: '/leaderboard',
        permanent: true,
      },
      // Case-sensitivity / legacy casing for create-duel.
      { source: '/createduel', destination: '/create-duel', statusCode: 301 },
      { source: '/createDuel', destination: '/create-duel', statusCode: 301 },
    ];
  },
};

export default nextConfig;
