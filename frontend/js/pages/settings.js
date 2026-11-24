import { requireAuth, getUser } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, errorState } from '../components/states.js';
import { el, clear, safeText, fmtDate } from '../utils.js';

await requireAuth();
const { content } = mountShell({ page:'settings', title:'Settings', description:'Manage your profile and review integration and data preferences. Settings that are not yet supported by the backend are clearly marked.' });
clear(content).append(loadingState('Loading settings...'));
try {
  const me = await api.get('/api/v1/me');
  render(me);
} catch (err) {
  clear(content).append(errorState({ description: err.message, onRetry: () => location.reload() }));
}

function render(me) {
  clear(content);
  content.append(
    section('Profile', 'Your account information from the DevTrail backend.', el('div', { class:'card' }, el('div', { class:'card-body detail-grid' },
      detail('Display name', safeText(me.display_name)),
      detail('Email', safeText(me.email)),
      detail('User ID', safeText(me.id), true),
      detail('Created', fmtDate(me.created_at)),
      detail('Updated', fmtDate(me.updated_at)),
    ), el('div', { class:'card-footer' }, el('p', { class:'small muted', text:'Profile updates are not yet available through the backend API. This view reflects real account data from /api/v1/me.' })))),
    section('Preferences', 'Display and workspace preferences.', unavailable('Theme, timezone, and display preferences are not yet persisted by the backend API. No local-only fake persistence is enabled for these settings.')),
    section('Integrations', 'CLI and source integrations.', unavailable('CLI device management and integration status endpoints are not yet available in the backend API. Git activity can be viewed in the Git Activity page when records are present.')),
    section('Data and Privacy', 'Data collection and account controls.', unavailable('Data export, privacy controls, device revocation, and account deletion are not yet available in the backend API. These actions will be enabled when supported endpoints exist.'))
  );
}

function section(title, desc, body) {
  return el('section', { class:'section' }, el('div', { class:'section-header' }, el('div', { class:'section-header-left' }, el('h2', { class:'section-title', text:title }), el('p', { class:'section-subtitle', text:desc }))), body);
}
function unavailable(message) {
  return el('div', { class:'card' }, el('div', { class:'card-body' }, el('div', { class:'badge badge-warning', text:'Backend dependency' }), el('p', { class:'prose', style:'margin-top:0.75rem;', text:message })));
}
function detail(label, value, mono=false) { return el('div', { class:'detail-field' }, el('div', { class:'detail-label', text:label }), el('div', { class:`detail-value ${mono?'mono':''}`, text:safeText(value) })); }
