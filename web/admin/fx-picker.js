// Visual pickers: the icon pair, the emptying effect and the rank animation are chosen
// from tiles showing the real widget (the overlay in preview mode, one part only),
// instead of a dropdown. The <select> stays as the form value, hidden.
(() => {
  const PICKERS = [
    { name: 'vesselStyle', page: 'look', solo: 'icons', demo: 'win', autoplay: null },
    { name: 'emptyEffect', page: 'motion', solo: 'vessel', demo: 'fillCycle', autoplay: 'all' },
    { name: 'rankFx', page: 'motion', solo: 'rank', demo: 'rankUp', autoplay: 'round' },
  ];
  const formOf = (cfg) => (cfg.page === 'look' ? form : motionForm);
  const ver = () => encodeURIComponent(knownAppVersion || '');
  const panelShown = (cfg) => {
    const p = document.querySelector(`.tab-panel[data-panel="${cfg.page}"]`);
    return !!p && p.offsetParent !== null && !document.hidden;
  };
  const built = new Map(); // name -> { grid, tiles: [{ value, tile, frame }] }

  function tileSettings(name, value) {
    const base = { ...readAppearance(), layout: 'detailed', alwaysVisible: true, __previewLocal: true };
    if (name === 'rankFx') return { ...base, rankFx: value, idlePulse: true };
    return { ...base, [name]: value, fillStyle: base.fillStyle || 'liquid' };
  }

  function post(frame, msg) {
    try { frame.contentWindow?.postMessage(msg, '*'); } catch { /* ignore */ }
  }

  function pushTile(name, t) {
    post(t.frame, { type: 'widget-preview-settings', settings: tileSettings(name, t.value) });
    if (lastSnap) post(t.frame, { type: 'widget-preview-state', state: lastSnap.state, view: lastSnap.view });
  }

  function play(name, t) {
    pushTile(name, t);
    const cfg = PICKERS.find((p) => p.name === name);
    post(t.frame, { type: 'widget-preview-demo', action: cfg.demo });
  }

  function markActive(name) {
    const b = built.get(name);
    if (!b) return;
    const value = formOf(b.cfg)[name]?.value;
    b.tiles.forEach((t) => t.tile.classList.toggle('active', t.value === value));
  }

  function build(cfg) {
    const select = formOf(cfg)[cfg.name];
    if (!select || built.has(cfg.name) || !panelShown(cfg)) return;
    const label = select.closest('label');
    select.classList.add('fx-select-hidden');
    label.classList.add('full', 'fx-pick-label');
    const grid = document.createElement('div');
    grid.className = `fx-picker fx-picker-${cfg.solo}`;
    grid.setAttribute('role', 'radiogroup');
    const tiles = [...select.options].map((opt) => {
      const tile = document.createElement('button');
      tile.type = 'button';
      tile.className = 'fx-tile';
      tile.dataset.value = opt.value;
      tile.innerHTML = `<span class="fx-thumb"><iframe tabindex="-1" aria-hidden="true" loading="lazy"
        src="/overlay/?preview=1&nows=1&still=1&layout=detailed&solo=${cfg.solo}&v=${ver()}"></iframe></span><span class="fx-name"></span>`;
      tile.querySelector('.fx-name').textContent = opt.textContent;
      const frame = tile.querySelector('iframe');
      const t = { value: opt.value, tile, frame, opt };
      frame.addEventListener('load', () => { pushTile(cfg.name, t); setTimeout(() => pushTile(cfg.name, t), 600); });
      tile.addEventListener('mouseenter', () => play(cfg.name, t));
      tile.addEventListener('click', () => {
        select.value = t.value;
        // Same path as using the dropdown: preview, auto-save with Undo.
        select.dispatchEvent(new Event('change', { bubbles: true }));
        markActive(cfg.name);
        pushPreviewSettings();
        sendPreviewDemo(cfg.demo);
      });
      grid.appendChild(tile);
      return t;
    });
    const hint = document.createElement('p');
    hint.className = 'setup-note fx-pick-hint';
    hint.dataset.i18n = 'fxpick.hint';
    hint.textContent = t('fxpick.hint');
    label.after(grid);
    grid.after(hint);
    hint.classList.add('full');
    grid.classList.add('full');
    // Hidden together with their field in layouts that don't use it (data-uses).
    if (label.dataset.uses) {
      grid.dataset.uses = label.dataset.uses;
      hint.dataset.uses = label.dataset.uses;
      grid.hidden = label.hidden;
      hint.hidden = label.hidden;
    }
    built.set(cfg.name, { grid, tiles, cfg, round: 0 });
    markActive(cfg.name);
  }

  // Tiles are built the first time their page opens (iframes are not free).
  function ensureBuilt() {
    PICKERS.forEach(build);
  }
  document.getElementById('tabs')?.addEventListener('click', () => setTimeout(ensureBuilt, 0));
  ensureBuilt();

  // Gentle autoplay while the page is visible: all emptying tiles in turn; rank tiles
  // one at a time (there are many of them).
  setInterval(() => {
    built.forEach((b) => {
      if (b.cfg.autoplay === 'round' && panelShown(b.cfg)) {
        const t = b.tiles[b.round++ % b.tiles.length];
        if (t) play(b.cfg.name, t);
      }
    });
  }, 1500);
  setInterval(() => {
    const b = built.get('emptyEffect');
    if (!b || !panelShown(b.cfg)) return;
    b.tiles.forEach((t, i) => setTimeout(() => play('emptyEffect', t), i * 450));
  }, 5200);

  // Colours / skin / icon style changes show up in the tiles too.
  let restyle = 0;
  const base = pushPreviewSettings;
  pushPreviewSettings = function pushPreviewSettingsTiles(...args) {
    base.apply(this, args);
    clearTimeout(restyle);
    restyle = setTimeout(() => {
      built.forEach((b) => { b.tiles.forEach((t) => pushTile(b.cfg.name, t)); markActive(b.cfg.name); });
    }, 250);
  };
  const sync = () => built.forEach((b) => markActive(b.cfg.name));
  motionForm.addEventListener('change', sync);
  form.addEventListener('change', sync);
  window.addEventListener('admin-i18n', () => built.forEach((b) => b.tiles.forEach((t) => {
    t.tile.querySelector('.fx-name').textContent = t.opt.textContent;
  })));
})();
