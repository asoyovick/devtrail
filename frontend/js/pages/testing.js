import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtRelative, fmtDate, safeText, statusBadge } from '../utils.js';

await requireAuth();
const { content } = mountShell({ page: 'testing', title: 'Testing History', description: 'Review test executions, outcomes, and failure summaries across your projects.' });
let projects = [], runs = []; const filters = { project_id: '', status: '' };
load();

async function load() {
  clear(content).append(loadingState('Loading test runs...'));
  try {
    const [p, t] = await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/test-runs', { query: { project_id: filters.project_id } })]);
    projects = p.projects || []; runs = t.test_runs || []; render();
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render() {
  clear(content);
  const pSel = el('select', { class:'select' }, el('option', { value:'', text:'All projects' }), ...projects.map(p => el('option', { value:p.id, text:p.name, selected: filters.project_id===p.id })));
  const sSel = el('select', { class:'select' }, el('option', { value:'', text:'All statuses' }), ...['passed','failed','errored','skipped','unknown'].map(s => el('option', { value:s, text:s, selected: filters.status===s })));
  const apply = el('button', { class:'btn btn-sm', type:'button', text:'Apply' });
  apply.onclick = () => { filters.project_id = pSel.value; filters.status = sSel.value; render(); };
  content.append(el('div', { class:'filters' },
    el('div', { class:'filter' }, el('label',{text:'Project'}), pSel),
    el('div', { class:'filter' }, el('label',{text:'Status'}), sSel),
    el('div', { class:'filter' }, el('label',{text:'\u00a0'}), apply)
  ));
  const visible = filters.status ? runs.filter(r => r.status === filters.status) : runs;
  if (!visible.length) { content.append(el('div', { class:'card' }, emptyState({ title:'No test runs found', description:'Test executions recorded for your projects will appear here.' }))); return; }
  content.append(el('div', { class:'card card-table' }, el('div', { class:'table-wrap' }, el('table', { class:'table' },
    el('thead', {}, el('tr', {}, ...['Command','Project','Status','Results','Duration','Started'].map(h=>el('th',{text:h})))),
    el('tbody', {}, ...visible.map(r => {
      const st = statusBadge(r.status);
      const project = projects.find(p => p.id === r.project_id);
      const results = r.total_tests != null ? `${r.passed_tests ?? 0}/${r.total_tests} passed` : (r.failed_tests != null ? `${r.failed_tests} failed` : '—');
      return el('tr', {},
        el('td', {}, el('div', { class:'cell-primary mono', text: safeText(r.command) }), r.failure_summary ? el('div', { class:'cell-secondary', text: safeText(r.failure_summary).slice(0,120) }) : null),
        el('td', { text: project?.name || String(r.project_id||'').slice(0,8) }),
        el('td', {}, el('span', { class:`badge ${st.cls}`, text: st.label })),
        el('td', { text: results }),
        el('td', { text: r.duration_ms != null ? `${Math.round(r.duration_ms/1000)}s` : '—' }),
        el('td', {}, el('div', { text: fmtRelative(r.started_at) }), el('div', { class:'cell-secondary', text: fmtDate(r.started_at) }))
      );
    }))
  ))));
}
