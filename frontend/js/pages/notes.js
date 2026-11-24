import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { openModal, closeModal, confirmModal } from '../components/modal.js';
import { toastSuccess } from '../components/toast.js';
import { el, clear, fmtRelative, safeText } from '../utils.js';
import { icon } from '../components/icons.js';

await requireAuth();
const { content, setActions } = mountShell({ page: 'notes', title: 'Technical Notes', description: 'Capture technical knowledge, observations, and project context in a lightweight notes workspace.' });
const btn = el('button',{class:'btn btn-primary',type:'button'}); btn.innerHTML=icon('plus')+'<span>New note</span>'; btn.onclick=()=>openForm(); setActions(btn);
let projects=[], notes=[]; const filters={project_id:'', q:''}; load();

async function load(){ clear(content).append(loadingState('Loading notes...')); try{ const [p,n]=await Promise.all([api.get('/api/v1/projects'), api.get('/api/v1/notes',{query:{project_id:filters.project_id}})]); projects=p.projects||[]; notes=n.notes||[]; render(); }catch(err){ clear(content).append(errorState({description:err.message,onRetry:load})); }}
function render(){ clear(content);
  const pSel=el('select',{class:'select'},el('option',{value:'',text:'All projects'}),...projects.map(p=>el('option',{value:p.id,text:p.name,selected:filters.project_id===p.id})));
  const q=el('input',{class:'input search-input',type:'search',placeholder:'Search loaded notes...',value:filters.q});
  const apply=el('button',{class:'btn btn-sm',type:'button',text:'Apply'}); apply.onclick=()=>{filters.project_id=pSel.value;filters.q=q.value.trim().toLowerCase();load();};
  q.addEventListener('keydown',(e)=>{ if(e.key==='Enter'){ e.preventDefault(); apply.click(); }});
  content.append(el('div',{class:'filters'},el('div',{class:'filter'},el('label',{text:'Project'}),pSel),el('div',{class:'filter filter-grow'},el('label',{text:'Search'}),q),el('div',{class:'filter'},el('label',{text:'\u00a0'}),apply)));
  const visible = filters.q ? notes.filter(n => (n.title+' '+n.body+' '+(n.tags||[]).join(' ')).toLowerCase().includes(filters.q)) : notes;
  if(!visible.length){ content.append(el('div',{class:'card'},emptyState({title:'No notes found',description:'Create a note to capture technical knowledge and context.',action:(()=>{const b=el('button',{class:'btn btn-primary',type:'button',text:'New note'});b.onclick=()=>openForm();return b;})()}))); return; }
  content.append(el('div',{class:'grid grid-2'},...visible.map(n => {
    const project=projects.find(p=>p.id===n.project_id);
    const card=el('div',{class:'card'},el('div',{class:'card-body'},
      el('div',{style:'display:flex;justify-content:space-between;gap:1rem;align-items:flex-start;'},el('div',{},el('div',{class:'card-title',text:safeText(n.title)}),el('div',{class:'card-subtitle',text:project?project.name:(n.project_id?'Project '+n.project_id.slice(0,8):'Global note')})),el('div',{style:'display:flex;gap:0.25rem;'},editBtn(n),delBtn(n))),
      el('div',{class:'prose',style:'margin-top:0.75rem;',text:safeText(n.body).slice(0,300)}),
      el('div',{class:'list-meta',style:'margin-top:0.75rem;'},...(n.tags||[]).map(t=>el('span',{class:'badge badge-neutral',text:t})),el('span',{text:fmtRelative(n.updated_at||n.created_at)}))
    ));
    return card;
  })));
}
function editBtn(n){ const b=el('button',{class:'btn btn-ghost btn-sm',type:'button',text:'Edit'}); b.onclick=()=>openForm(n); return b; }
function delBtn(n){ const b=el('button',{class:'btn btn-ghost btn-sm',type:'button',text:'Delete'}); b.onclick=()=>confirmModal({title:'Delete note',description:`Delete "${n.title}"? This action cannot be undone.`,confirmLabel:'Delete',danger:true,onConfirm:async()=>{ await api.del(`/api/v1/notes/${n.id}`); toastSuccess('Note deleted.'); await load(); }}); return b; }
function openForm(note=null){ const isEdit=!!note; const title=el('input',{class:'input',id:'n-title',value:note?.title||''}); const body=el('textarea',{class:'textarea',id:'n-body',style:'min-height:180px;'}); body.value=note?.body||''; const tags=el('input',{class:'input',id:'n-tags',value:(note?.tags||[]).join(', '),placeholder:'comma, separated, tags'}); const project=el('select',{class:'select',id:'n-project'},el('option',{value:'',text:'Global note (no project)'}),...projects.map(p=>el('option',{value:p.id,text:p.name,selected:note?.project_id===p.id}))); const errBox=el('div',{class:'form-error'}); const form=el('form',{class:'form'},errBox,field('n-title','Title',title),field('n-project','Project',project),field('n-body','Body',body),field('n-tags','Tags',tags,'Comma-separated tags')); const cancel=el('button',{class:'btn',type:'button',text:'Cancel'}); cancel.onclick=closeModal; const submit=el('button',{class:'btn btn-primary',type:'submit',text:isEdit?'Save changes':'Create note'});
  form.onsubmit=async(e)=>{ e.preventDefault(); errBox.textContent=''; if(!title.value.trim()||!body.value.trim()){ errBox.textContent='Title and body are required.'; return; } submit.disabled=true; submit.textContent='Saving...'; const payload={title:title.value.trim(),body:body.value.trim(),tags:tags.value.split(',').map(s=>s.trim()).filter(Boolean)}; if(!isEdit) payload.project_id=project.value||null; try{ if(isEdit) await api.patch(`/api/v1/notes/${note.id}`,{title:payload.title,body:payload.body,tags:payload.tags}); else await api.post('/api/v1/notes',payload); closeModal(); toastSuccess(isEdit?'Note updated.':'Note created.'); await load(); }catch(err){ errBox.textContent=err.message; submit.disabled=false; submit.textContent=isEdit?'Save changes':'Create note'; } };
  openModal({title:isEdit?'Edit note':'New note',description:isEdit?'Update your technical note.':'Capture a technical note for your workspace.',content:form,actions:[cancel,submit],wide:true}); submit.onclick=()=>form.requestSubmit();
}
function field(id,label,input,hint){ return el('div',{class:'field'},el('label',{for:id,text:label}),input,hint?el('div',{class:'field-hint',text:hint}):null); }
