import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { openModal, closeModal } from '../components/modal.js';
import { toastSuccess } from '../components/toast.js';
import { el, clear, fmtRelative, safeText, statusBadge } from '../utils.js';
import { icon } from '../components/icons.js';

await requireAuth();
const { content, setActions } = mountShell({ page: 'debugging', title: 'Debugging Journal', description: 'Document investigations, track what you tried, and preserve root causes and lessons learned.' });
const btn = el('button', { class:'btn btn-primary', type:'button' }); btn.innerHTML = icon('plus') + '<span>New investigation</span>'; btn.onclick = () => openForm(); setActions(btn);
let projects = [], items = []; const filters = { project_id: '', status: '' };
load();

async function load() {
  clear(content).append(loadingState('Loading investigations...'));
  try {
    const [p, i] = await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/investigations', { query: { project_id: filters.project_id } })]);
    projects = p.projects || []; items = i.investigations || []; render();
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render() {
  clear(content);
  const pSel = el('select', { class:'select' }, el('option', { value:'', text:'All projects' }), ...projects.map(p => el('option', { value:p.id, text:p.name, selected: filters.project_id===p.id })));
  const sSel = el('select', { class:'select' }, el('option', { value:'', text:'All statuses' }), ...['open','resolved','abandoned'].map(s => el('option', { value:s, text:s, selected: filters.status===s })));
  const apply = el('button', { class:'btn btn-sm', type:'button', text:'Apply' }); apply.onclick = () => { filters.project_id = pSel.value; filters.status = sSel.value; load(); };
  content.append(el('div', { class:'filters' }, filter('Project',pSel), filter('Status',sSel), el('div',{class:'filter'},el('label',{text:'\u00a0'}),apply)));
  const visible = filters.status ? items.filter(i => i.status === filters.status) : items;
  if (!visible.length) { content.append(el('div', { class:'card' }, emptyState({ title:'No investigations yet', description:'Document your first investigation to build engineering knowledge from debugging work.', action: (()=>{const b=el('button',{class:'btn btn-primary',type:'button',text:'New investigation'}); b.onclick=()=>openForm(); return b;})() }))); return; }
  content.append(el('div', { class:'card card-table' }, el('div', { class:'table-wrap' }, el('table', { class:'table' },
    el('thead', {}, el('tr', {}, ...['Investigation','Project','Status','Updated',''].map(h=>el('th',{text:h})))),
    el('tbody', {}, ...visible.map(i => {
      const st = statusBadge(i.status); const project = projects.find(p => p.id === i.project_id);
      return el('tr', {},
        el('td', {}, el('a', { class:'table-link', href:`/pages/debugging-details.html?id=${encodeURIComponent(i.id)}`, text:safeText(i.title) }), el('div', { class:'cell-secondary', text:safeText(i.problem_description).slice(0,120) })),
        el('td', { text: project?.name || i.project_id.slice(0,8) }),
        el('td', {}, el('span', { class:`badge ${st.cls}`, text:st.label })),
        el('td', { text: fmtRelative(i.updated_at || i.created_at) }),
        el('td', { class:'cell-actions' }, el('a', { class:'btn btn-ghost btn-sm', href:`/pages/debugging-details.html?id=${encodeURIComponent(i.id)}`, text:'Open' }))
      );
    }))
  ))));
}

function openForm() {
  const project = el('select', { class:'select', id:'i-project' }, el('option',{value:'',text:'Select project'}), ...projects.map(p=>el('option',{value:p.id,text:p.name})));
  const title = el('input', { class:'input', id:'i-title' });
  const problem = el('textarea', { class:'textarea', id:'i-problem' });
  const observed = el('textarea', { class:'textarea', id:'i-observed' });
  const errorSummary = el('input', { class:'input', id:'i-severity', placeholder:'e.g. high, medium, low' });
  const notes = el('textarea', { class:'textarea', id:'i-notes' });
  const errBox = el('div', { class:'form-error' });
  const form = el('form', { class:'form' }, errBox,
    field('i-project','Project',project), field('i-title','Title',title), field('i-problem','Problem description',problem),
    field('i-observed','Observed behaviour',observed), field('i-severity','Severity (optional)',errorSummary), field('i-notes','Initial investigation notes',notes)
  );
  const cancel = el('button', { class:'btn', type:'button', text:'Cancel' }); cancel.onclick = closeModal;
  const submit = el('button', { class:'btn btn-primary', type:'submit', text:'Create investigation' });
  form.onsubmit = async (e) => {
    e.preventDefault(); errBox.textContent='';
    if (!project.value || !title.value.trim() || !problem.value.trim()) { errBox.textContent='Project, title, and problem description are required.'; return; }
    submit.disabled=true; submit.textContent='Creating...';
    try {
      await api.post('/api/v1/investigations', { project_id: project.value, title: title.value.trim(), problem_description: problem.value.trim(), observed_behavior: observed.value.trim(), severity: errorSummary.value.trim(), investigation_steps: notes.value.trim() });
      closeModal(); toastSuccess('Investigation created.'); await load();
    } catch (err) { errBox.textContent = err.message; submit.disabled=false; submit.textContent='Create investigation'; }
  };
  openModal({ title:'New investigation', description:'Capture the problem now. You can add root cause and resolution as you learn more.', content: form, actions:[cancel,submit], wide:true });
  submit.onclick = () => form.requestSubmit();
}
function filter(label, input) { return el('div', { class:'filter' }, el('label',{text:label}), input); }
function field(id, label, input) { return el('div', { class:'field' }, el('label',{for:id,text:label}), input); }
