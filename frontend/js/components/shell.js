import { el, qs } from '../utils.js';
import { icon } from './icons.js';
import { getUser, logout } from '../auth.js';

const NAV = [
  { group: 'Workspace', items: [
    { key: 'overview', label: 'Overview', href: '/pages/overview.html', icon: 'overview' },
    { key: 'live', label: 'Live Development', href: '/pages/live.html', icon: 'live' },
    { key: 'activity', label: 'Activity Timeline', href: '/pages/activity.html', icon: 'activity' },
  ]},
  { group: 'Development', items: [
    { key: 'projects', label: 'Projects', href: '/pages/projects.html', icon: 'projects' },
    { key: 'sessions', label: 'Sessions', href: '/pages/sessions.html', icon: 'sessions' },
    { key: 'git', label: 'Git Activity', href: '/pages/git-activity.html', icon: 'git' },
    { key: 'testing', label: 'Testing History', href: '/pages/testing.html', icon: 'testing' },
  ]},
  { group: 'Engineering Knowledge', items: [
    { key: 'debugging', label: 'Debugging Journal', href: '/pages/debugging.html', icon: 'debugging' },
    { key: 'decisions', label: 'Decisions', href: '/pages/decisions.html', icon: 'decisions' },
    { key: 'notes', label: 'Technical Notes', href: '/pages/notes.html', icon: 'notes' },
  ]},
  { group: 'Intelligence', items: [
    { key: 'insights', label: 'Engineering Insights', href: '/pages/insights.html', icon: 'insights' },
    { key: 'recurring', label: 'Recurring Problems', href: '/pages/recurring-problems.html', icon: 'recurring' },
    { key: 'improvement', label: 'Improvement Areas', href: '/pages/improvement-areas.html', icon: 'improvement' },
  ]},
  { group: 'Reports', items: [
    { key: 'reports-daily', label: 'Daily Reports', href: '/pages/reports-daily.html', icon: 'reports' },
    { key: 'reports-weekly', label: 'Weekly Reports', href: '/pages/reports-weekly.html', icon: 'reports' },
    { key: 'reports-project', label: 'Project Reports', href: '/pages/reports-project.html', icon: 'reports' },
  ]},
];

export function mountShell({ page, title, description = '', breadcrumbs = [], actions = null }) {
  const user = getUser();
  const app = qs('#app');
  app.innerHTML = '';
  app.className = 'app-shell';

  const sidebar = el('aside', { class: 'sidebar', id: 'sidebar', 'aria-label': 'Primary navigation' });
  const brand = el('div', { class: 'sidebar-brand' },
    el('div', { class: 'brand-mark', text: 'D' }),
    el('div', {}, el('div', { class: 'brand-name', text: 'DevTrail' }), el('span', { class: 'brand-sub', text: 'Engineering intelligence' }))
  );
  const nav = el('nav', { class: 'sidebar-nav' });
  for (const group of NAV) {
    const g = el('div', { class: 'nav-group' }, el('div', { class: 'nav-group-label', text: group.group }));
    for (const item of group.items) {
      const a = el('a', { class: 'nav-item', href: item.href, dataset: { nav: item.key } });
      if (item.key === page) a.setAttribute('aria-current', 'page');
      const ic = el('span', { 'aria-hidden': 'true' }); ic.innerHTML = icon(item.icon);
      a.append(ic, el('span', { class: 'nav-label', text: item.label }));
      g.append(a);
    }
    nav.append(g);
  }

  const footer = el('div', { class: 'sidebar-footer' });
  const settings = el('a', { class: `nav-item ${page === 'settings' ? 'active' : ''}`, href: '/pages/settings.html' });
  if (page === 'settings') settings.setAttribute('aria-current', 'page');
  const sIcon = el('span', { 'aria-hidden': 'true' }); sIcon.innerHTML = icon('settings');
  settings.append(sIcon, el('span', { class: 'nav-label', text: 'Settings' }));

  const logoutBtn = el('button', { class: 'nav-item', type: 'button', style: 'width:100%;margin-top:2px;' });
  const lIcon = el('span', { 'aria-hidden': 'true' }); lIcon.innerHTML = icon('logout');
  logoutBtn.append(lIcon, el('span', { class: 'nav-label', text: 'Logout' }));
  logoutBtn.addEventListener('click', () => logout());

  const userBox = el('div', { class: 'user-menu', style: 'margin-top:0.5rem;' });
  const initials = (user?.display_name || user?.email || 'U').split(/\s+/).filter(Boolean).slice(0,2).map(s=>s[0].toUpperCase()).join('');
  userBox.append(
    el('div', { class: 'user-avatar', text: initials || 'U' }),
    el('div', { class: 'user-meta' },
      el('div', { class: 'user-name', text: user?.display_name || 'User' }),
      el('div', { class: 'user-email', text: user?.email || '' })
    )
  );
  footer.append(settings, logoutBtn, userBox);
  sidebar.append(brand, nav, footer);

  const overlay = el('div', { class: 'sidebar-overlay', id: 'sidebar-overlay' });
  const mainWrap = el('div', { class: 'main-wrap' });
  const topbar = el('header', { class: 'topbar' });
  const menuBtn = el('button', { class: 'mobile-nav-toggle icon-btn', type: 'button', 'aria-label': 'Open navigation', 'aria-controls': 'sidebar', 'aria-expanded': 'false' });
  menuBtn.innerHTML = icon('menu');

  const crumbs = el('div', { class: 'breadcrumbs', 'aria-label': 'Breadcrumb' });
  if (breadcrumbs.length) {
    breadcrumbs.forEach((bc, i) => {
      if (i > 0) crumbs.append(el('span', { class: 'breadcrumb-sep', text: '/' }));
      if (bc.href && i < breadcrumbs.length - 1) crumbs.append(el('a', { href: bc.href, text: bc.label }));
      else crumbs.append(el('span', { class: i === breadcrumbs.length - 1 ? 'current' : '', text: bc.label }));
    });
  } else {
    crumbs.append(el('span', { class: 'current', text: title }));
  }

  topbar.append(
    el('div', { class: 'topbar-left' }, menuBtn, crumbs),
    el('div', { class: 'topbar-right' }, actions || null)
  );

  const content = el('main', { class: 'content', id: 'main-content', tabindex: '-1' });
  const pageHeader = el('div', { class: 'page-header' },
    el('div', { class: 'page-header-main' },
      el('h1', { class: 'page-title', text: title }),
      description ? el('p', { class: 'page-description', text: description }) : null
    ),
    actions ? el('div', { class: 'page-actions' }) : null
  );
  // actions shown in header on mobile? keep in topbar-right; page header gets its own actions slot if passed separately.
  content.append(pageHeader);

  const pageBody = el('div', { id: 'page-body' });
  content.append(pageBody);
  mainWrap.append(topbar, content);
  app.append(sidebar, overlay, mainWrap);

  function closeNav() { sidebar.classList.remove('open'); overlay.classList.remove('open'); menuBtn.setAttribute('aria-expanded', 'false'); }
  menuBtn.addEventListener('click', () => {
    const open = sidebar.classList.toggle('open');
    overlay.classList.toggle('open', open);
    menuBtn.setAttribute('aria-expanded', String(open));
  });
  overlay.addEventListener('click', closeNav);
  nav.addEventListener('click', (e) => { if (e.target.closest('a')) closeNav(); });

  return { content: pageBody, pageHeader, topbarRight: topbar.querySelector('.topbar-right'), setActions(node) {
    const right = topbar.querySelector('.topbar-right'); right.innerHTML = ''; if (node) right.append(node);
  }};
}
