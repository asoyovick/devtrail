import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtRelative, safeText, eventTypeMeta } from '../utils.js';

await requireAuth();
const { content } = mountShell({ page: 'activity', title: 'Activity Timeline', description: 'A chronological view of engineering events captured across your projects and sessions.' });

let projects = [], events = [];
const filters = { project_id: '', event_type: '' };
load();

async function load() {
  clear(content).append(loadingState('Loading activity...'));
  try {
    const [pRes, eRes] = await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/activity', { query: { project_id: filters.project_id, limit: 100 } })]);
    projects = pRes.projects || []; events = eRes.events || [];
    render();
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render() {
  clear(content);
  const typeOptions = ['', ...new Set(events.map(e => e.event_type).filter(Boolean))];
  const projectSel = el('select', { class: 'select', id: 'f-project' }, el('option', { value: '', text: 'All projects' }), ...projects.map(p => el('option', { value: p.id, text: p.name, selected: filters.project_id === p.id })));
  const typeSel = el('select', { class: 'select', id: 'f-type' }, el('option', { value: '', text: 'All loaded types' }), ...typeOptions.filter(Boolean).map(t => { const m = eventTypeMeta(t); return el('option', { value: t, text: m.label, selected: filters.event_type === t }); }));
  const apply = el('button', { class: 'btn btn-sm', type: 'button', text: 'Apply' });
  apply.addEventListener('click', () => { filters.project_id = projectSel.value; filters.event_type = typeSel.value; load(); });
  content.append(el('div', { class: 'filters' },
    el('div', { class: 'filter' }, el('label', { for: 'f-project', text: 'Project' }), projectSel),
    el('div', { class: 'filter' }, el('label', { for: 'f-type', text: 'Event type' }), typeSel),
    el('div', { class: 'filter' }, el('label', { text: '\u00a0' }), apply)
  ));

  const visible = filters.event_type ? events.filter(e => e.event_type === filters.event_type) : events;
  if (!visible.length) { content.append(el('div', { class: 'card' }, emptyState({ title: 'No activity found', description: 'No events match the current filters, or no activity has been recorded yet.' }))); return; }
  content.append(el('div', { class: 'card' }, el('div', { class: 'list' }, ...visible.map(ev => {
    const m = eventTypeMeta(ev.event_type);
    const p = ev.payload || {};
    const desc = safeText(p.title || p.message || p.summary || p.command || p.problem_description || p.description || '', ev.event_type);
    const project = projects.find(x => x.id === ev.project_id);
    return el('div', { class: 'list-item' },
      el('div', { class: 'list-body' },
        el('div', { class: 'list-title', text: m.label }),
        el('div', { class: 'list-desc', text: desc }),
        el('div', { class: 'list-meta' },
          el('span', { class: `badge ${m.cls}`, text: m.label }),
          project ? el('a', { href: `/pages/project-details.html?id=${encodeURIComponent(project.id)}`, text: project.name }) : (ev.project_id ? el('span', { text: `Project ${ev.project_id.slice(0,8)}` }) : null),
          ev.session_id ? el('a', { href: `/pages/session-details.html?id=${encodeURIComponent(ev.session_id)}`, text: `Session ${ev.session_id.slice(0,8)}` }) : null,
          el('span', { class: 'badge badge-neutral', text: safeText(ev.source, 'unknown') })
        )
      ),
      el('div', { class: 'list-aside', text: fmtRelative(ev.occurred_at) })
    );
  }))));
}
