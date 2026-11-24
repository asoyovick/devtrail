import { el, qs } from '../utils.js';
import { register, redirectIfAuthenticated } from '../auth.js';

if (!redirectIfAuthenticated()) render();

function render() {
  const app = qs('#app');
  const formError = el('div', { class: 'form-error', id: 'form-error', role: 'alert' });
  const name = el('input', { class: 'input', id: 'name', type: 'text', autocomplete: 'name', required: true });
  const email = el('input', { class: 'input', id: 'email', type: 'email', autocomplete: 'email', required: true });
  const password = el('input', { class: 'input', id: 'password', type: 'password', autocomplete: 'new-password', required: true });
  const confirm = el('input', { class: 'input', id: 'confirm', type: 'password', autocomplete: 'new-password', required: true });
  const submit = el('button', { class: 'btn btn-primary btn-block', type: 'submit', text: 'Create account' });

  const field = (id, label, input, hint) => el('div', { class: 'field' },
    el('label', { for: id, text: label }), input,
    hint ? el('div', { class: 'field-hint', text: hint }) : null,
    el('div', { class: 'field-error', id: `${id}-error` })
  );

  const form = el('form', { class: 'form', novalidate: true },
    formError,
    field('name', 'Full name', name),
    field('email', 'Email address', email),
    field('password', 'Password', password, 'Use at least 8 characters.'),
    field('confirm', 'Confirm password', confirm),
    submit
  );

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    ['name','email','password','confirm','form'].forEach(id => { const n = qs(`#${id}-error`) || formError; n.textContent = ''; });
    formError.textContent = '';
    const v = { name: name.value.trim(), email: email.value.trim(), password: password.value, confirm: confirm.value };
    let ok = true;
    if (!v.name) { qs('#name-error').textContent = 'Full name is required.'; ok = false; }
    if (!v.email) { qs('#email-error').textContent = 'Email is required.'; ok = false; }
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.email)) { qs('#email-error').textContent = 'Enter a valid email address.'; ok = false; }
    if (!v.password) { qs('#password-error').textContent = 'Password is required.'; ok = false; }
    else if (v.password.length < 8) { qs('#password-error').textContent = 'Password must be at least 8 characters.'; ok = false; }
    if (!v.confirm) { qs('#confirm-error').textContent = 'Confirm your password.'; ok = false; }
    else if (v.password !== v.confirm) { qs('#confirm-error').textContent = 'Passwords do not match.'; ok = false; }
    if (!ok) return;
    submit.disabled = true; submit.textContent = 'Creating account...';
    try {
      await register(v.name, v.email, v.password);
      location.href = '/pages/overview.html';
    } catch (err) {
      formError.textContent = err.message || 'Unable to create account.';
      submit.disabled = false; submit.textContent = 'Create account';
    }
  });

  app.append(el('div', { class: 'auth-shell' },
    el('main', { class: 'auth-card', id: 'main-content' },
      el('div', { class: 'auth-brand' }, el('div', { class: 'brand-mark', text: 'D' }), el('div', { class: 'brand-name', text: 'DevTrail' })),
      el('h1', { class: 'auth-title', text: 'Create your account' }),
      el('p', { class: 'auth-desc', text: 'Start building your engineering intelligence workspace.' }),
      form,
      el('p', { class: 'auth-footer' }, 'Already have an account? ', el('a', { href: '/pages/login.html', text: 'Sign in' }))
    )
  ));
}
