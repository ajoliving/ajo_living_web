/*
 * 本地 SVG Mock 圖片產生器。
 * 1. 生成不依賴外部網路的展示圖片。
 * 2. 供列表、詳情、聊天與發布預覽共用。
 */

const escapeSvgText = (value: string) =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');

// 1. 產生可直接嵌入頁面的 SVG 圖片
export const buildMockImage = (
  title: string,
  accentHex: string,
  highlightHex: string,
) => {
  const safeTitle = escapeSvgText(title);

  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 760">
      <defs>
        <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="${accentHex}" />
          <stop offset="100%" stop-color="${highlightHex}" />
        </linearGradient>
      </defs>
      <rect width="1200" height="760" rx="48" fill="url(#bg)" />
      <circle cx="930" cy="120" r="140" fill="rgba(255,255,255,0.16)" />
      <circle cx="220" cy="660" r="220" fill="rgba(255,255,255,0.14)" />
      <rect x="110" y="128" width="980" height="500" rx="40" fill="rgba(255,255,255,0.12)" stroke="rgba(255,255,255,0.3)" />
      <rect x="176" y="196" width="850" height="290" rx="28" fill="rgba(15,23,42,0.12)" stroke="rgba(255,255,255,0.18)" />
      <text x="176" y="575" fill="#ffffff" font-family="Avenir Next, PingFang TC, sans-serif" font-size="62" font-weight="700">${safeTitle}</text>
      <text x="176" y="635" fill="rgba(255,255,255,0.82)" font-family="Avenir Next, PingFang TC, sans-serif" font-size="28">AJO Living Marketplace Prototype</text>
    </svg>
  `;

  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
};
