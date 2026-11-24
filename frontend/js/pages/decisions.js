import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { openModal, closeModal } from '../components/modal.js';
import { toastSuccess } from '../components/toast.js';
import { el, clear, fmtRelative, safeText, statusBadge } from '../utils.js';
import { icon } from '../components/icons.js';

await requireAuth();
const { content, setActions } = mountShell({ page: 'decisions', title: 'Engineering Decisions', description: 'Record architectural and technical decisions, alternatives, and consequences.' });
const btn = el('button', { class:'btn btn-primary', type:'button' }); btn.innerHTML = icon('plus')+'<span>New decision</span>'; btn.onclick=openForm; setActions(btn);
let projects=[], items=[]; const filters={project_id:'', status:''}; load();

async function load(){ clear(content).append(loadingState('Loading decisions...')); try{ const [p,d]=await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/decisions',{query:{project_id:filters.project_id}})]); projects=p.projects||[]; items=d.decisions||[]; render(); }catch(err){ clear(content).append(errorState({description:err.message,onRetry:load})); }}
function render(){ clear(content);
  const pSel=el('select',{class:'select'},el('option',{value:'',text:'All projects'}),...projects.map(p=>el('option',{value:p.id,text:p.name,selected:filters.project_id===p.id})));
  const sSel=el('select',{class:'select'},el('option',{value:'',text:'All statuses'}),...['proposed','accepted','superseded','deprecated'].map(s=>el('option',{value:s,text:s,selected:filters.status===s})));
  const apply=el('button',{class:'btn btn-sm',type:'button',text:'Apply'}); apply.onclick=()=>{filters.project_id=pSel.value;filters.status=sSel.value;load();};
  content.append(el('div',{class:'filters'},f('Project',pSel),f('Status',sSel),el('div',{class:'filter'},el('label',{text:'\u00a0'}),apply)));
  const visible=filters.status?items.filter(i=>i.status===filters.status):items;
  if(!visible.length){ content.append(el('div',{class:'card'},emptyState({title:'No decisions yet',description:'Record engineering decisions to preserve context and reasoning.',action:(()=>{const b=el('button',{class:'btn btn-primary',type:'button',text:'New decision'});b.onclick=openForm;return b;})()}))); return; }
  content.append(el('div',{class:'card card-table'},el('div',{class:'table-wrap'},el('table',{class:'table'},
    el('thead',{},el('tr',{},...['Decision','Project','Status','Decided',''].map(h=>el('th',{text:h})))),
    el('tbody',{},...visible.map(d=>{ const st=statusBadge(d.status); const project=projects.find(p=>p.id===d.project_id); return el('tr',{},
      el('td',{},el('a',{class:'table-link',href:`/pages/decision-details.html?id=${encodeURIComponent(d.id)}`,text:safeText(d.title)}),el('div',{class:'cell-secondary',text:safeText(d.problem).slice(0,120)})),
      el('td',{text:project?.name||d.project_id.slice(0,8)}), el('td',{},el('span',{class:`badge ${st.cls}`,text:st.label})), el('td',{text:fmtRelative(d.decided_at||d.created_at)}),
      el('td',{class:'cell-actions'},el('a',{class:'btn btn-ghost btn-sm',href:`/pages/decision-details.html?id=${encodeURIComponent(d.id)}`,text:'Open'}))
    ); }))
  ))));
}
function openForm(){
  const project=el('select',{class:'select',id:'d-project'},el('option',{value:'',text:'Select project'}),...projects.map(p=>el('option',{value:p.id,text:p.name})));
  const fields={title:inp('d-title'),context:ta('d-context'),problem:ta('d-problem'),alternatives:ta('d-alternatives'),selected_approach:ta('d-selected'),reasoning:ta('d-reasoning'),consequences:ta('d-consequences')};
  const status=el('select',{class:'select',id:'d-status'},...['proposed','accepted','superseded','deprecated'].map(s=>el('option',{value:s,text:s,selected:s==='accepted'})));
  const errBox=el('div',{class:'form-error'});
  const form=el('form',{class:'form'},errBox, fld('d-project','Project',project), fld('d-title','Title',fields.title), fld('d-context','Context',fields.context), fld('d-problem','Problem',fields.problem), fld('d-alternatives','Alternatives considered',fields.alternatives), fld('d-selected','Selected approach',fields.selected_approach), fld('d-reasoning','Reasoning',fields.reasoning), fld('d-consequences','Consequences',fields.consequences), fld('d-status','Status',status));
  const cancel=el('button',{class:'btn',type:'button',text:'Cancel'}); cancel.onclick=closeModal;
  const submit=el('button',{class:'btn btn-primary',type:'submit',text:'Create decision'});
  form.onsubmit=async(e)=>{ e.preventDefault(); errBox.textContent=''; if(!project.value||!fields.title.value.trim()||!fields.context.value.trim()||!fields.problem.value.trim()||!fields.selected_approach.value.trim()||!fields.reasoning.value.trim()){ errBox.textContent='Project, title, context, problem, selected approach, and reasoning are required.'; return; } submit.disabled=true; submit.textContent='Creating...'; try{ await api.post('/api/v1/decisions',{project_id:project.value,title:fields.title.value.trim(),context:fields.context.value.trim(),problem:fields.problem.value.trim(),alternatives:fields.alternatives.value.trim(),selected_approach:fields.selected_approach.value.trim(),reasoning:fields.reasoning.value.trim(),consequences:fields.consequences.value.trim(),status:status.value}); closeModal(); toastSuccess('Decision created.'); await load(); }catch(err){ errBox.textContent=err.message; submit.disabled=false; submit.textContent='Create decision'; }};
  openModal({title:'New decision',description:'Record the context, alternatives, and reasoning behind an engineering decision.',content:form,actions:[cancel,submit],wide:true}); submit.onclick=()=>form.requestSubmit();
}
function inp(id){return el('input',{class:'input',id});} function ta(id){const t=el('textarea',{class:'textarea',id}); return t;} function fld(id,l,i){return el('div',{class:'field'},el('label',{for:id,text:l}),i);} function f(l,i){return el('div',{class:'filter'},el('label',{text:l}),i);}
