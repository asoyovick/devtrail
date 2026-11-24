import { el } from '../utils.js';
import { icon } from './icons.js';

export function loadingState(message = 'Loading...') {
  const wrap = el('div', { class: 'loading-state', role: 'status', 'aria-live': 'polite' });
  const spinner = el('div', { class: 'loading-spinner spin' });
  spinner.innerHTML = icon('spinner');
  wrap.append(spinner, el('div', { class: 'empty-description', text: message }));
  return wrap;
}

export function emptyState({ title = 'Nothing here yet', description = '', action = null, iconName = 'empty' } = {}) {
  const ic = el('div', { class: 'empty-icon' }); ic.innerHTML = icon(iconName);
  return el('div', { class: 'empty-state' },
    ic,
    el('div', { class: 'empty-title', text: title }),
    description ? el('div', { class: 'empty-description', text: description }) : null,
    action || null
  );
}

export function errorState({ title = 'Unable to load data', description = 'Something went wrong while loading this section.', onRetry = null } = {}) {
  const ic = el('div', { class: 'error-icon' }); ic.innerHTML = icon('error');
  const retry = onRetry ? el('button', { class: 'btn btn-sm', type: 'button', text: 'Retry' }) : null;
  if (retry) retry.addEventListener('click', onRetry);
  return el('div', { class: 'error-state', role: 'alert' },
    ic,
    el('div', { class: 'error-title', text: title }),
    description ? el('div', { class: 'error-description', text: description }) : null,
    retry
  );
}
