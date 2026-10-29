/**
 * Admin Help page: quick start, searchable topics in plain language, and a
 * "contact the developer" block. Texts live in i18n.js (help.*).
 * Expects window.AdminI18n.t and callbacks from admin.js.
 */
(function (global) {
  // Where questions and bug reports go. Empty discord hides that button.
  const DEV_CONTACT = {
    // Developer's Discord profile (opens a DM from the profile page).
    discord: 'https://discord.com/users/407943409618255883',
    issues: 'https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/issues',
  };

  const START = [
    { key: 'help.start.obs', goto: 'connections' },
    { key: 'help.start.game', goto: 'main' },
    { key: 'help.start.look', goto: 'layout' },
    { key: 'help.start.test', demo: 'fullCycle' },
  ];

  // group → topics; each topic: id, actions (goto page / preview demo).
  const GROUPS = [
    { key: 'help.g.start', items: [
      { id: 'show', goto: 'connections' },
      { id: 'obs', goto: 'connections' },
      { id: 'saving' },
    ] },
    { key: 'help.g.stream', items: [
      { id: 'count', goto: 'main', demo: 'win' },
      { id: 'rank', goto: 'main', demo: 'rankShowcase' },
      { id: 'session', goto: 'main' },
      { id: 'showhide', goto: 'look' },
    ] },
    { key: 'help.g.look', items: [
      { id: 'layouts', goto: 'layout', demo: 'fullCycle' },
      { id: 'skins', goto: 'skins' },
      { id: 'vessels', goto: 'motion', demo: 'fillCycle' },
      { id: 'card', goto: 'layout' },
    ] },
    { key: 'help.g.auto', items: [
      { id: 'auto', goto: 'auto' },
      { id: 'autorank', goto: 'auto' },
    ] },
    { key: 'help.g.deck', items: [
      { id: 'deck', goto: 'connections' },
    ] },
    { key: 'help.g.trouble', items: [
      { id: 'empty', goto: 'connections' },
      { id: 'hides', goto: 'look' },
      { id: 'wrong', goto: 'main' },
      { id: 'update' },
    ] },
  ];

  function t(key) {
    return (global.AdminI18n && global.AdminI18n.t) ? global.AdminI18n.t(key) : key;
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  // Paragraphs split by a blank line; lines starting with "• " become a list.
  function renderBody(text) {
    return String(text || '').split(/\n\n+/).map((block) => {
      const lines = block.split('\n');
      if (lines.every((l) => l.startsWith('• '))) {
        return `<ul>${lines.map((l) => `<li>${escapeHtml(l.slice(2))}</li>`).join('')}</ul>`;
      }
      return `<p>${escapeHtml(block).replace(/\n/g, '<br>')}</p>`;
    }).join('');
  }

  function actionsHtml(item) {
    const out = [];
    if (item.goto) out.push(`<button type="button" class="ghost help-go" data-goto="${item.goto}">${t('help.open')} →</button>`);
    if (item.demo) out.push(`<button type="button" class="ghost help-demo" data-demo="${item.demo}">${t('help.showMe')}</button>`);
    return out.length ? `<div class="faq-actions">${out.join('')}</div>` : '';
  }

  function render(listEl) {
    const start = document.getElementById('helpStart');
    if (start) {
      start.innerHTML = `<h3>${escapeHtml(t('help.startTitle'))}</h3><ol class="help-steps">${START.map((s, i) => `
        <li><span class="help-step-n">${i + 1}</span><span class="help-step-text">${escapeHtml(t(s.key))}</span>
        ${s.goto ? `<button type="button" class="ghost help-go" data-goto="${s.goto}">${t('help.open')} →</button>` : ''}
        ${s.demo ? `<button type="button" class="ghost help-demo" data-demo="${s.demo}">${t('help.showMe')}</button>` : ''}</li>`).join('')}</ol>`;
    }
    listEl.innerHTML = GROUPS.map((g) => `
      <div class="help-group" data-group="${g.key}">
        <h3 class="help-group-title">${escapeHtml(t(g.key))}</h3>
        ${g.items.map((item) => `
        <details class="faq-item" data-faq-id="${item.id}">
          <summary><span class="faq-title">${escapeHtml(t(`help.${item.id}.q`))}</span></summary>
          <div class="faq-body"><div class="faq-text">${renderBody(t(`help.${item.id}.a`))}</div>${actionsHtml(item)}</div>
        </details>`).join('')}
      </div>`).join('');
    const discord = document.getElementById('helpDiscord');
    if (discord) {
      discord.hidden = !DEV_CONTACT.discord;
      if (DEV_CONTACT.discord) discord.href = DEV_CONTACT.discord;
    }
    const issues = document.getElementById('helpIssues');
    if (issues) issues.href = DEV_CONTACT.issues;
    applySearch();
  }

  function applySearch() {
    const input = document.getElementById('helpSearch');
    const listEl = document.getElementById('faqList');
    if (!input || !listEl) return;
    const q = input.value.trim().toLowerCase();
    let shown = 0;
    listEl.querySelectorAll('.help-group').forEach((group) => {
      let groupShown = 0;
      group.querySelectorAll('.faq-item').forEach((item) => {
        const hit = !q || item.textContent.toLowerCase().includes(q);
        item.hidden = !hit;
        if (hit) groupShown++;
        // While searching, open the matches so the answer is visible.
        if (q) item.open = hit;
      });
      group.hidden = groupShown === 0;
      shown += groupShown;
    });
    const empty = document.getElementById('helpEmpty');
    if (empty) empty.hidden = shown > 0;
  }

  // Everything a bug report needs, in one paste.
  function diagnostics() {
    const snap = typeof lastSnap !== 'undefined' ? lastSnap : null;
    const s = snap?.settings || {};
    const v = snap?.view || {};
    const obsOk = document.getElementById('pillObs')?.classList.contains('on');
    const autoOn = document.getElementById('autoToggle')?.checked;
    return [
      `Widget Stats ${document.getElementById('appVersion')?.textContent || '?'}`,
      `Game: ${v.game || '?'} · mode: ${v.mode || '?'} · role: ${v.role || '-'}`,
      `Layout: ${s.layout || '?'} · skin: ${s.skinId || '?'} · rank FX: ${s.rankFx || '?'}`,
      `OBS connected: ${obsOk ? 'yes' : 'no'} · Auto W/L: ${autoOn ? 'on' : 'off'}`,
      `Browser: ${navigator.userAgent}`,
    ].join('\n');
  }

  function init(opts = {}) {
    const listEl = document.getElementById('faqList');
    if (!listEl) return;
    render(listEl);

    const runDemo = (action) => {
      if (typeof opts.sendPreviewDemo === 'function') opts.sendPreviewDemo(action);
      else if (typeof global.sendPreviewDemo === 'function') global.sendPreviewDemo(action);
      if (typeof opts.onDemo === 'function') opts.onDemo(action);
    };

    const panel = listEl.closest('[data-panel]') || document;
    panel.addEventListener('click', async (e) => {
      const demo = e.target.closest('.help-demo');
      if (demo) {
        e.preventDefault();
        runDemo(demo.dataset.demo);
        return;
      }
      const go = e.target.closest('.help-go');
      if (go) {
        e.preventDefault();
        document.querySelector(`.tab[data-tab="${go.dataset.goto}"]`)?.click();
        return;
      }
      if (e.target.closest('#helpCopyInfo')) {
        const btn = e.target.closest('#helpCopyInfo');
        try {
          await navigator.clipboard.writeText(diagnostics());
          btn.textContent = t('help.copied');
        } catch {
          btn.textContent = t('help.copyFailed');
        }
        setTimeout(() => { btn.textContent = t('help.copyInfo'); }, 2500);
      }
    });

    listEl.addEventListener('toggle', (e) => {
      const item = e.target;
      if (!(item instanceof HTMLDetailsElement) || !item.open) return;
      if (document.getElementById('helpSearch')?.value.trim()) return;
      // Accordion when browsing: one answer open at a time.
      listEl.querySelectorAll('details.faq-item').forEach((d) => { if (d !== item) d.open = false; });
    }, true);

    document.getElementById('helpSearch')?.addEventListener('input', applySearch);

    global.AdminFAQ.refreshI18n = () => render(listEl);
  }

  global.AdminFAQ = { init, refreshI18n: () => {}, contact: DEV_CONTACT };
})(typeof window !== 'undefined' ? window : globalThis);
