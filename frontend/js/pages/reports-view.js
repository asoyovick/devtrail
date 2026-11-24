import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, safeText, fmtDate } from '../utils.js';

export async function initReportPage({ page, title, description, type }) {
  await requireAuth();
  const { content } = mountShell({ page, title, description });
  let projects = [];
  try { projects = (await api.get('/api/v1/projects')).projects || []; } catch {}
  const date = el('input', { class:'input', type:'date', id:'r-date', value: new Date().toISOString().slice(0,10) });
  const project = el('select', { class:'select', id:'r-project' }, el('option', { value:'', text:'Select project' }), ...projects.map(p => el('option', { value:p.id, text:p.name })));
  const run = el('button', { class:'btn btn-primary btn-sm', type:'button', text:'Generate report' });
  const out = el('div');
  const controls = el('div', { class:'filters' });
  if (type !== 'project') controls.append(el('div', { class:'filter' }, el('label', { for:'r-date', text:'Date' }), date));
  if (type === 'project') controls.append(el('div', { class:'filter filter-grow' }, el('label', { for:'r-project', text:'Project' }), project));
  controls.append(el('div', { class:'filter' }, el('label', { text:'\u00a0' }), run));
  content.append(controls, out);
  run.onclick = () => generate();
  if (type !== 'project') generate();

  async function generate() {
    clear(out).append(loadingState('Generating report...'));
    try {
      let path;
      const query = {};
      if (type === 'daily') { path = '/api/v1/reports/daily'; query.date = date.value; }
      else if (type === 'weekly') { path = '/api/v1/reports/weekly'; query.date = date.value; }
      else { if (!project.value) { clear(out).append(el('div',{class:'card'},emptyState({title:'Select a project',description:'Choose a project to generate its report.'}))); return; } path = `/api/v1/reports/projects/${encodeURIComponent(project.value)}`; }
      const report = await api.get(path, { query });
      renderReport(out, report, path, query);
    } catch (err) { clear(out).append(errorState({ description: err.message, onRetry: generate })); }
  }
}

function renderReport(out, report, path, query) {
  clear(out);
  const cards = [
    ['Sessions', report.sessions], ['Events', report.events], ['Commits', report.commits], ['Test runs', report.test_runs],
    ['Failed tests', report.failed_test_runs], ['Investigations', report.investigations], ['Open investigations', report.open_investigations],
    ['Decisions', report.decisions], ['Notes', report.notes]
  ];
  const exportMd = el('button', { class:'btn btn-sm', type:'button', text:'Export Markdown' });
  const exportJson = el('button', { class:'btn btn-sm', type:'button', text:'Export JSON' });
  exportMd.onclick = async () => { try { const md = await api.get(path, { query: { ...query, format: 'md' } }); download(String(md), `devtrail-${report.report_type}-report.md`, 'text/markdown;charset=utf-8'); } catch (e) { alert(e.message); } };
  exportJson.onclick = () => download(JSON.stringify(report, null, 2), `devtrail-${report.report_type}-report.json`, 'application/json');
  out.append(
    el('div', { class:'section' },
      el('div', { class:'section-header' }, el('div', { class:'section-header-left' }, el('h2', { class:'section-title', text:`${safeText(report.report_type,'Report').toUpperCase()} report` }), el('p', { class:'section-subtitle', text:`${fmtDate(report.period_start)} — ${fmtDate(report.period_end)}` })), el('div', { class:'section-actions' }, exportMd, exportJson)),
      el('div', { class:'grid grid-3' }, ...cards.map(([label, value]) => el('div', { class:'card stat-card' }, el('div', { class:'stat-label', text:label }), el('div', { class:'stat-value', style:'font-size:1.5rem;', text:String(value ?? 0) }))))
    ),
    el('div', { class:'card' }, el('div', { class:'card-body' }, el('div', { class:'detail-label', text:'Report data' }), el('pre', { class:'prose', style:'margin-top:0.5rem;font-size:12px;' , text: JSON.stringify(report, null, 2) })))
  );
}

function download(content, filename, type) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a'); a.href = url; a.download = filename; a.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
