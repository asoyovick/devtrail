import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { openModal, closeModal } from '../components/modal.js';
import { toastSuccess } from '../components/toast.js';
import { el, clear, fmtRelative, fmtDuration, safeText, statusBadge } from '../utils.js';
import { icon } from '../components/icons.js';

await requireAuth();
const { content, setActions } = mountShell({ page: 'sessions', title: 'Development Sessions', description: 'Track focused periods of development work, their status, and associated engineering activity.' });
const newBtn = el('button', { class: 'btn btn-primary', type: 'button' }); newBtn.innerHTML = icon('plus') + '<span>Start session</span>';
newBtn.addEventListener('click', openStart); setActions(newBtn);
let projects = [], sessions = [];
load();

async function load() {
  clear(content).append(loadingState('Loading sessions...'));
  try {
    const [p, s] = await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/sessions')]);
    projects = p.projects || []; sessions = s.sessions || []; render();
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render() {
  clear(content);
  if (!sessions.length) { content.append(el('div', { class: 'card' }, emptyState({ title: 'No sessions yet', description: 'Start a development session to capture focused engineering work.', action: (() => { const b = el('button', { class:'btn btn-primary', type:'button', text:'Start session' }); b.onclick = openStart; return b; })() }))); return; }
  const table = el('table', { class: 'table' },
    el('thead', {}, el('tr', {}, ...['Session','Project','Status','Duration','Started',''].map(h => el('th', { text: h })))),
    el('tbody', {}, ...sessions.map(s => {
      const st = statusBadge(s.status);
      const project = projects.find(p => p.id === s.project_id);
      const actions = el('td', { class: 'cell-actions' });
      actions.append(actionButtons(s));
      return el('tr', {},
        el('td', {}, el('a', { class: 'table-link', href: `/pages/session-details.html?id=${encodeURIComponent(s.id)}`, text: safeText(s.title, 'Untitled session') })),
        el('td', { text: project?.name || s.project_id.slice(0,8) }),
        el('td', {}, el('span', { class: `badge ${st.cls}`, text: st.label })),
        el('td', { text: fmtDuration(s.duration_seconds) }),
        el('td', { text: fmtRelative(s.started_at) }),
        actions
      );
    }))
  );
  content.append(el('div', { class: 'card card-table' }, el('div', { class: 'table-wrap' }, table)));
}

function actionButtons(s) {
  const wrap = el('div', { style: 'display:inline-flex;gap:0.25rem;' });
  const mk = (label, fn, show) => { if (!show) return; const b = el('button', { class:'btn btn-ghost btn-sm', type:'button', text: label }); b.onclick = fn; wrap.append(b); };
  mk('Pause', () => transition(s.id, 'pause'), s.status === 'active');
  mk('Resume', () => transition(s.id, 'resume'), s.status === 'paused');
  mk('End', () => transition(s.id, 'end'), s.status === 'active' || s.status === 'paused');
  return wrap;
}

async function transition(id, action) {
  try { await api.patch(`/api/v1/sessions/${id}/${action}`); toastSuccess(`Session ${action}d.`); await load(); }
  catch (err) { toastError(err.message || 'Action failed'); }
}

function openStart() {
  const project = el('select', { class: 'select', id: 's-project' }, el('option', { value: '', text: 'Select project' }), ...projects.map(p => el('option', { value: p.id, text: p.name })));
  const title = el('input', { class: 'input', id: 's-title', placeholder: 'Optional session title' });
  const errBox = el('div', { class: 'form-error' });
  const form = el('form', { class: 'form' }, errBox, field('s-project','Project',project), field('s-title','Title (optional)',title));
  const cancel = el('button', { class: 'btn', type:'button', text:'Cancel' }); cancel.onclick = closeModal;
  const submit = el('button', { class: 'btn btn-primary', type:'submit', text:'Start session' });
  form.addEventListener('submit', async (e) => {
    e.preventDefault(); errBox.textContent = '';
    if (!project.value) { errBox.textContent = 'Project is required.'; return; }
    submit.disabled = true; submit.textContent = 'Starting...';
    try { await api.post('/api/v1/sessions', { project_id: project.value, title: title.value.trim(), source: 'web' }); closeModal(); toastSuccess('Session started.'); await load(); }
    catch (err) { errBox.textContent = err.message; submit.disabled = false; submit.textContent = 'Start session'; }
  });
  openModal({ title: 'Start session', description: 'Begin a focused development session for a project.', content: form, actions: [cancel, submit] });
  submit.onclick = () => form.requestSubmit();
}
function field(id, label, input) { return el('div', { class: 'field' }, el('label', { for: id, text: label }), input); }
