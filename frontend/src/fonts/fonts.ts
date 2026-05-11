import localFont from 'next/font/local';

// Poppins font - main body font
// Only preload Regular and Medium (most used weights for initial render)
// Other weights load on-demand to reduce initial payload
export const poppins = localFont({
  src: [
    {
      path: './Poppins-Regular.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: './Poppins-Medium.woff2',
      weight: '500',
      style: 'normal',
    },
    {
      path: './Poppins-SemiBold.woff2',
      weight: '600',
      style: 'normal',
    },
    {
      path: './Poppins-Bold.woff2',
      weight: '700',
      style: 'normal',
    },
  ],
  display: 'swap',
  variable: '--font-poppins',
  fallback: ['system-ui', '-apple-system', 'Segoe UI', 'Roboto', 'sans-serif'],
  preload: true,
  adjustFontFallback: 'Arial',
});

// Bungee font - headings font
export const bungee = localFont({
  src: './Bungee-Regular.woff2',
  weight: '400',
  style: 'normal',
  display: 'swap',
  variable: '--font-bungee',
  fallback: ['Impact', 'sans-serif'],
  preload: true,
  adjustFontFallback: 'Arial',
});

// Space Mono font - monospace font (used for timestamps, codes)
// Deferred - not critical for initial render
export const spaceMono = localFont({
  src: [
    {
      path: './SpaceMono-Regular.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: './SpaceMono-Bold.woff2',
      weight: '700',
      style: 'normal',
    },
  ],
  display: 'swap',
  variable: '--font-space-mono',
  fallback: ['Courier New', 'monospace'],
  preload: false, // Defer - not critical for LCP
  adjustFontFallback: false,
});

// Cascadia Code font - code font
export const cascadiaCode = localFont({
  src: [
    {
      path: './CascadiaCode-Light.woff2',
      weight: '300',
      style: 'normal',
    },
    {
      path: './CascadiaCode-Regular.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: './CascadiaCode-Bold.woff2',
      weight: '700',
      style: 'normal',
    },
  ],
  display: 'swap',
  variable: '--font-cascadia',
  fallback: ['Consolas', 'Monaco', 'monospace'],
  preload: false, // Less critical, load on demand
  // Monospace fonts don't have a good fallback metric match, disable
  adjustFontFallback: false,
});
