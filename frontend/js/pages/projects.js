import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { openModal, closeModal, confirmModal } from '../components/modal.js';
import { toastSuccess, toastError } from '../components/toast.js';
import { el, clear, fmtRelative, safeText, statusBadge } from '../utils.js';
import { icon } from '../components/icons.js';

await requireAuth();
const { content, setActions } = mountShell({
  page: 'projects',
  title: 'Projects',
  description: 'Register and manage the software projects whose engineering history you want DevTrail to track.',
});

const createBtn = el('button', { class: 'btn btn-primary', type: 'button' });
createBtn.innerHTML = icon('plus') + '<span>New project</span>';
createBtn.addEventListener('click', () => openProjectForm());
setActions(createBtn);

let projects = [];
load();

async function load() {
  clear(content).append(loadingState('Loading projects...'));
  try {
    const res = await api.get('/api/v1/projects');
    projects = res.projects || [];
    render();
  } catch (err) {
    clear(content).append(errorState({ description: err.message, onRetry: load }));
  }
}

function render() {
  clear(content);
  if (!projects.length) {
    content.append(el('div', { class: 'card' }, emptyState({
      title: 'No projects yet',
      description: 'Register a project to start tracking sessions, activity, decisions, and debugging history.',
      action: (() => { const b = el('button', { class: 'btn btn-primary', type: 'button', text: 'Create project' }); b.addEventListener('click', () => openProjectForm()); return b; })()
    })));
    return;
  }
  const table = el('table', { class: 'table' },
    el('thead', {}, el('tr', {}, ...['Project', 'Status', 'Branch', 'Last updated', ''].map(h => el('th', { text: h })))),
    el('tbody', {}, ...projects.map(p => {
      const st = statusBadge(p.status);
      const row = el('tr', {},
        el('td', {}, el('a', { class: 'table-link', href: `/pages/project-details.html?id=${encodeURIComponent(p.id)}`, text: safeText(p.name) }), el('div', { class: 'cell-secondary', text: safeText(p.description, 'No description') })),
        el('td', {}, el('span', { class: `badge ${st.cls}`, text: st.label })),
        el('td', {}, el('span', { class: 'cell-mono', text: safeText(p.default_branch, '—') })),
        el('td', { text: fmtRelative(p.updated_at || p.created_at) }),
        el('td', { class: 'cell-actions' })
      );
      const actions = row.lastChild;
      const edit = el('button', { class: 'btn btn-ghost btn-sm', type: 'button', text: 'Edit' });
      edit.addEventListener('click', () => openProjectForm(p));
      const archive = el('button', { class: 'btn btn-ghost btn-sm', type: 'button', text: p.status === 'archived' ? 'Archived' : 'Archive', disabled: p.status === 'archived' });
      archive.addEventListener('click', () => confirmModal({
        title: 'Archive project',
        description: `Archive "${p.name}"? It will be marked as archived but remain in your history.`,
        confirmLabel: 'Archive',
        onConfirm: async () => { await api.del(`/api/v1/projects/${p.id}`); toastSuccess('Project archived.'); await load(); }
      }));
      actions.append(edit, archive);
      return row;
    }))
  );
  content.append(el('div', { class: 'card card-table' }, el('div', { class: 'table-wrap' }, table)));
}

function openProjectForm(project = null) {
  const isEdit = !!project;
  const name = el('input', { class: 'input', id: 'p-name', required: true, value: project?.name || '' });
  const desc = el('textarea', { class: 'textarea', id: 'p-desc' }); desc.value = project?.description || '';
  const branch = el('input', { class: 'input', id: 'p-branch', value: project?.default_branch || '' });
  const repo = el('input', { class: 'input', id: 'p-repo', value: project?.repo_root_hash || '', disabled: isEdit });
  const status = el('select', { class: 'select', id: 'p-status' },
    ...['active','archived'].map(s => el('option', { value: s, text: s.charAt(0).toUpperCase()+s.slice(1), selected: (project?.status || 'active') === s }))
  );
  const formError = el('div', { class: 'form-error' });
  const form = el('form', { class: 'form', id: 'project-form' },
    formError,
    field('p-name', 'Project name', name),
    field('p-desc', 'Description', desc),
    field('p-branch', 'Default branch', branch),
    !isEdit ? field('p-repo', 'Repository root hash (optional)', repo, 'Optional stable identifier; avoid raw machine paths.') : null,
    isEdit ? field('p-status', 'Status', status) : null,
  );
  const cancel = el('button', { class: 'btn', type: 'button', text: 'Cancel' });
  const submit = el('button', { class: 'btn btn-primary', type: 'submit', text: isEdit ? 'Save changes' : 'Create project' });
  cancel.addEventListener('click', closeModal);
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    formError.textContent = '';
    if (!name.value.trim()) { formError.textContent = 'Project name is required.'; return; }
    submit.disabled = true; submit.textContent = 'Saving...';
    try {
      if (isEdit) {
        await api.patch(`/api/v1/projects/${project.id}`, { name: name.value.trim(), description: desc.value.trim(), default_branch: branch.value.trim(), status: status.value });
        toastSuccess('Project updated.');
      } else {
        await api.post('/api/v1/projects', { name: name.value.trim(), description: desc.value.trim(), default_branch: branch.value.trim(), repo_root_hash: repo.value.trim() });
        toastSuccess('Project created.');
      }
      closeModal();
      await load();
    } catch (err) {
      formError.textContent = err.message || 'Failed to save project.';
      submit.disabled = false; submit.textContent = isEdit ? 'Save changes' : 'Create project';
    }
  });
  openModal({ title: isEdit ? 'Edit project' : 'New project', description: isEdit ? 'Update project details.' : 'Register a project to track its engineering history.', content: form, actions: [cancel, submit] });
  submit.onclick = () => form.requestSubmit();
}

function field(id, label, input, hint) {
  return el('div', { class: 'field' }, el('label', { for: id, text: label }), input, hint ? el('div', { class: 'field-hint', text: hint }) : null);
}
