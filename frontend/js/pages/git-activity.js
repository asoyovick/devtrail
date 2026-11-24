import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, fmtRelative, fmtDate, safeText } from '../utils.js';

await requireAuth();
const { content } = mountShell({ page: 'git', title: 'Git Activity', description: 'Review commit history captured across your registered projects.' });
let projects = [], commits = []; const filters = { project_id: '' };
load();

async function load() {
  clear(content).append(loadingState('Loading git activity...'));
  try {
    const [p, c] = await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/git/commits', { query: { project_id: filters.project_id } })]);
    projects = p.projects || []; commits = c.commits || []; render();
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render() {
  clear(content);
  const sel = el('select', { class: 'select' }, el('option', { value:'', text:'All projects' }), ...projects.map(p => el('option', { value:p.id, text:p.name, selected: filters.project_id===p.id })));
  const apply = el('button', { class:'btn btn-sm', type:'button', text:'Apply' });
  apply.onclick = () => { filters.project_id = sel.value; load(); };
  content.append(el('div', { class:'filters' }, el('div', { class:'filter' }, el('label', { text:'Project' }), sel), el('div', { class:'filter' }, el('label', { text:'\u00a0' }), apply)));
  if (!commits.length) { content.append(el('div', { class:'card' }, emptyState({ title:'No git activity yet', description:'Commits recorded by DevTrail will appear here.' }))); return; }
  content.append(el('div', { class:'card card-table' }, el('div', { class:'table-wrap' }, el('table', { class:'table' },
    el('thead', {}, el('tr', {}, ...['Commit','Branch','Project','Author','Committed'].map(h=>el('th',{text:h})))),
    el('tbody', {}, ...commits.map(c => {
      const project = projects.find(p => p.id === c.project_id);
      return el('tr', {},
        el('td', {}, el('div', { class:'cell-primary', text: safeText(c.message_summary, c.commit_hash) }), el('div', { class:'cell-secondary cell-mono', text: String(c.commit_hash||'').slice(0,12) })),
        el('td', {}, c.branch ? el('span', { class:'badge badge-accent', text:c.branch }) : el('span', { class:'muted', text:'—' })),
        el('td', {}, project ? el('a', { class:'link', href:`/pages/project-details.html?id=${encodeURIComponent(project.id)}`, text:project.name }) : el('span', { class:'cell-mono', text:String(c.project_id||'').slice(0,8) })),
        el('td', { text: safeText(c.author_name, '—') }),
        el('td', {}, el('div', { text: fmtRelative(c.committed_at) }), el('div', { class:'cell-secondary', text: fmtDate(c.committed_at) }))
      );
    }))
  ))));
}
