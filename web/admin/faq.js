/**
 * Admin FAQ / help tab — accordion + live preview demos.
 * Expects window.AdminI18n.t and callbacks from admin.js.
 */
(function (global) {
  const ITEMS = [
    {
      id: 'start',
      icon: '1',
      titleKey: 'faq.start.title',
      bodyKey: 'faq.start.body',
      demo: 'fullCycle',
      demoLabelKey: 'faq.demo.full',
    },
    {
      id: 'obs',
      icon: '2',
      titleKey: 'faq.obs.title',
      bodyKey: 'faq.obs.body',
    },
    {
      id: 'games',
      icon: '3',
      titleKey: 'faq.games.title',
      bodyKey: 'faq.games.body',
    },
    {
      id: 'look',
      icon: '4',
      titleKey: 'faq.look.title',
      bodyKey: 'faq.look.body',
      demo: 'fullCycle',
      demoLabelKey: 'faq.demo.look',
    },
    {
      id: 'vessels',
      icon: '5',
      titleKey: 'faq.vessels.title',
      bodyKey: 'faq.vessels.body',
      demo: 'fillCycle',
      demoLabelKey: 'faq.demo.fill',
    },
    {
      id: 'motion',
      icon: '6',
      titleKey: 'faq.motion.title',
      bodyKey: 'faq.motion.body',
      demos: [
        { action: 'win', labelKey: 'faq.demo.win' },
        { action: 'loss', labelKey: 'faq.demo.loss' },
        { action: 'fillCycle', labelKey: 'faq.demo.fill' },
      ],
    },
    {
      id: 'rankfx',
      icon: '7',
      titleKey: 'faq.rankfx.title',
      bodyKey: 'faq.rankfx.body',
      demo: 'rankShowcase',
      demoLabelKey: 'faq.demo.rank',
    },
    {
      id: 'skins',
      icon: '8',
      titleKey: 'faq.skins.title',
      bodyKey: 'faq.skins.body',
      demo: 'fullCycle',
      demoLabelKey: 'faq.demo.skin',
    },
    {
      id: 'deck',
      icon: '9',
      titleKey: 'faq.deck.title',
      bodyKey: 'faq.deck.body',
    },
    {
      id: 'tips',
      icon: 'i',
      titleKey: 'faq.tips.title',
      bodyKey: 'faq.tips.body',
    },
  ];

  function t(key) {
    return (global.AdminI18n && global.AdminI18n.t) ? global.AdminI18n.t(key) : key;
  }

  function renderBody(text) {
    // Allow simple paragraphs separated by \n\n
    return String(text || '')
      .split(/\n\n+/)
      .map((p) => `<p>${escapeHtml(p).replace(/\n/g, '<br>')}</p>`)
      .join('');
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function buildList(listEl) {
    listEl.innerHTML = '';
    ITEMS.forEach((item, idx) => {
      const details = document.createElement('details');
      details.className = 'faq-item';
      details.dataset.faqId = item.id;
      if (idx === 0) details.open = true;

      const summary = document.createElement('summary');
      summary.innerHTML = `<span class="faq-num">${item.icon}</span><span class="faq-title" data-i18n="${item.titleKey}">${t(item.titleKey)}</span>`;
      details.appendChild(summary);

      const body = document.createElement('div');
      body.className = 'faq-body';
      body.innerHTML = `<div class="faq-text" data-i18n-html="${item.bodyKey}">${renderBody(t(item.bodyKey))}</div>`;

      const actions = document.createElement('div');
      actions.className = 'faq-actions';
      const demos = item.demos || (item.demo ? [{ action: item.demo, labelKey: item.demoLabelKey || 'faq.demo.run' }] : []);
      demos.forEach((d) => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'ghost faq-demo-btn';
        btn.dataset.demo = d.action;
        btn.setAttribute('data-i18n', d.labelKey);
        btn.textContent = t(d.labelKey);
        actions.appendChild(btn);
      });
      if (demos.length) body.appendChild(actions);
      details.appendChild(body);
      listEl.appendChild(details);
    });
  }

  function refreshI18n(listEl) {
    if (!listEl) return;
    listEl.querySelectorAll('[data-i18n]').forEach((el) => {
      el.textContent = t(el.getAttribute('data-i18n'));
    });
    listEl.querySelectorAll('[data-i18n-html]').forEach((el) => {
      el.innerHTML = renderBody(t(el.getAttribute('data-i18n-html')));
    });
  }

  function init(opts = {}) {
    const listEl = document.getElementById('faqList');
    const hintEl = document.getElementById('faqDemoHint');
    if (!listEl) return;

    buildList(listEl);

    const runDemo = (action) => {
      if (typeof opts.sendPreviewDemo === 'function') {
        opts.sendPreviewDemo(action);
      } else if (typeof global.sendPreviewDemo === 'function') {
        global.sendPreviewDemo(action);
      }
      if (hintEl) {
        hintEl.textContent = t('faq.preview.playing');
        setTimeout(() => { hintEl.textContent = t('faq.preview.hint'); }, 2200);
      }
      if (typeof opts.onDemo === 'function') opts.onDemo(action);
    };

    listEl.addEventListener('click', (e) => {
      const btn = e.target.closest('.faq-demo-btn');
      if (!btn) return;
      e.preventDefault();
      e.stopPropagation();
      runDemo(btn.dataset.demo);
    });

    listEl.addEventListener('toggle', (e) => {
      const item = e.target;
      if (!(item instanceof HTMLDetailsElement) || !item.open) return;
      // Accordion: close siblings
      listEl.querySelectorAll('details.faq-item').forEach((d) => {
        if (d !== item) d.open = false;
      });
    });

    document.getElementById('faqReplayDemoBtn')?.addEventListener('click', () => {
      runDemo('fullCycle');
    });

    global.AdminFAQ = global.AdminFAQ || {};
    global.AdminFAQ.refreshI18n = () => refreshI18n(listEl);
  }

  global.AdminFAQ = { ITEMS, init, refreshI18n: () => {} };
})(typeof window !== 'undefined' ? window : globalThis);
