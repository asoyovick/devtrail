import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtDate, fmtRelative, fmtDuration, safeText, statusBadge, eventTypeMeta, getQuery } from '../utils.js';

await requireAuth();
const id = getQuery().id;
const shell = mountShell({ page: 'sessions', title: 'Session', description: 'Loading session...', breadcrumbs: [{label:'Sessions',href:'/pages/sessions.html'},{label:'Session'}] });
const content = shell.content;
if (!id) { clear(content).append(errorState({ title:'Missing session', description:'No session ID was provided.' })); } else load();

async function load() {
  clear(content).append(loadingState('Loading session...'));
  try {
    const session = await api.get(`/api/v1/sessions/${id}`);
    const [projectR, activityR] = await Promise.allSettled([
      api.get(`/api/v1/projects/${session.project_id}`),
      api.get('/api/v1/activity', { query: { project_id: session.project_id, limit: 100 } })
    ]);
    render(session, projectR.status==='fulfilled'?projectR.value:null, activityR.status==='fulfilled'?(activityR.value.events||[]):[]);
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render(s, project, events) {
  const st = statusBadge(s.status);
  shell.pageHeader.querySelector('.page-title').textContent = safeText(s.title, 'Untitled session');
  shell.pageHeader.querySelector('.page-description').textContent = project ? `Project: ${project.name}` : `Project ${s.project_id}`;
  const bc = shell.topbarRight.closest('.topbar').querySelector('.breadcrumbs');
  bc.innerHTML = ''; bc.append(el('a',{href:'/pages/sessions.html',text:'Sessions'}), el('span',{class:'breadcrumb-sep',text:'/'}), el('span',{class:'current',text:safeText(s.title,'Session')}));
  const sessionEvents = events.filter(e => e.session_id === s.id);
  clear(content);
  content.append(
    el('div', { class: 'card section' }, el('div', { class: 'card-body detail-grid' },
      detail('Status', el('span', { class: `badge ${st.cls}`, text: st.label })),
      detail('Project', project ? el('a',{class:'link',href:`/pages/project-details.html?id=${encodeURIComponent(project.id)}`,text:project.name}) : s.project_id),
      detail('Started', fmtDate(s.started_at)),
      detail('Ended', s.ended_at ? fmtDate(s.ended_at) : '—'),
      detail('Duration', fmtDuration(s.duration_seconds)),
      detail('Source', safeText(s.source)),
      detail('Session ID', s.id, true),
    )),
    el('section', { class: 'section' },
      el('div', { class: 'section-header' }, el('div', { class:'section-header-left' }, el('h2',{class:'section-title',text:'Session Activity'}), el('p',{class:'section-subtitle',text:'Events associated with this session from loaded project activity. A dedicated session-events endpoint would improve completeness.'}))),
      el('div', { class: 'card' }, sessionEvents.length ? el('div', { class:'list' }, ...sessionEvents.map(ev => {
        const m = eventTypeMeta(ev.event_type); const p = ev.payload || {};
        return el('div', { class:'list-item' }, el('div', { class:'list-body' }, el('div', { class:'list-title', text:m.label }), el('div', { class:'list-desc', text:safeText(p.title||p.message||p.summary||p.command||p.description||ev.event_type) }), el('div', { class:'list-meta' }, el('span', { class:`badge ${m.cls}`, text:m.label }), el('span', { text: fmtRelative(ev.occurred_at) }))));
      })) : emptyState({ title: 'No session events found', description: 'No events in the loaded project activity were linked to this session.' }))
    )
  );
}
function detail(label, value, mono=false) {
  const v = value?.nodeType ? value : el('div', { class: `detail-value ${mono?'mono':''}`, text: safeText(value) });
  return el('div', { class:'detail-field' }, el('div', { class:'detail-label', text: label }), v);
}
