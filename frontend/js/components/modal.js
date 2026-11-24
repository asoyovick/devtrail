import { el, qs } from '../utils.js';
import { icon } from './icons.js';

let activeModal = null;

export function openModal({ title, description, content, actions = [], wide = false }) {
  closeModal();
  const backdrop = el('div', { class: 'modal-backdrop open', role: 'dialog', 'aria-modal': 'true' });
  const modal = el('div', { class: 'modal', style: wide ? 'width:min(760px,100%)' : '' });
  const closeBtn = el('button', { class: 'icon-btn', type: 'button', 'aria-label': 'Close dialog' });
  closeBtn.innerHTML = icon('close');
  closeBtn.addEventListener('click', closeModal);

  modal.append(
    el('div', { class: 'modal-header' },
      el('div', {},
        el('div', { class: 'modal-title', text: title || '' }),
        description ? el('div', { class: 'modal-description', text: description }) : null,
      ),
      closeBtn
    ),
    el('div', { class: 'modal-body' }, content),
    actions.length ? el('div', { class: 'modal-footer' }, ...actions) : null
  );
  backdrop.append(modal);
  backdrop.addEventListener('click', (e) => { if (e.target === backdrop) closeModal(); });
  document.addEventListener('keydown', escHandler);
  document.body.append(backdrop);
  activeModal = backdrop;
  const focusable = modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
  (focusable || closeBtn).focus();
  return { close: closeModal, el: modal };
}

function escHandler(e) { if (e.key === 'Escape') closeModal(); }
export function closeModal() {
  if (activeModal) { activeModal.remove(); activeModal = null; }
  document.removeEventListener('keydown', escHandler);
}

export function confirmModal({ title, description, confirmLabel = 'Confirm', danger = false, onConfirm }) {
  let busy = false;
  const cancel = el('button', { class: 'btn', type: 'button', text: 'Cancel' });
  const confirm = el('button', { class: `btn ${danger ? 'btn-danger' : 'btn-primary'}`, type: 'button', text: confirmLabel });
  cancel.addEventListener('click', closeModal);
  confirm.addEventListener('click', async () => {
    if (busy) return; busy = true; confirm.disabled = true; confirm.textContent = 'Working...';
    try { await onConfirm?.(); closeModal(); }
    catch (err) { confirm.disabled = false; confirm.textContent = confirmLabel; busy = false; throw err; }
  });
  return openModal({ title, description, content: el('div'), actions: [cancel, confirm] });
}
