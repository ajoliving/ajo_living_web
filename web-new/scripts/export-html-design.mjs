/*
 * HTML 設計稿匯出工具。
 * 1. 使用本機 Chrome 渲染部署網站公開頁面。
 * 2. 將渲染後 DOM、樣式與基礎資源路徑整理為可打開的 HTML。
 * 3. 輸出設計稿索引頁，便於產品審閱。
 */
import { mkdir, rm, writeFile } from 'node:fs/promises';
import { mkdtemp } from 'node:fs/promises';
import { once } from 'node:events';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { spawn } from 'node:child_process';
import { setTimeout as delay } from 'node:timers/promises';

import { JSDOM } from 'jsdom';

const DEFAULT_BASE_URL = 'https://ajoliving.skylinedances.com';
const DEFAULT_OUTPUT_DIR = '../docs/web-new-html-design';
const DEFAULT_WAIT_MS = 7000;
const DEFAULT_TIMEOUT_MS = 25000;
const DEFAULT_WINDOW_SIZE = '1440,1200';
const DEFAULT_ROUTES = [
  { path: '/', name: 'home', title: '首頁' },
  { path: '/properties', name: 'properties', title: '樓盤放售' },
  { path: '/serviced-residences', name: 'serviced-residences', title: '服務式住宅' },
  { path: '/furniture', name: 'furniture', title: '家具' },
  { path: '/supermarket-offers', name: 'supermarket-offers', title: '綜合優惠' },
  { path: '/marketplace/discover', name: 'marketplace-discover', title: '二手市集' },
  { path: '/notifications', name: 'notifications', title: '通知中心' },
  { path: '/login', name: 'login', title: '登入' },
];

// 1. 解析命令列參數
const readOption = (name, fallback) => {
  const prefix = `--${name}=`;
  const option = process.argv.find((arg) => arg.startsWith(prefix));

  return option ? option.slice(prefix.length) : fallback;
};

const normalizeBaseUrl = (value) => value.replace(/\/+$/, '');

const baseUrl = normalizeBaseUrl(readOption('base', DEFAULT_BASE_URL));
const outputDir = resolve(process.cwd(), readOption('out', DEFAULT_OUTPUT_DIR));
const waitMs = Number(readOption('wait', String(DEFAULT_WAIT_MS)));
const timeoutMs = Number(readOption('timeout', String(DEFAULT_TIMEOUT_MS)));
const windowSize = readOption('window-size', DEFAULT_WINDOW_SIZE);
const routeArg = readOption('routes', '');
const routes = routeArg
  ? routeArg.split(',').map((path) => ({
      path: path.trim().startsWith('/') ? path.trim() : `/${path.trim()}`,
      name: path.trim().replace(/^\/+|\/+$/g, '').replaceAll('/', '-') || 'home',
      title: path.trim() || '/',
    }))
  : DEFAULT_ROUTES;

// 2. 尋找可用 Chrome
const chromeCandidates = [
  process.env.CHROME_PATH,
  '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  '/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge',
  'google-chrome',
  'chromium',
  'chromium-browser',
].filter(Boolean);

const runCommand = (command, args, commandTimeoutMs = timeoutMs) =>
  new Promise((resolveCommand, rejectCommand) => {
    const child = spawn(command, args, {
      detached: true,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    let stdout = '';
    let stderr = '';
    let isTimedOut = false;

    const timer = setTimeout(() => {
      isTimedOut = true;

      if (child.pid) {
        try {
          process.kill(-child.pid, 'SIGTERM');
        } catch {
          child.kill('SIGTERM');
        }
      }
    }, commandTimeoutMs);

    child.stdout.on('data', (chunk) => {
      stdout += chunk.toString();
    });

    child.stderr.on('data', (chunk) => {
      stderr += chunk.toString();
    });

    child.on('error', (error) => {
      clearTimeout(timer);
      rejectCommand(error);
    });
    child.on('close', (code) => {
      clearTimeout(timer);

      if (isTimedOut) {
        rejectCommand(new Error(`${command} timed out after ${commandTimeoutMs}ms`));
        return;
      }

      if (code === 0) {
        resolveCommand(stdout);
        return;
      }

      rejectCommand(new Error(stderr || `${command} exited with code ${code}`));
    });
  });

const findChrome = async () => {
  for (const candidate of chromeCandidates) {
    try {
      await runCommand(candidate, ['--version'], 5000);
      return candidate;
    } catch {
      // 1. 嘗試下一個候選瀏覽器
    }
  }

  throw new Error('找不到可用 Chrome，請安裝 Google Chrome 或設定 CHROME_PATH。');
};

const getFreePort = () =>
  new Promise((resolvePort, rejectPort) => {
    const server = createServer();

    server.once('error', rejectPort);
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();

      server.close(() => {
        resolvePort(address.port);
      });
    });
  });

const waitForChrome = async (port) => {
  const endpoint = `http://127.0.0.1:${port}/json/version`;
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    try {
      const response = await fetch(endpoint);

      if (response.ok) {
        return response.json();
      }
    } catch {
      // 1. Chrome 尚未完成啟動，短暫等待後重試
    }

    await delay(250);
  }

  throw new Error('Chrome 調試端口啟動逾時。');
};

const stopChrome = async (child) => {
  if (!child || child.exitCode !== null) {
    return;
  }

  if (child.pid) {
    try {
      process.kill(-child.pid, 'SIGTERM');
    } catch {
      child.kill('SIGTERM');
    }
  }

  try {
    await Promise.race([
      once(child, 'exit'),
      delay(3000).then(() => {
        if (child.exitCode === null && child.pid) {
          process.kill(-child.pid, 'SIGKILL');
        }
      }),
    ]);
  } catch {
    // 1. Chrome 可能已自行退出，無需額外處理
  }
};

// 3. 啟動 Chrome 並透過 CDP 取得渲染後 DOM
const launchChrome = async (chromePath, userDataDir, port) => {
  const child = spawn(chromePath, [
    '--headless=new',
    '--disable-gpu',
    '--no-sandbox',
    '--disable-background-networking',
    '--disable-default-apps',
    '--disable-extensions',
    '--disable-sync',
    '--no-first-run',
    `--remote-debugging-port=${port}`,
    '--remote-debugging-address=127.0.0.1',
    `--user-data-dir=${userDataDir}`,
    `--window-size=${windowSize}`,
    'about:blank',
  ], {
    detached: true,
    stdio: ['ignore', 'ignore', 'pipe'],
  });

  let stderr = '';

  child.stderr.on('data', (chunk) => {
    stderr += chunk.toString();
  });

  try {
    await waitForChrome(port);
    return child;
  } catch (error) {
    await stopChrome(child);
    throw new Error(stderr || error.message);
  }
};

const createCdpClient = (webSocketUrl) =>
  new Promise((resolveClient, rejectClient) => {
    const socket = new WebSocket(webSocketUrl);
    const callbacks = new Map();
    let nextId = 1;
    let isOpen = false;

    const client = {
      send(method, params = {}) {
        return new Promise((resolveCommand, rejectCommand) => {
          const id = nextId;
          nextId += 1;
          callbacks.set(id, { resolveCommand, rejectCommand });
          socket.send(JSON.stringify({ id, method, params }));
        });
      },
      close() {
        socket.close();
      },
    };

    socket.addEventListener('open', () => {
      isOpen = true;
      resolveClient(client);
    });

    socket.addEventListener('message', (event) => {
      const message = JSON.parse(String(event.data));

      if (!message.id || !callbacks.has(message.id)) {
        return;
      }

      const callback = callbacks.get(message.id);
      callbacks.delete(message.id);

      if (message.error) {
        callback.rejectCommand(new Error(message.error.message));
        return;
      }

      callback.resolveCommand(message.result);
    });

    socket.addEventListener('error', (error) => {
      if (!isOpen) {
        rejectClient(error);
      }
    });

    socket.addEventListener('close', () => {
      callbacks.forEach((callback) => {
        callback.rejectCommand(new Error('Chrome 調試連線已關閉。'));
      });
      callbacks.clear();
    });
  });

const openTarget = async (port, url) => {
  const response = await fetch(
    `http://127.0.0.1:${port}/json/new?${encodeURIComponent(url)}`,
    { method: 'PUT' },
  );

  if (!response.ok) {
    throw new Error(`建立頁面失敗：${url} ${response.status}`);
  }

  return response.json();
};

const closeTarget = async (port, targetId) => {
  try {
    await fetch(`http://127.0.0.1:${port}/json/close/${targetId}`);
  } catch {
    // 1. 頁面關閉失敗不影響設計稿輸出
  }
};

const renderPage = async (port, url) => {
  const target = await openTarget(port, url);
  const client = await createCdpClient(target.webSocketDebuggerUrl);

  try {
    await client.send('Page.enable');
    await client.send('Runtime.enable');
    await delay(waitMs);

    const result = await client.send('Runtime.evaluate', {
      expression: 'document.documentElement.outerHTML',
      returnByValue: true,
      awaitPromise: true,
    });

    if (result.exceptionDetails) {
      throw new Error('讀取頁面 DOM 失敗。');
    }

    return `<!doctype html>\n${result.result.value}`;
  } finally {
    client.close();
    await closeTarget(port, target.id);
  }
};

const toAbsoluteUrl = (value) => {
  if (!value || value.startsWith('data:') || value.startsWith('mailto:') || value.startsWith('tel:')) {
    return value;
  }

  return new URL(value, baseUrl).toString();
};

const rewriteCssUrls = (cssText) =>
  cssText.replace(/url\((['"]?)(?!data:|https?:|\/\/)([^'")]+)\1\)/g, (_match, quote, url) => {
    const absoluteUrl = toAbsoluteUrl(url.trim());

    return `url(${quote}${absoluteUrl}${quote})`;
  });

const stripEmoji = (value) =>
  value
    .replace(/[\u{1F000}-\u{1FAFF}]/gu, '')
    .replace(/[\u{2600}-\u{27BF}]/gu, '')
    .replace(/\uFE0F/gu, '');

const fetchText = async (url) => {
  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`讀取樣式失敗：${url} ${response.status}`);
  }

  return response.text();
};

// 4. 將動態頁面整理成靜態設計稿
const cleanupDocument = async (html, pageTitle) => {
  const dom = new JSDOM(html);
  const { document } = dom.window;

  document.documentElement.lang = 'zh-HK';
  document.body.innerHTML = stripEmoji(document.body.innerHTML);

  document.querySelectorAll('script, link[rel="modulepreload"]').forEach((node) => {
    node.remove();
  });

  const stylesheetLinks = [...document.querySelectorAll('link[rel="stylesheet"][href]')];
  for (const link of stylesheetLinks) {
    const href = toAbsoluteUrl(link.getAttribute('href'));

    try {
      const style = document.createElement('style');
      style.textContent = rewriteCssUrls(await fetchText(href));
      link.replaceWith(style);
    } catch {
      link.setAttribute('href', href);
    }
  }

  document.querySelectorAll('[href]').forEach((node) => {
    const value = node.getAttribute('href');

    if (value && (value.startsWith('/') || value.startsWith('./') || value.startsWith('../'))) {
      node.setAttribute('href', toAbsoluteUrl(value));
    }
  });

  document.querySelectorAll('[src]').forEach((node) => {
    const value = node.getAttribute('src');

    if (value && (value.startsWith('/') || value.startsWith('./') || value.startsWith('../'))) {
      node.setAttribute('src', toAbsoluteUrl(value));
    }
  });

  const title = document.querySelector('title') ?? document.createElement('title');
  title.textContent = `${pageTitle} | AJO Living HTML Design`;

  if (!title.parentElement) {
    document.head.appendChild(title);
  }

  const marker = document.createElement('meta');
  marker.setAttribute('name', 'ajo-html-design-source');
  marker.setAttribute('content', baseUrl);
  document.head.appendChild(marker);

  return `<!doctype html>\n${document.documentElement.outerHTML}\n`;
};

const buildIndexHtml = (results) => {
  const links = results
    .map(
      (result) => `
        <a class="card" href="./${result.file}">
          <span>${result.title}</span>
          <small>${result.path}</small>
        </a>`,
    )
      .join('');

  return `<!doctype html>
<html lang="zh-HK">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AJO Living HTML Design</title>
  <style>
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Avenir Next", "PingFang TC", "Noto Sans TC", sans-serif;
      color: #1a1a1a;
      background: #f7f7f7;
    }
    main {
      width: min(960px, calc(100vw - 32px));
      margin: 0 auto;
      padding: 48px 0;
    }
    h1 {
      margin: 0 0 8px;
      font-size: 28px;
      font-weight: 600;
    }
    p {
      margin: 0 0 24px;
      color: #666;
      line-height: 1.7;
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
      gap: 12px;
    }
    .card {
      display: flex;
      flex-direction: column;
      gap: 8px;
      padding: 18px;
      color: inherit;
      text-decoration: none;
      background: #fff;
      border: 1px solid #e4e4e4;
      border-radius: 6px;
    }
    .card:hover {
      border-color: #c04600;
    }
    .card span {
      font-size: 16px;
      font-weight: 600;
    }
    .card small {
      color: #777;
    }
  </style>
</head>
<body>
  <main>
    <h1>AJO Living HTML Design</h1>
    <p>來源：${baseUrl}</p>
    <p><a class="card" href="./all-in-one.html"><span>合併設計稿</span><small>全部頁面整合為單一 HTML</small></a></p>
    <section class="grid">${links}
    </section>
  </main>
</body>
</html>
`;
};

const escapeSrcdoc = (value) =>
  value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');

const buildAllInOneHtml = (results) => {
  const sections = results
    .map(
      (result) => `
        <section id="${result.name}" class="page-block">
          <div class="page-block__header">
            <div>
              <p>${result.path}</p>
              <h2>${result.title}</h2>
            </div>
            <a href="${baseUrl}${result.path}" target="_blank" rel="noreferrer">查看來源頁</a>
          </div>
          <iframe title="${result.title}" loading="lazy" srcdoc="${escapeSrcdoc(result.html)}"></iframe>
        </section>`,
    )
    .join('');

  return `<!doctype html>
<html lang="zh-HK">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AJO Living All-in-One HTML Design</title>
  <style>
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Avenir Next", "PingFang TC", "Noto Sans TC", sans-serif;
      color: #1a1a1a;
      background: #f4f4f4;
    }
    main {
      width: min(1200px, calc(100vw - 32px));
      margin: 0 auto;
      padding: 24px 0 56px;
    }
    .page-block {
      margin-bottom: 28px;
      border: 1px solid #e4e4e4;
      border-radius: 8px;
      background: #fff;
      overflow: hidden;
    }
    .page-block__header {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      padding: 16px 18px;
      border-bottom: 1px solid #e4e4e4;
    }
    .page-block__header p {
      margin: 0 0 4px;
      color: #777;
      font-size: 12px;
    }
    .page-block__header h2 {
      margin: 0;
      font-size: 18px;
      font-weight: 650;
    }
    .page-block__header a {
      flex: 0 0 auto;
      color: #c04600;
      font-size: 13px;
      font-weight: 650;
      text-decoration: none;
    }
    iframe {
      display: block;
      width: 100%;
      height: 980px;
      border: 0;
      background: #fff;
    }
    @media (max-width: 640px) {
      .page-block__header {
        align-items: flex-start;
        flex-direction: column;
      }
      iframe {
        height: 820px;
      }
    }
  </style>
</head>
<body>
  <main>${sections}
  </main>
</body>
</html>
`;
};

// 5. 執行匯出流程
const main = async () => {
  const chromePath = await findChrome();
  const userDataDir = await mkdtemp(join(tmpdir(), 'ajo-html-design-'));
  const port = await getFreePort();
  const chrome = await launchChrome(chromePath, userDataDir, port);
  const results = [];

  await mkdir(outputDir, { recursive: true });

  try {
    for (const route of routes) {
      const url = `${baseUrl}${route.path}`;
      try {
        const rawHtml = await renderPage(port, url);
        const html = await cleanupDocument(rawHtml, route.title);
        const file = `${route.name}.html`;

        await writeFile(join(outputDir, file), html);
        results.push({ ...route, file, html });
        console.log(`exported ${route.path} -> ${file}`);
      } catch (error) {
        console.error(`failed ${route.path}: ${error.message}`);
      }
    }
  } finally {
    await stopChrome(chrome);
    await rm(userDataDir, { recursive: true, force: true });
  }

  await writeFile(join(outputDir, 'index.html'), buildIndexHtml(results));
  await writeFile(join(outputDir, 'all-in-one.html'), buildAllInOneHtml(results));
  console.log(`done ${outputDir}`);
};

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
