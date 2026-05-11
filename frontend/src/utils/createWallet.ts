export const generateKeys = async (): Promise<{
  publicKeyPem: string;
  privateKeyPem: string;
}> => {
  const keyPair = await window.crypto.subtle.generateKey(
    {
      name: 'RSA-PSS',
      modulusLength: 2048,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: { name: 'SHA-256' },
    },
    true,
    ['sign', 'verify'],
  );

  const publicKey = await window.crypto.subtle.exportKey(
    'spki',
    keyPair.publicKey,
  );
  const privateKey = await window.crypto.subtle.exportKey(
    'pkcs8',
    keyPair.privateKey,
  );

  const publicKeyPem = `-----BEGIN PUBLIC KEY-----\n${btoa(String.fromCharCode(...new Uint8Array(publicKey)))}\n-----END PUBLIC KEY-----`;
  const privateKeyPem = `-----BEGIN PRIVATE KEY-----\n${btoa(String.fromCharCode(...new Uint8Array(privateKey)))}\n-----END PRIVATE KEY-----`;

  return { publicKeyPem, privateKeyPem };
};

export const decryptData = async (
  privateKeyPem: string,
  encryptedData: string,
): Promise<string> => {
  const pemHeader = '-----BEGIN PRIVATE KEY-----';
  const pemFooter = '-----END PRIVATE KEY-----';
  const pemContents = privateKeyPem
    .replace(pemHeader, '')
    .replace(pemFooter, '')
    .replace(/\n/g, '');
  const privateKeyBytes = Uint8Array.from(atob(pemContents), (c) =>
    c.charCodeAt(0),
  );

  const privateKey = await window.crypto.subtle.importKey(
    'pkcs8',
    privateKeyBytes.buffer,
    {
      name: 'RSA-OAEP',
      hash: { name: 'SHA-256' },
    },
    true,
    ['decrypt'],
  );

  const encryptedBytes = Uint8Array.from(atob(encryptedData), (c) =>
    c.charCodeAt(0),
  );

  const decryptedArrayBuffer = await window.crypto.subtle.decrypt(
    {
      name: 'RSA-OAEP',
    },
    privateKey,
    encryptedBytes,
  );

  const decoder = new TextDecoder();

  return decoder.decode(decryptedArrayBuffer);
};
