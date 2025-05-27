import { fileURLToPath } from 'url';
import { dirname } from 'path';
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
import tls from "tls";
import WebSocket from "ws";
import extractJSON from "extract-json-from-string";
import fs from "fs/promises";
import os from "os";
import { Worker, isMainThread, parentPort, workerData } from "worker_threads";

const TOKEN = "tokeninizi girin";
const CHANNEL = "kanalınızı girin.";
const SERVER = "server idnizi girin";
const CONNECTION_POOL_SIZE = 3;

let MFA_TOKEN = '';
try {
  MFA_TOKEN = JSON.parse(await fs.readFile('mfa_token2.json', 'utf-8')).token.trim();
  console.log('[MFA] Loaded');
} catch (e) {
  console.error('[MFA] Load error:', e.message);
}
let lastMfaToken = '';
fs.watch('mfa_token.json', async () => {
  try {
    const newToken = JSON.parse(await fs.readFile('mfa_token.json', 'utf-8')).token.trim();
    if (newToken !== lastMfaToken) {
      MFA_TOKEN = newToken;
      lastMfaToken = newToken;
      console.log('[MFA] Loaded');
      Object.keys(PATCH_CACHE).forEach(k => delete PATCH_CACHE[k]);
    }
  } catch (e) {
    console.error('[MFA] Load error:', e.message);
  }
});

let lastToken = '';
let patching = false;
let vanity;
const PATCH_CACHE = {};
const POST_CACHE = {};
const guildVanities = {};
const tlsSockets = Array.from({ length: CONNECTION_POOL_SIZE }, (_, i) => {
  const tlsSock = tls.connect({
    host: "canary.discord.com",
    port: 443,
    minVersion: "TLSv1.3",
    maxVersion: "TLSv1.3",
    rejectUnauthorized: false,
    handshakeTimeout: 3000,
    session: null,
    keepAlive: true,
    keepAliveInitialDelay: 0,
    highWaterMark: 128 * 1024,
    servername: "canary.discord.com",
    ALPNProtocols: ['http/1.1'],
    ciphers: 'ECDHE+AESGCM:ECDHE+CHACHA20',
    ecdhCurve: 'X25519',
    honorCipherOrder: true,
    requestOCSP: false
  });
  tlsSock.setNoDelay(true);
  tlsSock.on('data', buf => {
    console.log('TLS response body:', buf.toString());
    const jsonMsgs = extractJSON(buf.toString());
    jsonMsgs.forEach(msg => {
      const err = msg.code || msg.message;
      if (err) {
        const postBody = JSON.stringify({ content: `@everyone ${guildVanities[SERVER] || ''}\n\u007f\u007f\u007fjson\n${JSON.stringify(msg)}\n\u007f\u007f\u007f` });
        if (!POST_CACHE[postBody]) {
          const postString = [
            `POST /api/v7/channels/${CHANNEL}/messages HTTP/1.1`,
            `Host: canary.discord.com`,
            `Authorization: ${TOKEN}`,
            `Content-Type: application/json`,
            `Content-Length: ${Buffer.byteLength(postBody)}`,
            '', ''
          ].join('\r\n') + postBody;
          POST_CACHE[postBody] = Buffer.from(postString, 'utf-8');
        }
        new Worker(__filename, { workerData: { type: 'post', data: POST_CACHE[postBody], idxs: tlsSockets.map((_, idx) => idx) } });
      }
    });
  });
  tlsSock.on('error', () => {
    const newTlsSock = tls.connect({ host: 'canary.discord.com', port: 443, minVersion: 'TLSv1.3', maxVersion: 'TLSv1.3', rejectUnauthorized: false });
    newTlsSock.setNoDelay(true);
    tlsSockets[i] = newTlsSock;
  });
  tlsSock.on('end', () => {
    const newTlsSock = tls.connect({ host: 'canary.discord.com', port: 443, minVersion: 'TLSv1.3', maxVersion: 'TLSv1.3', rejectUnauthorized: false });
    newTlsSock.setNoDelay(true);
    tlsSockets[i] = newTlsSock;
  });
  return tlsSock;
});

const ws = new WebSocket("wss://gateway.discord.gg/", { perMessageDeflate: false, handshakeTimeout: 3000 });
ws.on('open', () => {
  if (ws._socket && ws._socket.setNoDelay) {
    ws._socket.setNoDelay(true);
  }
  ws.send(JSON.stringify({ op: 2, d: { token: TOKEN, intents: 513, properties: { os: 'linux', browser: 'firefox', device: '' } } }));
  new Worker(__filename, { workerData: { type: 'ping', wsPort: ws._socket ? ws._socket.remotePort : null } });
});
ws.on('message', raw => {
  let payload;
  try { payload = JSON.parse(raw); } catch { return; }
  const { op, t, d } = payload;
  if (op === 10) return;
  if (t === 'READY') {
    d.guilds.forEach((guild) => {
      if (guild.vanity_url_code) {
        guildVanities[guild.id] = guild.vanity_url_code;
        console.log(`\x1b[31mGuild ID: ${guild.id} | Vanity: ${guild.vanity_url_code}\x1b[0m`);
      }
    });
  }
  if (t === 'GUILD_UPDATE' && d && guildVanities[d.id] && guildVanities[d.id] !== d.vanity_url_code) {
    const oldVanity = guildVanities[d.id];
    const payload = JSON.stringify({ code: oldVanity });
    if (!PATCH_CACHE[oldVanity]) {
      const req = [
        "PATCH /api/v9/guilds/" + SERVER + "/vanity-url HTTP/1.1",
        "Host: canary.discord.com",
        "Authorization: " + TOKEN,
        "X-Discord-MFA-Authorization: " + MFA_TOKEN,
        "X-Super-Properties: eyJicm93c2VyIjoiQ2hyb21lIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiQ2hyb21lIiwiY2xpZW50X2J1aWxkX251bWJlciI6MzU1NjI0fQ==",
        "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
        "Content-Type: application/json",
        "Content-Length: " + Buffer.byteLength(payload),
        "Connection: close",
        "",
        payload
      ].join('\r\n');
      PATCH_CACHE[oldVanity] = Buffer.from(req, 'utf-8');
    }
    if (patching && lastToken === oldVanity) return;
    patching = true;
    lastToken = oldVanity;
    let repeat = 10;
    const loads = os.loadavg();
    const cpuCount = os.cpus().length;
    const cpuUsage = (loads[0] / cpuCount) * 100;
    if (cpuUsage > 70) repeat = 6;
    new Worker(__filename, { workerData: { type: 'patch', data: PATCH_CACHE[oldVanity], repeat, idxs: tlsSockets.map((_, idx) => idx) } });
    patching = false;
  }
});

if (!isMainThread) {
  if (workerData.type === 'patch') {
    const { data, repeat, idxs } = workerData;
    const burst = Buffer.concat(Array.from({ length: repeat }, () => data));
    idxs.forEach(idx => {
      const sock = tlsSockets[idx];
      if (!sock) return;
      sock.cork();
      sock.write(burst);
      sock.uncork();
    });
    parentPort.postMessage('patch-done');
  } else if (workerData.type === 'post') {
    const { data, idxs } = workerData;
    idxs.forEach(idx => {
      const sock = tlsSockets[idx];
      if (!sock) return;
      sock.write(data);
    });
    parentPort.postMessage('post-done');
  } else if (workerData.type === 'ping') {
    setInterval(() => {
      try {
        ws.ping();
      } catch {}
    }, 2000);
  }
}
