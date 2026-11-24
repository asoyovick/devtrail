import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtRelative, fmtDate, fmtDuration, safeText, statusBadge, eventTypeMeta, getQuery } from '../utils.js';

await requireAuth();
const id = getQuery().id;
const shell = mountShell({ page: 'projects', title: 'Project', description: 'Loading project...', breadcrumbs: [{label:'Projects', href:'/pages/projects.html'}, {label:'Project'}] });
const content = shell.content;
if (!id) { clear(content).append(errorState({ title: 'Missing project', description: 'No project ID was provided.' })); }
else load();

async function load() {
  clear(content).append(loadingState('Loading project...'));
  try {
    const [project, sessionsR, activityR, investigationsR, decisionsR, testRunsR, commitsR, notesR, insightsR] = await Promise.all([
      api.get(`/api/v1/projects/${id}`),
      api.get('/api/v1/sessions', { query: { project_id: id } }).catch(e => ({ error: e })),
      api.get('/api/v1/activity', { query: { project_id: id, limit: 20 } }).catch(e => ({ error: e })),
      api.get('/api/v1/investigations', { query: { project_id: id } }).catch(e => ({ error: e })),
      api.get('/api/v1/decisions', { query: { project_id: id } }).catch(e => ({ error: e })),
      api.get('/api/v1/test-runs', { query: { project_id: id } }).catch(e => ({ error: e })),
      api.get('/api/v1/git/commits', { query: { project_id: id } }).catch(e => ({ error: e })),
      api.get('/api/v1/notes', { query: { project_id: id } }).catch(e => ({ error: e })),
      api.get('/api/v1/insights').catch(e => ({ error: e })),
    ]);
    render(project, { sessionsR, activityR, investigationsR, decisionsR, testRunsR, commitsR, notesR, insightsR });
  } catch (err) {
    clear(content).append(errorState({ description: err.message, onRetry: load }));
  }
}

function arr(res, key) { return res?.error ? [] : (res?.[key] || []); }
function err(res) { return res?.error; }

function render(project, data) {
  const st = statusBadge(project.status);
  shell.pageHeader.querySelector('.page-title').textContent = project.name;
  shell.pageHeader.querySelector('.page-description').textContent = project.description || 'Project engineering workspace';
  const bc = shell.topbarRight.closest('.topbar').querySelector('.breadcrumbs');
  bc.innerHTML = ''; bc.append(el('a', { href: '/pages/projects.html', text: 'Projects' }), el('span', { class: 'breadcrumb-sep', text: '/' }), el('span', { class: 'current', text: project.name }));

  const sessions = arr(data.sessionsR, 'sessions');
  const events = arr(data.activityR, 'events');
  const investigations = arr(data.investigationsR, 'investigations');
  const decisions = arr(data.decisionsR, 'decisions');
  const testRuns = arr(data.testRunsR, 'test_runs');
  const commits = arr(data.commitsR, 'commits');
  const notes = arr(data.notesR, 'notes');
  const insights = arr(data.insightsR, 'insights');

  const tabs = [
    ['overview', 'Overview'], ['activity', 'Activity'], ['sessions', 'Sessions'], ['git', 'Git History'],
    ['debugging', 'Debugging'], ['decisions', 'Decisions'], ['testing', 'Testing'], ['insights', 'Insights'], ['notes', 'Notes']
  ];
  let active = 'overview';
  const tabBar = el('div', { class: 'tabs', role: 'tablist' });
  const panel = el('div');
  function renderTab() {
    clear(panel);
    if (active === 'overview') panel.append(overviewTab(project, { sessions, events, investigations, decisions, testRuns, commits, notes, st }));
    if (active === 'activity') panel.append(listCard(events, 'No activity yet', 'Activity for this project will appear here.', e => activityRow(e)));
    if (active === 'sessions') panel.append(listCard(sessions, 'No sessions yet', 'Development sessions for this project will appear here.', s => sessionRow(s)));
    if (active === 'git') panel.append(listCard(commits, 'No git activity yet', 'Commits recorded for this project will appear here.', c => commitRow(c)));
    if (active === 'debugging') panel.append(listCard(investigations, 'No investigations yet', 'Debugging investigations for this project will appear here.', i => investigationRow(i)));
    if (active === 'decisions') panel.append(listCard(decisions, 'No decisions yet', 'Engineering decisions for this project will appear here.', d => decisionRow(d)));
    if (active === 'testing') panel.append(listCard(testRuns, 'No test runs yet', 'Test executions for this project will appear here.', t => testRow(t)));
    if (active === 'insights') panel.append(listCard(insights, 'No insights yet', 'Global insights are shown here; project-scoped insight filtering depends on backend support.', i => insightRow(i)));
    if (active === 'notes') panel.append(listCard(notes, 'No notes yet', 'Technical notes for this project will appear here.', n => noteRow(n)));
  }
  for (const [key, label] of tabs) {
    const b = el('button', { class: 'tab', role: 'tab', type: 'button', text: label, 'aria-selected': key === active ? 'true' : 'false' });
    b.addEventListener('click', () => { active = key; [...tabBar.children].forEach(c => c.setAttribute('aria-selected', 'false')); b.setAttribute('aria-selected', 'true'); renderTab(); });
    tabBar.append(b);
  }
  clear(content);
  content.append(
    el('div', { class: 'card section', style: 'margin-bottom:1rem;' }, el('div', { class: 'card-body detail-grid' },
      detail('Status', el('span', { class: `badge ${st.cls}`, text: st.label })),
      detail('Default branch', safeText(project.default_branch)),
      detail('Created', fmtDate(project.created_at)),
      detail('Last updated', fmtDate(project.updated_at)),
      detail('Description', safeText(project.description, 'No description')),
      detail('Project ID', project.id, true),
    )),
    tabBar, panel
  );
  renderTab();
}

function detail(label, value, mono=false) {
  const v = value?.nodeType ? value : el('div', { class: `detail-value ${mono ? 'mono' : ''}`, text: safeText(value) });
  return el('div', { class: 'detail-field' }, el('div', { class: 'detail-label', text: label }), v);
}
function overviewTab(project, d) {
  return el('div', { class: 'grid grid-4' },
    ...[['Sessions', d.sessions.length], ['Events', d.events.length], ['Investigations', d.investigations.length], ['Decisions', d.decisions.length], ['Commits', d.commits.length], ['Test runs', d.testRuns.length], ['Notes', d.notes.length]].map(([l,v]) =>
      el('div', { class: 'card stat-card' }, el('div', { class: 'stat-label', text: l }), el('div', { class: 'stat-value', style:'font-size:1.5rem;', text: String(v) }))
    )
  );
}
function listCard(items, emptyTitle, emptyDesc, rowFn) {
  const card = el('div', { class: 'card' });
  if (!items.length) return card.append(emptyState({ title: emptyTitle, description: emptyDesc })), card;
  card.append(el('div', { class: 'list' }, ...items.map(rowFn)));
  return card;
}
function activityRow(ev) { const m = eventTypeMeta(ev.event_type); return el('div', { class: 'list-item' }, el('div', { class: 'list-body' }, el('div', { class: 'list-title', text: m.label }), el('div', { class: 'list-desc', text: describeEvent(ev) }), el('div', { class: 'list-meta' }, el('span', { class: `badge ${m.cls}`, text: m.label }), el('span', { text: fmtRelative(ev.occurred_at) })))); }
function describeEvent(ev) { const p = ev.payload || {}; return safeText(p.title || p.message || p.summary || p.command || p.problem_description || p.description || ev.event_type, ev.event_type); }
function sessionRow(s) { const st = statusBadge(s.status); return el('a', { class: 'list-item list-item-interactive', href: `/pages/session-details.html?id=${encodeURIComponent(s.id)}` }, el('div', { class: 'list-body' }, el('div', { class: 'list-title', text: safeText(s.title,'Untitled session') }), el('div', { class: 'list-meta' }, el('span', { class:`badge ${st.cls}`, text: st.label }), el('span', { text: fmtDuration(s.duration_seconds) }), el('span', { text: fmtRelative(s.started_at) })))); }
function commitRow(c) { return el('div', { class: 'list-item' }, el('div', { class: 'list-body' }, el('div', { class: 'list-title', text: safeText(c.message_summary, c.commit_hash) }), el('div', { class: 'list-meta' }, el('span', { class: 'cell-mono', text: String(c.commit_hash||'').slice(0,8) }), c.branch ? el('span', { class:'badge badge-accent', text: c.branch }) : null, el('span', { text: fmtRelative(c.committed_at) })))); }
function investigationRow(i) { const st = statusBadge(i.status); return el('a', { class:'list-item list-item-interactive', href:`/pages/debugging-details.html?id=${encodeURIComponent(i.id)}` }, el('div', { class:'list-body' }, el('div', { class:'list-title', text: safeText(i.title) }), el('div', { class:'list-desc', text: safeText(i.problem_description) }), el('div', { class:'list-meta' }, el('span', { class:`badge ${st.cls}`, text: st.label }), el('span', { text: fmtRelative(i.updated_at) })))); }
function decisionRow(d) { const st = statusBadge(d.status); return el('a', { class:'list-item list-item-interactive', href:`/pages/decision-details.html?id=${encodeURIComponent(d.id)}` }, el('div', { class:'list-body' }, el('div', { class:'list-title', text: safeText(d.title) }), el('div', { class:'list-desc', text: safeText(d.problem) }), el('div', { class:'list-meta' }, el('span', { class:`badge ${st.cls}`, text: st.label }), el('span', { text: fmtRelative(d.decided_at||d.created_at) })))); }
function testRow(t) { const st = statusBadge(t.status); return el('div', { class:'list-item' }, el('div', { class:'list-body' }, el('div', { class:'list-title', text: safeText(t.command) }), el('div', { class:'list-meta' }, el('span', { class:`badge ${st.cls}`, text: st.label }), t.failed_tests != null ? el('span', { text:`${t.passed_tests ?? 0} passed / ${t.failed_tests ?? 0} failed` }) : null, el('span', { text: fmtRelative(t.started_at) })))); }
function insightRow(i) { return el('div', { class:'list-item' }, el('div', { class:'list-body' }, el('div', { class:'list-title', text: safeText(i.summary,i.type) }), el('div', { class:'list-desc', text: safeText(i.observation) }), el('div', { class:'list-meta' }, el('span', { class:'badge badge-accent', text: safeText(i.type) }), el('span', { class:'badge badge-neutral', text:`Confidence: ${safeText(i.confidence,'unknown')}` })))); }
function noteRow(n) { return el('div', { class:'list-item' }, el('div', { class:'list-body' }, el('div', { class:'list-title', text: safeText(n.title) }), el('div', { class:'list-desc', text: safeText(n.body).slice(0,160) }), el('div', { class:'list-meta' }, ...(n.tags||[]).slice(0,3).map(t => el('span', { class:'badge badge-neutral', text: t })), el('span', { text: fmtRelative(n.updated_at) })))); }
