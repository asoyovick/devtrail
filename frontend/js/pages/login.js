import { el, qs, getQuery } from '../utils.js';
import { login, redirectIfAuthenticated } from '../auth.js';

if (!redirectIfAuthenticated()) render();

function render() {
  const app = qs('#app');
  const formError = el('div', { class: 'form-error', id: 'form-error', role: 'alert' });
  const email = el('input', { class: 'input', id: 'email', type: 'email', autocomplete: 'email', required: true });
  const password = el('input', { class: 'input', id: 'password', type: 'password', autocomplete: 'current-password', required: true });
  const submit = el('button', { class: 'btn btn-primary btn-block', type: 'submit', text: 'Sign in' });

  const form = el('form', { class: 'form', novalidate: true },
    formError,
    el('div', { class: 'field' }, el('label', { for: 'email', text: 'Email' }), email, el('div', { class: 'field-error', id: 'email-error' })),
    el('div', { class: 'field' }, el('label', { for: 'password', text: 'Password' }), password, el('div', { class: 'field-error', id: 'password-error' })),
    submit
  );

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    formError.textContent = '';
    qs('#email-error').textContent = '';
    qs('#password-error').textContent = '';
    const em = email.value.trim();
    const pw = password.value;
    let ok = true;
    if (!em) { qs('#email-error').textContent = 'Email is required.'; ok = false; }
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(em)) { qs('#email-error').textContent = 'Enter a valid email address.'; ok = false; }
    if (!pw) { qs('#password-error').textContent = 'Password is required.'; ok = false; }
    if (!ok) return;
    submit.disabled = true; submit.textContent = 'Signing in...';
    try {
      await login(em, pw);
      const next = getQuery().next;
      location.href = next && next.startsWith('/') ? next : '/pages/overview.html';
    } catch (err) {
      formError.textContent = err.message || 'Unable to sign in.';
      submit.disabled = false; submit.textContent = 'Sign in';
    }
  });

  app.append(el('div', { class: 'auth-shell' },
    el('main', { class: 'auth-card', id: 'main-content' },
      el('div', { class: 'auth-brand' }, el('div', { class: 'brand-mark', text: 'D' }), el('div', { class: 'brand-name', text: 'DevTrail' })),
      el('h1', { class: 'auth-title', text: 'Sign in' }),
      el('p', { class: 'auth-desc', text: 'Access your engineering intelligence workspace.' }),
      form,
      el('p', { class: 'auth-footer' }, 'New to DevTrail? ', el('a', { href: '/pages/register.html', text: 'Create an account' }))
    )
  ));
}
