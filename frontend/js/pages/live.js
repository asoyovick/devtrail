import { requireAuth, getToken } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtRelative, safeText, eventTypeMeta } from '../utils.js';

await requireAuth();
const { content, setActions } = mountShell({
  page: 'live',
  title: 'Live Development',
  description: 'Watch development events arrive in real time from connected local agents and synchronized sources.',
});

const API_BASE = location.protocol === 'file:' ? 'http://localhost:8080' : (location.port && location.port !== '8080' ? 'http://localhost:8080' : '');
let events = [];
let paused = false;
let connected = false;
let abort = null;
let reconnectTimer = null;

const statusBadge = el('span', { class: 'badge badge-neutral', text: 'Connecting...' });
const pauseBtn = el('button', { class: 'btn btn-sm', type: 'button', text: 'Pause updates' });
pauseBtn.onclick = () => { paused = !paused; pauseBtn.textContent = paused ? 'Resume updates' : 'Pause updates'; statusBadge.textContent = paused ? 'Paused' : (connected ? 'Live' : 'Reconnecting...'); statusBadge.className = `badge ${paused ? 'badge-warning' : (connected ? 'badge-success' : 'badge-neutral')}`; };
setActions(el('div', { style: 'display:flex;gap:0.5rem;align-items:center;' }, statusBadge, pauseBtn));

const listWrap = el('div', { class: 'card' });
const feed = el('div', { class: 'list' });
listWrap.append(feed);
content.append(listWrap);
clear(feed).append(loadingState('Loading recent activity...'));

await loadInitial();
connect();

async function loadInitial() {
  try {
    const res = await api.get('/api/v1/activity', { query: { limit: 50 } });
    events = (res.events || []).map(normalize).sort((a,b)=>new Date(b.occurred_at)-new Date(a.occurred_at));
    render();
  } catch (err) {
    clear(feed).append(errorState({ description: err.message, onRetry: loadInitial }));
  }
}

function normalize(ev) {
  return {
    id: ev.id || ev.event_id,
    type: ev.type || ev.event_type,
    project_id: ev.project_id,
    session_id: ev.session_id,
    source: ev.source || 'unknown',
    occurred_at: ev.occurred_at,
    payload: ev.payload || {},
  };
}

function render() {
  clear(feed);
  if (!events.length) { feed.append(emptyState({ title: 'No live activity yet', description: 'Connect a local directory with the DevTrail CLI and run `devtrail watch` to see events appear here.' })); return; }
  feed.append(...events.slice(0, 100).map(renderEvent));
}

function renderEvent(ev) {
  const meta = eventTypeMeta(ev.type);
  const src = sourceBadge(ev.source);
  const p = ev.payload || {};
  const desc = safeText(p.path || p.title || p.message || p.summary || p.command || p.description || ev.type, ev.type);
  return el('div', { class: 'list-item' },
    el('div', { class: 'list-body' },
      el('div', { style: 'display:flex;gap:0.5rem;align-items:center;flex-wrap:wrap;' }, el('div', { class: 'list-title', text: meta.label }), el('span', { class: `badge ${src.cls}`, text: src.label })),
      el('div', { class: 'list-desc', text: desc }),
      el('div', { class: 'list-meta' }, el('span', { class: `badge ${meta.cls}`, text: meta.label }), ev.project_id ? el('span', { text: `Project ${String(ev.project_id).slice(0,8)}` }) : null, ev.session_id ? el('span', { text: `Session ${String(ev.session_id).slice(0,8)}` }) : null)
    ),
    el('div', { class: 'list-aside', text: fmtRelative(ev.occurred_at) })
  );
}

function sourceBadge(source) {
  const s = String(source || '').toLowerCase();
  if (s === 'local_agent' || s === 'cli') return { label: 'Local', cls: 'badge-accent' };
  if (s === 'github') return { label: 'GitHub', cls: 'badge-violet' };
  if (s === 'system' || s === 'calculated') return { label: 'Calculated', cls: 'badge-neutral' };
  return { label: 'Synced', cls: 'badge-neutral' };
}

async function connect() {
  clearReconnect();
  abort = new AbortController();
  const token = getToken();
  if (!token) return;
  try {
    const res = await fetch(`${API_BASE}/api/v1/stream/activity`, {
      headers: { 'Authorization': `Bearer ${token}`, 'Accept': 'text/event-stream' },
      signal: abort.signal,
    });
    if (!res.ok || !res.body) throw new Error(`Stream failed (${res.status})`);
    connected = true; if (!paused) { statusBadge.textContent = 'Live'; statusBadge.className = 'badge badge-success'; }
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buf = '';
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      let idx;
      while ((idx = buf.indexOf('\n\n')) >= 0) {
        const chunk = buf.slice(0, idx); buf = buf.slice(idx + 2);
        handleChunk(chunk);
      }
    }
  } catch (err) {
    if (abort?.signal.aborted) return;
    connected = false; if (!paused) { statusBadge.textContent = 'Reconnecting...'; statusBadge.className = 'badge badge-warning'; }
  }
  connected = false;
  if (!paused) { statusBadge.textContent = 'Reconnecting...'; statusBadge.className = 'badge badge-warning'; }
  reconnectTimer = setTimeout(connect, 3000);
}

function handleChunk(chunk) {
  const lines = chunk.split('\n');
  let eventName = 'message'; let data = '';
  for (const line of lines) {
    if (line.startsWith('event:')) eventName = line.slice(6).trim();
    if (line.startsWith('data:')) data += line.slice(5).trim();
  }
  if (!data || eventName !== 'activity') return;
  try {
    const ev = normalize(JSON.parse(data));
    if (events.some(e => e.id === ev.id)) return;
    events.unshift(ev);
    if (!paused) render();
  } catch {}
}

function clearReconnect() { if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; } }
window.addEventListener('beforeunload', () => { clearReconnect(); abort?.abort(); });
