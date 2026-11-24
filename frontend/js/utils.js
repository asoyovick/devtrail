export function qs(sel, root = document) { return root.querySelector(sel); }
export function qsa(sel, root = document) { return [...root.querySelectorAll(sel)]; }

export function el(tag, attrs = {}, ...children) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs || {})) {
    if (v == null || v === false) continue;
    if (k === 'class' || k === 'className') node.className = v;
    else if (k === 'text') node.textContent = v;
    else if (k === 'html') { /* intentionally not supported for safety */ }
    else if (k === 'dataset') Object.assign(node.dataset, v);
    else if (k.startsWith('on') && typeof v === 'function') node.addEventListener(k.slice(2).toLowerCase(), v);
    else if (k === 'checked' || k === 'disabled' || k === 'selected' || k === 'hidden') node[k] = !!v;
    else node.setAttribute(k, v === true ? '' : String(v));
  }
  for (const child of children.flat()) {
    if (child == null || child === false) continue;
    node.append(child.nodeType ? child : document.createTextNode(String(child)));
  }
  return node;
}

export function clear(node) { while (node.firstChild) node.removeChild(node.firstChild); return node; }

export function fmtDate(value, opts = {}) {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return String(value);
  return d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short', ...opts });
}

export function fmtDateOnly(value) {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return String(value);
  return d.toLocaleDateString(undefined, { dateStyle: 'medium' });
}

export function fmtRelative(value) {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return String(value);
  const diff = Date.now() - d.getTime();
  const abs = Math.abs(diff);
  const mins = Math.round(abs / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ${diff >= 0 ? 'ago' : 'from now'}`;
  const hrs = Math.round(mins / 60);
  if (hrs < 24) return `${hrs}h ${diff >= 0 ? 'ago' : 'from now'}`;
  const days = Math.round(hrs / 24);
  if (days < 30) return `${days}d ${diff >= 0 ? 'ago' : 'from now'}`;
  return fmtDateOnly(value);
}

export function fmtDuration(seconds) {
  if (seconds == null || Number.isNaN(Number(seconds))) return '—';
  const s = Math.max(0, Math.floor(Number(seconds)));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m`;
  return `${s}s`;
}

export function shortId(id) {
  if (!id) return '—';
  const s = String(id);
  return s.length > 12 ? `${s.slice(0, 8)}` : s;
}

export function escapeText(s) { return s == null ? '' : String(s); }

export function getQuery() {
  const params = new URLSearchParams(location.search);
  return Object.fromEntries(params.entries());
}

export function setQuery(updates, { replace = true } = {}) {
  const params = new URLSearchParams(location.search);
  for (const [k, v] of Object.entries(updates)) {
    if (v == null || v === '') params.delete(k);
    else params.set(k, v);
  }
  const url = `${location.pathname}${params.toString() ? `?${params}` : ''}${location.hash}`;
  replace ? history.replaceState(null, '', url) : history.pushState(null, '', url);
}

export function debounce(fn, ms = 250) {
  let t; return (...args) => { clearTimeout(t); t = setTimeout(() => fn(...args), ms); };
}

export function titleCase(s) {
  return String(s || '').replace(/[_-]+/g, ' ').replace(/\b\w/g, c => c.toUpperCase()).trim();
}

export function statusBadge(status) {
  const s = String(status || '').toLowerCase();
  const map = {
    active: 'badge-success', completed: 'badge-neutral', ended: 'badge-neutral', paused: 'badge-warning',
    open: 'badge-warning', resolved: 'badge-success', abandoned: 'badge-neutral',
    passed: 'badge-success', failed: 'badge-danger', errored: 'badge-danger', skipped: 'badge-neutral', unknown: 'badge-neutral',
    accepted: 'badge-success', proposed: 'badge-accent', superseded: 'badge-neutral', deprecated: 'badge-neutral', rejected: 'badge-danger',
    archived: 'badge-neutral',
  };
  return { label: titleCase(s || 'unknown'), cls: map[s] || 'badge-neutral' };
}

export function eventTypeMeta(type) {
  const t = String(type || '');
  const map = {
    'git.commit': { label: 'Git Commit', cls: 'badge-accent' },
    'git.branch_changed': { label: 'Branch Change', cls: 'badge-violet' },
    'file.changed': { label: 'File Change', cls: 'badge-neutral' },
    'session.started': { label: 'Session Started', cls: 'badge-success' },
    'session.paused': { label: 'Session Paused', cls: 'badge-warning' },
    'session.resumed': { label: 'Session Resumed', cls: 'badge-success' },
    'session.ended': { label: 'Session Ended', cls: 'badge-neutral' },
    'test.run': { label: 'Test Run', cls: 'badge-accent' },
    'investigation.created': { label: 'Investigation', cls: 'badge-warning' },
    'investigation.updated': { label: 'Investigation Updated', cls: 'badge-warning' },
    'decision.created': { label: 'Decision', cls: 'badge-violet' },
    'note.created': { label: 'Note', cls: 'badge-neutral' },
    'project.registered': { label: 'Project', cls: 'badge-accent' },
  };
  return map[t] || { label: titleCase(t || 'event'), cls: 'badge-neutral' };
}

export function safeText(value, fallback = '—') {
  if (value == null) return fallback;
  const s = String(value).trim();
  return s || fallback;
}
