import { requireAuth } from '../auth.js';
import { api } from '../api.js';
import { mountShell } from '../components/shell.js';
import { loadingState, emptyState, errorState } from '../components/states.js';
import { el, clear, safeText } from '../utils.js';

export async function runInsightsPage({ page, title, description, endpoint }) {
  await requireAuth();
  const { content } = mountShell({ page, title, description });
  clear(content).append(loadingState('Loading insights...'));
  try {
    const res = await api.get(endpoint);
    const items = res.insights || [];
    clear(content);
    if (!items.length) { content.append(el('div', { class:'card' }, emptyState({ title:'No insights available', description:'Insights are generated from your recorded engineering history when enough data is available.' }))); return; }
    content.append(el('div', { class:'grid grid-2' }, ...items.map(card)));
  } catch (err) { clear(content).append(errorState({ description: err.message })); }
}

function card(i) {
  const evidence = el('pre', { class:'prose', style:'margin:0;font-size:12px;white-space:pre-wrap;' });
  evidence.textContent = typeof i.evidence === 'string' ? i.evidence : JSON.stringify(i.evidence ?? {}, null, 2);
  return el('article', { class:'card' }, el('div', { class:'card-body' },
    el('div', { style:'display:flex;gap:0.5rem;flex-wrap:wrap;margin-bottom:0.75rem;' },
      el('span', { class:'badge badge-accent', text: safeText(i.type, 'Insight') }),
      el('span', { class:'badge badge-neutral', text: `Scope: ${safeText(i.scope, 'global')}` }),
      el('span', { class:'badge badge-neutral', text: `Confidence: ${safeText(i.confidence, 'unknown')}` }),
    ),
    el('h3', { class:'h3', style:'font-size:1rem;margin-bottom:0.5rem;', text: safeText(i.summary, 'Observation') }),
    el('p', { class:'prose', style:'margin-bottom:0.75rem;', text: safeText(i.observation, '') }),
    i.recommendation ? el('div', { style:'margin-bottom:0.75rem;' }, el('div', { class:'detail-label', text:'Recommendation' }), el('div', { class:'prose', text: safeText(i.recommendation) })) : null,
    el('div', {}, el('div', { class:'detail-label', text:'Evidence' }), el('div', { class:'card', style:'background:var(--bg);border-color:var(--border-subtle);' }, el('div', { class:'card-body', style:'padding:0.75rem;' }, evidence)))
  ));
}
