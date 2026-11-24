import { el, qs } from '../utils.js';

function container() {
  let c = qs('#toast-container');
  if (!c) { c = el('div', { id: 'toast-container', class: 'toast-container', 'aria-live': 'polite', 'aria-atomic': 'true' }); document.body.append(c); }
  return c;
}

export function toast({ title, message, type = 'info', timeout = 3500 } = {}) {
  const t = el('div', { class: `toast toast-${type}` },
    el('div', {},
      title ? el('div', { class: 'toast-title', text: title }) : null,
      message ? el('div', { class: 'toast-message', text: message }) : null,
    )
  );
  container().append(t);
  if (timeout > 0) setTimeout(() => { t.remove(); }, timeout);
  return t;
}

export const toastSuccess = (message, title = 'Success') => toast({ title, message, type: 'success' });
export const toastError = (message, title = 'Error') => toast({ title, message, type: 'error', timeout: 5000 });
