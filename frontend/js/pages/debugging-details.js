import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { openModal, closeModal } from '../components/modal.js';
import { toastSuccess } from '../components/toast.js';
import { el, clear, fmtDate, safeText, statusBadge, getQuery } from '../utils.js';
import { icon } from '../components/icons.js';

await requireAuth();
const id = getQuery().id;
const shell = mountShell({ page: 'debugging', title: 'Investigation', description: 'Loading investigation...', breadcrumbs: [{label:'Debugging Journal',href:'/pages/debugging.html'},{label:'Investigation'}] });
const content = shell.content;
if (!id) { clear(content).append(errorState({ title:'Missing investigation', description:'No investigation ID was provided.' })); } else load();

async function load() {
  clear(content).append(loadingState('Loading investigation...'));
  try {
    const inv = await api.get(`/api/v1/investigations/${id}`);
    let project = null; try { project = await api.get(`/api/v1/projects/${inv.project_id}`); } catch {}
    render(inv, project);
  } catch (err) { clear(content).append(errorState({ description: err.message, onRetry: load })); }
}

function render(inv, project) {
  const st = statusBadge(inv.status);
  shell.pageHeader.querySelector('.page-title').textContent = inv.title;
  shell.pageHeader.querySelector('.page-description').textContent = project ? `Project: ${project.name}` : 'Debugging investigation';
  const bc = shell.topbarRight.closest('.topbar').querySelector('.breadcrumbs');
  bc.innerHTML = ''; bc.append(el('a',{href:'/pages/debugging.html',text:'Debugging Journal'}), el('span',{class:'breadcrumb-sep',text:'/'}), el('span',{class:'current',text:inv.title}));
  const editBtn = el('button', { class:'btn btn-primary', type:'button' }); editBtn.innerHTML = icon('plus').replace('<path d="M12 5v14M5 12h14"/>','<path d="M11 4h2v16h-2z"/><path d="M4 11h16v2H4z"/>') + '<span>Edit</span>';
  editBtn.innerHTML = '<span>Edit investigation</span>'; editBtn.onclick = () => openEdit(inv);
  shell.setActions(editBtn);
  clear(content);
  content.append(
    el('div', { class:'card section' }, el('div', { class:'card-body detail-grid' },
      detail('Status', el('span', { class:`badge ${st.cls}`, text: st.label })),
      detail('Project', project ? el('a',{class:'link',href:`/pages/project-details.html?id=${encodeURIComponent(project.id)}`,text:project.name}) : inv.project_id),
      detail('Severity', safeText(inv.severity)),
      detail('Created', fmtDate(inv.created_at)),
      detail('Updated', fmtDate(inv.updated_at)),
      detail('Resolved', inv.resolved_at ? fmtDate(inv.resolved_at) : '—'),
    )),
    section('Problem', inv.problem_description),
    section('Observed behaviour', inv.observed_behavior),
    section('Investigation steps', inv.investigation_steps),
    section('Approaches attempted', inv.attempted_approaches),
    section('Root cause', inv.root_cause),
    section('Resolution', inv.resolution),
    section('Lessons learned', inv.lessons_learned),
    section('Prevention strategy', inv.prevention_strategies),
  );
}

function section(title, value) {
  return el('section', { class:'section' }, el('div', { class:'section-header' }, el('div', { class:'section-header-left' }, el('h2', { class:'section-title', text:title }))), el('div', { class:'card' }, el('div', { class:'card-body' }, el('div', { class:'prose', text: safeText(value, 'Not recorded yet.') }))));
}
function detail(label, value, mono=false) { const v = value?.nodeType ? value : el('div', { class:`detail-value ${mono?'mono':''}`, text:safeText(value) }); return el('div', { class:'detail-field' }, el('div', { class:'detail-label', text:label }), v); }

function openEdit(inv) {
  const fields = [
    ['title','Title','input',inv.title], ['status','Status','select',inv.status], ['severity','Severity','input',inv.severity||''],
    ['problem_description','Problem description','textarea',inv.problem_description], ['observed_behavior','Observed behaviour','textarea',inv.observed_behavior||''],
    ['investigation_steps','Investigation steps','textarea',inv.investigation_steps||''], ['attempted_approaches','Approaches attempted','textarea',inv.attempted_approaches||''],
    ['root_cause','Root cause','textarea',inv.root_cause||''], ['resolution','Resolution','textarea',inv.resolution||''],
    ['lessons_learned','Lessons learned','textarea',inv.lessons_learned||''], ['prevention_strategies','Prevention strategy','textarea',inv.prevention_strategies||''],
  ];
  const errBox = el('div', { class:'form-error' });
  const form = el('form', { class:'form' }, errBox);
  const refs = {};
  for (const [k,label,type,val] of fields) {
    let input;
    if (type==='textarea') { input = el('textarea', { class:'textarea', id:`e-${k}` }); input.value = val; }
    else if (type==='select') { input = el('select', { class:'select', id:`e-${k}` }, ...['open','resolved','abandoned'].map(s=>el('option',{value:s,text:s,selected:s===val}))); }
    else { input = el('input', { class:'input', id:`e-${k}`, value: val }); }
    refs[k]=input;
    form.append(el('div', { class:'field' }, el('label',{for:`e-${k}`,text:label}), input));
  }
  const cancel = el('button', { class:'btn', type:'button', text:'Cancel' }); cancel.onclick = closeModal;
  const submit = el('button', { class:'btn btn-primary', type:'submit', text:'Save changes' });
  form.onsubmit = async (e) => {
    e.preventDefault(); errBox.textContent='';
    submit.disabled=true; submit.textContent='Saving...';
    const body = {}; for (const [k] of fields) body[k] = refs[k].value;
    try { await api.patch(`/api/v1/investigations/${id}`, body); closeModal(); toastSuccess('Investigation updated.'); await load(); }
    catch (err) { errBox.textContent = err.message; submit.disabled=false; submit.textContent='Save changes'; }
  };
  openModal({ title:'Edit investigation', description:'Update investigation details and status.', content: form, actions:[cancel,submit], wide:true });
  submit.onclick = () => form.requestSubmit();
}
