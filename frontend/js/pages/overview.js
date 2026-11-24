import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtRelative, fmtDuration, safeText, statusBadge, eventTypeMeta } from '../utils.js';

await requireAuth();
const { content } = mountShell({
  page: 'overview',
  title: 'Overview',
  description: 'Your engineering command centre. Track recent work, sessions, investigations, and emerging patterns across projects.',
});

init();

async function init() {
  content.append(loadingState('Loading workspace overview...'));
  const [projectsR, sessionsR, activityR, investigationsR, testRunsR, insightsR] = await Promise.allSettled([
    api.get('/api/v1/projects'),
    api.get('/api/v1/sessions'),
    api.get('/api/v1/activity', { query: { limit: 8 } }),
    api.get('/api/v1/investigations'),
    api.get('/api/v1/test-runs'),
    api.get('/api/v1/insights'),
  ]);
  clear(content);

  const projects = val(projectsR, 'projects', []);
  const sessions = val(sessionsR, 'sessions', []);
  const events = val(activityR, 'events', []);
  const investigations = val(investigationsR, 'investigations', []);
  const testRuns = val(testRunsR, 'test_runs', []);
  const insights = val(insightsR, 'insights', []);

  const stats = [
    { label: 'Projects', value: projects.length, meta: 'Registered workspaces' },
    { label: 'Sessions', value: sessions.length, meta: 'Development sessions' },
    { label: 'Activities', value: events.length + (activityR.status === 'fulfilled' ? '+' : ''), meta: 'Recent loaded events' },
    { label: 'Investigations', value: investigations.length, meta: 'Debugging records' },
    { label: 'Test Runs', value: testRuns.length, meta: 'Recorded executions' },
  ];

  content.append(
    el('section', { class: 'section' },
      el('div', { class: 'grid grid-5' }, ...stats.map(s => el('div', { class: 'card stat-card' },
        el('div', { class: 'stat-label', text: s.label }),
        el('div', { class: 'stat-value', text: String(s.value) }),
        el('div', { class: 'stat-meta', text: s.meta }),
      )))
    ),
    el('div', { class: 'split' },
      el('section', { class: 'section' }, sectionHead('Recent Activity', 'Latest engineering events across your workspace', '/pages/activity.html'), cardOrState(activityR, events.length ? activityList(events) : emptyState({ title: 'No activity yet', description: 'Activity will appear here as sessions, commits, tests, and investigations are recorded.', action: linkBtn('/pages/activity.html', 'View activity') }))),
      el('section', { class: 'section' }, sectionHead('Engineering Observations', 'Rule-based patterns from your records', '/pages/insights.html'), cardOrState(insightsR, insights.length ? insightList(insights.slice(0, 4)) : emptyState({ title: 'No observations yet', description: 'Insights appear when enough engineering history is available for rule-based analysis.', action: linkBtn('/pages/insights.html', 'View insights') })))
    ),
    el('div', { class: 'split' },
      el('section', { class: 'section' }, sectionHead('Active Projects', 'Projects with recent engineering context', '/pages/projects.html'), cardOrState(projectsR, projects.length ? projectList(projects.slice(0, 5)) : emptyState({ title: 'No projects yet', description: 'Register a project to start tracking engineering history.', action: linkBtn('/pages/projects.html', 'Create project') }))),
      el('section', { class: 'section' }, sectionHead('Recent Sessions', 'Latest development sessions', '/pages/sessions.html'), cardOrState(sessionsR, sessions.length ? sessionList(sessions.slice(0, 5)) : emptyState({ title: 'No sessions yet', description: 'Start a development session to capture focused work.', action: linkBtn('/pages/sessions.html', 'View sessions') })))
    ),
    el('section', { class: 'section' }, sectionHead('Recent Debugging Investigations', 'Latest problem-solving records', '/pages/debugging.html'), cardOrState(investigationsR, investigations.length ? investigationList(investigations.slice(0, 5)) : emptyState({ title: 'No investigations yet', description: 'Document your first investigation to preserve debugging knowledge.', action: linkBtn('/pages/debugging.html', 'Open debugging journal') }))),
  );
}

function val(result, key, fallback) { return result.status === 'fulfilled' ? (result.value?.[key] ?? fallback) : fallback; }
function sectionHead(title, subtitle, href) {
  return el('div', { class: 'section-header' },
    el('div', { class: 'section-header-left' }, el('h2', { class: 'section-title', text: title }), el('p', { class: 'section-subtitle', text: subtitle })),
    href ? el('a', { class: 'btn btn-ghost btn-sm', href, text: 'View all' }) : null
  );
}
function linkBtn(href, text) { return el('a', { class: 'btn btn-sm', href, text }); }
function cardOrState(result, node) {
  const card = el('div', { class: 'card' });
  if (result.status === 'rejected') card.append(errorState({ description: result.reason?.message || 'Failed to load this section.' }));
  else card.append(node);
  return card;
}
function activityList(events) {
  return el('div', { class: 'list' }, ...events.map(ev => {
    const meta = eventTypeMeta(ev.event_type);
    const item = el('div', { class: 'list-item' },
      el('div', { class: 'list-icon' }, el('span', { class: 'dot', style: 'width:8px;height:8px;' })),
      el('div', { class: 'list-body' },
        el('div', { class: 'list-title', text: meta.label }),
        el('div', { class: 'list-desc', text: describeEvent(ev) }),
        el('div', { class: 'list-meta' },
          el('span', { class: `badge ${meta.cls}`, text: meta.label }),
          ev.project_id ? el('span', { text: `Project ${short(ev.project_id)}` }) : null,
          ev.session_id ? el('span', { text: `Session ${short(ev.session_id)}` }) : null,
        )
      ),
      el('div', { class: 'list-aside', text: fmtRelative(ev.occurred_at) })
    );
    return item;
  }));
}
function describeEvent(ev) {
  const p = ev.payload || {};
  if (typeof p === 'string') return p;
  return safeText(p.title || p.message || p.summary || p.command || p.problem_description || p.description || ev.event_type, ev.event_type);
}
function projectList(projects) {
  return el('div', { class: 'list' }, ...projects.map(p => {
    const s = statusBadge(p.status);
    return el('a', { class: 'list-item list-item-interactive', href: `/pages/project-details.html?id=${encodeURIComponent(p.id)}` },
      el('div', { class: 'list-body' },
        el('div', { class: 'list-title', text: safeText(p.name) }),
        el('div', { class: 'list-desc', text: safeText(p.description, 'No description') }),
        el('div', { class: 'list-meta' }, el('span', { class: `badge ${s.cls}`, text: s.label }), p.default_branch ? el('span', { text: `Branch ${p.default_branch}` }) : null)
      ),
      el('div', { class: 'list-aside', text: fmtRelative(p.updated_at || p.created_at) })
    );
  }));
}
function sessionList(sessions) {
  return el('div', { class: 'list' }, ...sessions.map(s => {
    const st = statusBadge(s.status);
    return el('a', { class: 'list-item list-item-interactive', href: `/pages/session-details.html?id=${encodeURIComponent(s.id)}` },
      el('div', { class: 'list-body' },
        el('div', { class: 'list-title', text: safeText(s.title, 'Untitled session') }),
        el('div', { class: 'list-desc', text: `Project ${short(s.project_id)}` }),
        el('div', { class: 'list-meta' }, el('span', { class: `badge ${st.cls}`, text: st.label }), el('span', { text: fmtDuration(s.duration_seconds) }))
      ),
      el('div', { class: 'list-aside', text: fmtRelative(s.started_at) })
    );
  }));
}
function investigationList(items) {
  return el('div', { class: 'list' }, ...items.map(i => {
    const st = statusBadge(i.status);
    return el('a', { class: 'list-item list-item-interactive', href: `/pages/debugging-details.html?id=${encodeURIComponent(i.id)}` },
      el('div', { class: 'list-body' },
        el('div', { class: 'list-title', text: safeText(i.title) }),
        el('div', { class: 'list-desc', text: safeText(i.problem_description) }),
        el('div', { class: 'list-meta' }, el('span', { class: `badge ${st.cls}`, text: st.label }), el('span', { text: `Project ${short(i.project_id)}` }))
      ),
      el('div', { class: 'list-aside', text: fmtRelative(i.updated_at || i.created_at) })
    );
  }));
}
function insightList(items) {
  return el('div', { class: 'list' }, ...items.map(i => el('div', { class: 'list-item' },
    el('div', { class: 'list-body' },
      el('div', { class: 'list-title', text: safeText(i.summary, i.type) }),
      el('div', { class: 'list-desc', text: safeText(i.observation, '') }),
      el('div', { class: 'list-meta' }, el('span', { class: 'badge badge-accent', text: safeText(i.type, 'Insight') }), el('span', { class: 'badge badge-neutral', text: `Confidence: ${safeText(i.confidence, 'unknown')}` }))
    )
  )));
}
function short(id) { return id ? String(id).slice(0, 8) : '—'; }
