import pako from 'pako';

export interface ShareData {
  mdflow: string;
  template?: string;
  createdAt: number;
}

/**
 * Compress and encode data for URL sharing
 * Flow: JSON → gzip → base64 → URL-safe base64
 */
export function encodeShareData(data: ShareData): string {
  try {
    const jsonStr = JSON.stringify(data);
    const compressed = pako.gzip(jsonStr);
    // Convert Uint8Array to base64
    const base64 = btoa(String.fromCharCode.apply(null, Array.from(compressed)));
    // Make URL-safe: replace + with -, / with _, remove padding =
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  } catch (error) {
    console.error('Failed to encode share data:', error);
    throw new Error('Failed to create share link');
  }
}

/**
 * Decode and decompress data from URL
 * Flow: URL-safe base64 → base64 → gunzip → JSON
 */
export function decodeShareData(encoded: string): ShareData | null {
  try {
    // Restore base64: replace - with +, _ with /, add padding
    let base64 = encoded.replace(/-/g, '+').replace(/_/g, '/');
    // Add padding if needed
    while (base64.length % 4) {
      base64 += '=';
    }
    
    // Decode base64 to Uint8Array
    const binaryStr = atob(base64);
    const bytes = new Uint8Array(binaryStr.length);
    for (let i = 0; i < binaryStr.length; i++) {
      bytes[i] = binaryStr.charCodeAt(i);
    }
    
    // Decompress
    const decompressed = pako.ungzip(bytes, { to: 'string' });
    return JSON.parse(decompressed) as ShareData;
  } catch (error) {
    console.error('Failed to decode share data:', error);
    return null;
  }
}

/**
 * Generate a shareable URL
 */
export function generateShareURL(data: ShareData): string {
  const encoded = encodeShareData(data);
  const baseUrl = typeof window !== 'undefined' 
    ? window.location.origin 
    : 'https://md-spec-tool.vercel.app';
  return `${baseUrl}/share?d=${encoded}`;
}

/**
 * Check if a share URL data would be too long
 * Most browsers support URLs up to 2000-8000 characters
 * We'll use 6000 as a safe limit
 */
export function isShareDataTooLong(data: ShareData): boolean {
  try {
    const encoded = encodeShareData(data);
    return encoded.length > 6000;
  } catch {
    return true;
  }
}

/**
 * Get approximate size of share data in KB
 */
export function getShareDataSize(data: ShareData): number {
  try {
    const encoded = encodeShareData(data);
    return Math.round(encoded.length / 1024 * 10) / 10;
  } catch {
    return 0;
  }
}
