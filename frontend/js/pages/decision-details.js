import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, errorState } from '../components/states.js';
import { openModal, closeModal } from '../components/modal.js';
import { toastSuccess } from '../components/toast.js';
import { el, clear, fmtDate, safeText, statusBadge, getQuery } from '../utils.js';

await requireAuth();
const id = getQuery().id;
const shell = mountShell({ page:'decisions', title:'Decision', description:'Loading...', breadcrumbs:[{label:'Decisions',href:'/pages/decisions.html'},{label:'Decision'}] });
const content = shell.content;
if(!id){ clear(content).append(errorState({title:'Missing decision',description:'No decision ID was provided.'})); } else load();

async function load(){ clear(content).append(loadingState('Loading decision...')); try{ const d = await api.get(`/api/v1/decisions/${id}`); let project=null; try{ project=await api.get(`/api/v1/projects/${d.project_id}`);}catch{} render(d,project);}catch(err){ clear(content).append(errorState({description:err.message,onRetry:load})); }}
function render(d,project){ const st=statusBadge(d.status); shell.pageHeader.querySelector('.page-title').textContent=d.title; shell.pageHeader.querySelector('.page-description').textContent=project?`Project: ${project.name}`:'Engineering decision record'; const bc=shell.topbarRight.closest('.topbar').querySelector('.breadcrumbs'); bc.innerHTML=''; bc.append(el('a',{href:'/pages/decisions.html',text:'Decisions'}),el('span',{class:'breadcrumb-sep',text:'/'}),el('span',{class:'current',text:d.title})); const edit=el('button',{class:'btn btn-primary',type:'button',text:'Edit decision'}); edit.onclick=()=>openEdit(d); shell.setActions(edit); clear(content); content.append(
  el('div',{class:'card section'},el('div',{class:'card-body detail-grid'}, detail('Status',el('span',{class:`badge ${st.cls}`,text:st.label})), detail('Project',project?el('a',{class:'link',href:`/pages/project-details.html?id=${encodeURIComponent(project.id)}`,text:project.name}):d.project_id), detail('Decided',fmtDate(d.decided_at)), detail('Updated',fmtDate(d.updated_at)))),
  sec('Context',d.context), sec('Problem',d.problem), sec('Alternatives considered',d.alternatives), sec('Selected approach',d.selected_approach), sec('Reasoning',d.reasoning), sec('Consequences',d.consequences)
);}
function sec(title,value){ return el('section',{class:'section'},el('div',{class:'section-header'},el('div',{class:'section-header-left'},el('h2',{class:'section-title',text:title}))),el('div',{class:'card'},el('div',{class:'card-body'},el('div',{class:'prose',text:safeText(value,'Not recorded yet.')})))); }
function detail(l,v){ const val=v?.nodeType?v:el('div',{class:'detail-value',text:safeText(v)}); return el('div',{class:'detail-field'},el('div',{class:'detail-label',text:l}),val); }
function openEdit(d){ const errBox=el('div',{class:'form-error'}); const form=el('form',{class:'form'},errBox); const refs={}; const flds=[['title','Title','input',d.title],['status','Status','select',d.status],['context','Context','textarea',d.context],['problem','Problem','textarea',d.problem],['alternatives','Alternatives','textarea',d.alternatives||''],['selected_approach','Selected approach','textarea',d.selected_approach],['reasoning','Reasoning','textarea',d.reasoning],['consequences','Consequences','textarea',d.consequences||'']]; for(const [k,l,t,v] of flds){ let i; if(t==='textarea'){i=el('textarea',{class:'textarea',id:`e-${k}`}); i.value=v;} else if(t==='select'){i=el('select',{class:'select',id:`e-${k}`},...['proposed','accepted','superseded','deprecated'].map(s=>el('option',{value:s,text:s,selected:s===v})));} else {i=el('input',{class:'input',id:`e-${k}`,value:v});} refs[k]=i; form.append(el('div',{class:'field'},el('label',{for:`e-${k}`,text:l}),i)); }
  const cancel=el('button',{class:'btn',type:'button',text:'Cancel'}); cancel.onclick=closeModal; const submit=el('button',{class:'btn btn-primary',type:'submit',text:'Save changes'});
  form.onsubmit=async(e)=>{ e.preventDefault(); errBox.textContent=''; submit.disabled=true; submit.textContent='Saving...'; const body={}; for(const [k] of flds) body[k]=refs[k].value; try{ await api.patch(`/api/v1/decisions/${id}`,body); closeModal(); toastSuccess('Decision updated.'); await load(); }catch(err){ errBox.textContent=err.message; submit.disabled=false; submit.textContent='Save changes'; }};
  openModal({title:'Edit decision',description:'Update decision details and status.',content:form,actions:[cancel,submit],wide:true}); submit.onclick=()=>form.requestSubmit();
}
