const mimeToExt: Record<string, string> = {
  'image/jpeg': 'jpg',
  'image/jpg': 'jpg',
  'image/png': 'png',
  'image/webp': 'webp',
  'image/gif': 'gif',
  'image/svg+xml': 'svg',
  'image/bmp': 'bmp',
  'image/tiff': 'tiff',
  'image/x-icon': 'ico',
};

const sanitizeBaseName = (name: string) =>
  name
    .trim()
    .replace(/\.[a-z0-9]+$/i, '')
    .replace(/[\\/:*?"<>|]+/g, '-')
    .replace(/\s+/g, '_')
    .slice(0, 120) || 'file';

export const dataURLtoFile = (dataUrl: string, baseName = 'image'): File => {
  const [meta, payload] = dataUrl.split(',');

  const mimeMatch = meta?.match(/data:([^;]+);/i);
  const mime = mimeMatch?.[1] ?? 'image/png';

  const ext =
    mimeToExt[mime] ?? (mime.includes('/') ? mime.split('/')[1] : 'bin');
  const filename = `${sanitizeBaseName(baseName)}.${ext}`;

  const bstr = atob(payload ?? '');
  const u8arr = new Uint8Array(bstr.length);

  for (let i = 0; i < bstr.length; i++) {
    u8arr[i] = bstr.charCodeAt(i);
  }

  return new File([u8arr], filename, { type: mime });
};
