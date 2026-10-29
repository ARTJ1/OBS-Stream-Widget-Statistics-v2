const OW_RANKS = [
  ...generateRanks('Bronze', 5, 'assets/uploads/Bronze.png'),
  ...generateRanks('Silver', 5, 'assets/uploads/Silver.png'),
  ...generateRanks('Gold', 5, 'assets/uploads/Gold.png'),
  ...generateRanks('Platinum', 5, 'assets/uploads/Platinum.png'),
  ...generateRanks('Emerald', 5, 'assets/uploads/Emerald.png'),
  ...generateRanks('Diamond', 5, 'assets/uploads/Diamond.png'),
  ...generateRanks('Master', 5, 'assets/uploads/Master.png'),
  ...generateRanks('Grandmaster', 5, 'assets/uploads/Grandmaster.png'),
  ...generateRanks('Champion', 5, 'assets/uploads/Champion.png'),
  ...Array.from({ length: 500 }, (_, i) => ({
    type: 'Top',
    level: 500 - i,
    img: 'assets/uploads/Top_500.png',
  })),
];

const LAYOUTS = [
  'detailed', 'mini', 'card', 'badge', 'strip', 'minimal', 'column',
  'ring', 'slant', 'scoreboard', 'session', 'form', 'ladder', 'roles', 'hex',
  'feed', 'tug', 'flip', 'tcard',
];
// Layouts with little room use abbreviated tier names ("Dia 3", "GM 1").
const SHORT_NAME_LAYOUTS = new Set(['mini', 'column', 'scoreboard']);
const RANK_SHORT = {
  Rookie: 'Rook', Bronze: 'Bro', Silver: 'Sil', Gold: 'Gold', Platinum: 'Plat', Emerald: 'Em',
  Diamond: 'Dia', Master: 'Mas', Grandmaster: 'GM', Champion: 'Champ',
};

const APEX_ROMAN = { 4: 'IV', 3: 'III', 2: 'II', 1: 'I' };
const APEX_RANKS = [
  ...generateApexTier('Rookie'),
  ...generateApexTier('Bronze'),
  ...generateApexTier('Silver'),
  ...generateApexTier('Gold'),
  ...generateApexTier('Platinum'),
  ...generateApexTier('Diamond'),
  ...generateApexTier('Master'),
  ...Array.from({ length: 750 }, (_, i) => ({
    type: 'Predator',
    level: 750 - i,
    img: 'assets/apex/Predator.webp',
  })),
];

function generateRanks(type, levels, img) {
  return Array.from({ length: levels }, (_, i) => ({
    type,
    level: levels - i,
    img,
  }));
}

function generateApexTier(type) {
  return Array.from({ length: 4 }, (_, i) => {
    const level = 4 - i;
    return {
      type,
      level,
      roman: APEX_ROMAN[level],
      img: `assets/apex/${type}.webp`,
    };
  });
}

document.addEventListener('DOMContentLoaded', () => {
  const isPreview = new URLSearchParams(location.search).get('preview') === '1';
  if (isPreview) {
    document.documentElement.classList.add('preview-mode');
    document.body.style.background = 'transparent';
  }

  const dom = {
    widget: document.getElementById('widget'),
    widgetFxLayer: document.getElementById('widgetFxLayer'),
    wins: document.getElementById('wins'),
    losses: document.getElementById('losses'),
    winsIconContainer: document.getElementById('winsIconContainer'),
    lossesIconContainer: document.getElementById('lossesIconContainer'),
    winsLiquid: document.getElementById('winsLiquid'),
    lossesLiquid: document.getElementById('lossesLiquid'),
    rankAnimationWrapper: document.getElementById('rank-animation-wrapper'),
    rankFxLayer: document.getElementById('rankFxLayer'),
    rankSection: document.getElementById('rankSection'),
    rankImage: document.getElementById('rankImage'),
    rankImageUnder: document.getElementById('rankImageUnder'),
    rankValue: document.getElementById('rankValue'),
    roleBadge: document.getElementById('roleBadge'),
    roleIcon: document.getElementById('roleIcon'),
    wrFill: document.getElementById('wrFill'),
    wrRingWin: document.getElementById('wrRingWin'),
    stats: document.getElementById('stats'),
    rankDelta: document.getElementById('rankDelta'),
    formPips: document.getElementById('formPips'),
    ladderSteps: document.getElementById('ladderSteps'),
    ladderNext: document.getElementById('ladderNext'),
    rolesList: document.getElementById('rolesList'),
    feedList: document.getElementById('feedList'),
    tugKnot: document.getElementById('tugKnot'),
    tugNet: document.getElementById('tugNet'),
    flipDigits: document.getElementById('flipDigits'),
    tcardStars: document.getElementById('tcardStars'),
    tcardBack: document.getElementById('tcardBack'),
    tcardShine: document.getElementById('tcardShine'),
    wrValue: document.getElementById('wrValue'),
  };

  const ROLE_ICONS = {
    tank: 'assets/roles/tank.png',
    support: 'assets/roles/support.png',
    damage: 'assets/roles/damage.png',
  };

  // Query-string URLs (?preview=1) break svg clip-path url(#id) in Chromium —
  // liquid won't clip to trophy/skull. Rewrite to absolute fragment refs.
  function fixSvgClipRefs() {
    const base = `${location.pathname}${location.search}`;
    document.querySelectorAll('svg [clip-path]').forEach((el) => {
      const raw = el.getAttribute('clip-path') || '';
      const m = /#([^)'"\s]+)/.exec(raw);
      if (m) el.setAttribute('clip-path', `url("${base}#${m[1]}")`);
    });
  }
  fixSvgClipRefs();

  const fillLimitDefault = 10;
  let fillLimit = fillLimitDefault;
  let fillDurationMs = 650;
  let emptyDurationMs = 1600;
  let emptyEffect = 'drain';
  let numberAnimMs = 500;
  let vesselWave = true;
  let idlePulse = true;
  let appearEffect = 'slide';
  let fillStyle = 'liquid';
  let rankFx = 'classic';
  let vesselStyleId = 'classic';
  let layout = 'detailed';
  let isRankTransitionAnimating = false;
  let showToken = 0;
  let hideWidgetTimeout;
  let reappearTimeout;
  let animDurationIn = 0.5;
  let animStayTime = 10;
  let animDurationOut = 0.5;
  let reappearDelayMs = 8000;
  let animDirection = 'left';
  let alwaysVisible = false;
  let waveRaf = 0;
  let waveTime = 0;

  let currentState = { wins: 0, losses: 0, rank: 0, mode: 'classic', role: 'tank', game: 'overwatch' };
  let previousState = { ...currentState };
  let settings = null;

  const sepEl = document.getElementById('statSep');

  function activeRanks() {
    return currentState.game === 'apex' ? APEX_RANKS : OW_RANKS;
  }

  function getRankInfo(absoluteRankIndex) {
    const ranks = activeRanks();
    if (absoluteRankIndex < 0 || absoluteRankIndex >= ranks.length || !ranks[absoluteRankIndex]) {
      const fallback = currentState.game === 'apex' ? 'assets/apex/Rookie.webp' : 'assets/uploads/Bronze.png';
      return { type: 'Unranked', level: 0, img: fallback, display: 'Unranked' };
    }
    const rank = ranks[absoluteRankIndex];
    if (rank.type === 'Top' || rank.type === 'Predator') {
      return { ...rank, display: `#${rank.level}` };
    }
    const name = SHORT_NAME_LAYOUTS.has(layout) ? (RANK_SHORT[rank.type] || rank.type) : rank.type;
    if (rank.roman) {
      return { ...rank, display: `${name} ${rank.roman}` };
    }
    return { ...rank, display: `${name} ${rank.level}` };
  }

  function setLiquidPaint(liquidEl, color) {
    if (!liquidEl) return;
    liquidEl.querySelectorAll('.liquid-body, .liquid-meniscus').forEach((el) => {
      el.setAttribute('fill', color);
    });
  }

  // Opaque copy of the panel colour (cards and flip tiles need a solid face) and a
  // light/dark flag, so layout details adapt to light skins.
  function applySkinTone(bgColor) {
    const probe = document.createElement('span');
    probe.style.color = bgColor;
    document.body.appendChild(probe);
    const m = (getComputedStyle(probe).color.match(/[\d.]+/g) || ['0', '0', '0']).map(Number);
    probe.remove();
    const [r, g, b] = m;
    document.documentElement.style.setProperty('--bg-solid', `rgb(${r}, ${g}, ${b})`);
    const lin = (v) => { v /= 255; return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4; };
    const lum = 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
    document.body.classList.toggle('skin-light', lum > 0.4);
  }

  // Fixed-size layouts (hex, card, badge) shrink a long rank name to fit
  // instead of letting wide fonts run past the edge.
  const FIT_NAME_LAYOUTS = { hex: 0.8, tcard: 0.9 };
  // Hex narrows towards the bottom: the score row may use about half the width.
  const FIT_STATS_LAYOUTS = { hex: 0.5 };
  function fitToWidth(el, share) {
    el.style.fontSize = '';
    if (!share || !el.textContent) return;
    const room = dom.widget.clientWidth * share;
    const need = el.scrollWidth;
    if (need > room) {
      const base = parseFloat(getComputedStyle(el).fontSize);
      el.style.fontSize = `${Math.max(4, (base * room) / need).toFixed(2)}px`;
    }
  }
  function fitRankName() {
    fitToWidth(dom.rankValue, FIT_NAME_LAYOUTS[layout]);
    if (dom.stats) fitToWidth(dom.stats, FIT_STATS_LAYOUTS[layout]);
  }

  function applyLayout() {
    document.body.classList.remove(...LAYOUTS.map((id) => `layout-${id}`));
    document.body.classList.add(`layout-${layout}`);
    // Short/full rank names differ per layout; a running rank animation sets its own text.
    if (!isRankTransitionAnimating) dom.rankValue.textContent = getRankInfo(currentState.rank).display;
    renderLayoutExtras();
  }

  // Win rate follows whatever numbers are on screen (including demo/animation steps).
  function updateWinRate() {
    const w = parseInt(dom.wins.textContent, 10) || 0;
    const l = parseInt(dom.losses.textContent, 10) || 0;
    const pct = w + l > 0 ? Math.round((100 * w) / (w + l)) : 0;
    // No games yet: a dash, not "0%" (which reads as a losing record).
    const label = w + l > 0 ? `${pct}%` : '–';
    if (dom.wrValue) dom.wrValue.textContent = label;
    if (dom.wrFill) dom.wrFill.style.width = `${w + l > 0 ? pct : 50}%`;
    if (dom.stats) dom.stats.dataset.wr = label;
    if (dom.wrRingWin) {
      const c = 2 * Math.PI * 33;
      const share = w + l > 0 ? w / (w + l) : 0.5;
      dom.wrRingWin.setAttribute('stroke-dasharray', `${(c * share).toFixed(1)} ${c.toFixed(1)}`);
    }
  }

  // Divisions gained/lost since the session started (app launch or W/L reset).
  function updateRankDelta() {
    if (!dom.rankDelta) return;
    const start = Number.isFinite(currentState.startRank) ? currentState.startRank : currentState.rank;
    const d = (currentState.rank || 0) - start;
    dom.rankDelta.textContent = d > 0 ? `▲ +${d}` : d < 0 ? `▼ ${d}` : '±0';
    dom.rankDelta.classList.toggle('up', d > 0);
    dom.rankDelta.classList.toggle('down', d < 0);
  }

  // Last 5 results as tiles, newest last, plus the current streak.
  function renderFormPips() {
    if (!dom.formPips || layout !== 'form') return;
    const recent = (currentState.recent || []).slice(-5);
    const pips = [];
    for (let i = 0; i < 5 - recent.length; i++) pips.push('<span class="pip"></span>');
    recent.forEach((r, i) => {
      const last = i === recent.length - 1 ? ' last' : '';
      pips.push(`<span class="pip ${r === 'w' ? 'w' : 'l'}${last}">${r === 'w' ? 'W' : 'L'}</span>`);
    });
    const all = currentState.recent || [];
    let streak = 0;
    for (let i = all.length - 1; i >= 0 && all[i] === all[all.length - 1]; i--) streak++;
    const label = streak >= 2 ? `<span class="streak ${all[all.length - 1]}">${streak}${all[all.length - 1] === 'w' ? 'W' : 'L'} streak</span>` : '';
    dom.formPips.innerHTML = pips.join('') + label;
  }

  // Division steps inside the current tier, and the next tier's icon.
  function renderLadder() {
    if (!dom.ladderSteps || layout !== 'ladder') return;
    const ranks = activeRanks();
    const idx = currentState.rank || 0;
    const rank = ranks[idx];
    const top = !rank || rank.type === 'Top' || rank.type === 'Predator';
    dom.ladderSteps.hidden = top;
    dom.ladderNext.hidden = top;
    if (top) return;
    const size = currentState.game === 'apex' ? 4 : 5;
    const steps = [];
    for (let d = size; d >= 1; d--) {
      steps.push(`<span class="step${d > rank.level ? ' on' : d === rank.level ? ' cur' : ''}"></span>`);
    }
    dom.ladderSteps.innerHTML = steps.join('');
    const next = ranks[idx + rank.level];
    dom.ladderNext.hidden = !next;
    if (next) dom.ladderNext.src = next.img;
  }

  // Roles layout: the played role uses the real (animated) rank block; the other
  // two roles are drawn here as static rows. Rows keep the in-game order.
  const ROLES_ROWS = ['tank', 'damage', 'support'];
  function renderRolesList() {
    if (!dom.rolesList || layout !== 'roles') return;
    const roles = currentState.roles;
    const byRole = currentState.game !== 'apex' && currentState.mode !== 'classic' && roles;
    const active = ROLES_ROWS.indexOf(currentState.role);
    dom.widget.style.setProperty('--active-row', String(byRole && active >= 0 ? active + 1 : 1));
    if (!byRole) {
      dom.rolesList.innerHTML = '';
      return;
    }
    const split = currentState.mode === 'roles_split';
    dom.rolesList.innerHTML = '<span class="roles-hl"></span>' + ROLES_ROWS.map((id, i) => {
      if (id === currentState.role) return '';
      const r = roles[id] || {};
      const info = getRankInfo(r.rank || 0);
      const wl = split ? `<span class="wl-cell"><span class="w">${r.wins || 0}</span>–<span class="l">${r.losses || 0}</span></span>` : '';
      return `<span class="role-row" style="--row:${i + 1}">`
        + `<span class="role-icon role-cell" style="--role-icon-url:url('${ROLE_ICONS[id]}')"></span>`
        + `<img class="rank-cell" src="${info.img}" alt="">`
        + `<span class="name-cell">${info.display}</span>${wl}</span>`;
    }).join('');
  }

  // Kill feed: newest result on top (next to the live rank), two older ones below.
  function renderFeed() {
    if (!dom.feedList || layout !== 'feed') return;
    const recent = (currentState.recent || []).slice(-3).reverse();
    const rows = recent.length ? recent : [''];
    dom.feedList.innerHTML = rows.map((r, i) => {
      const cls = r === 'w' ? 'w' : r === 'l' ? 'l' : 'none';
      const tag = r === 'w' ? 'VICTORY' : r === 'l' ? 'DEFEAT' : 'NO GAMES YET';
      return `<span class="feed-bg ${cls}" style="--row:${i + 1}"></span>`
        + `<span class="feed-tag ${cls}" style="--row:${i + 1}">${tag}</span>`;
    }).join('');
  }

  // Tug of war: the knot sits toward whichever side is ahead.
  function renderTug() {
    if (!dom.tugKnot || layout !== 'tug') return;
    const w = parseInt(dom.wins.textContent, 10) || 0;
    const l = parseInt(dom.losses.textContent, 10) || 0;
    const total = w + l;
    dom.tugKnot.style.left = `${total ? 50 + (50 * (w - l)) / total : 50}%`;
    const net = w - l;
    dom.tugNet.textContent = net > 0 ? `+${net} AHEAD` : net < 0 ? `${net} BEHIND` : 'EVEN';
  }

  // Flip clock: split-flap tiles. Tiles persist between renders and only a digit
  // that actually changed flips: numbers are rewritten every frame while they count
  // up, so rebuilding the tiles would cancel the animation one frame in.
  const flipHalf = (cls, d) => `<span class="h ${cls}"><b>${d}</b></span>`;
  function flipTile(tile, d) {
    const old = tile.dataset.d;
    if (old === d) return;
    tile.dataset.d = d;
    if (old === undefined) {
      tile.innerHTML = flipHalf('t', d) + flipHalf('b', d);
      return;
    }
    // New top shows behind the falling old top; old bottom waits under the landing flap.
    tile.innerHTML = flipHalf('t', d) + flipHalf('b', old) + flipHalf('t ft', old) + flipHalf('b fb', d);
    // Settle on the new digit when the flip ends, and also after its 0.4s run in
    // case animations are paused (hidden OBS source, background tab) and never end.
    const settle = () => {
      if (tile.dataset.d === d && tile.querySelector('.fb')) tile.innerHTML = flipHalf('t', d) + flipHalf('b', d);
    };
    tile.querySelector('.fb').addEventListener('animationend', settle, { once: true });
    setTimeout(settle, 500);
  }
  function renderFlipSide(kind, value) {
    let side = dom.flipDigits.querySelector(`.flip-side.${kind}`);
    if (!side) {
      dom.flipDigits.innerHTML = '<span class="flip-side w"></span><span class="flip-colon">:</span><span class="flip-side l"></span>';
      side = dom.flipDigits.querySelector(`.flip-side.${kind}`);
    }
    const digits = String(value).padStart(2, '0').split('');
    while (side.children.length < digits.length) side.insertAdjacentHTML('afterbegin', `<span class="flip-d ${kind}"></span>`);
    while (side.children.length > digits.length) side.firstElementChild.remove();
    digits.forEach((d, i) => flipTile(side.children[i], d));
  }
  function renderFlip() {
    if (!dom.flipDigits || layout !== 'flip') return;
    renderFlipSide('w', parseInt(dom.wins.textContent, 10) || 0);
    renderFlipSide('l', parseInt(dom.losses.textContent, 10) || 0);
  }

  // Trading card: 1-5 stars by tier, spread over the game's tier ladder. A star
  // gained since the last render pops in.
  let cardStarsShown = 0;
  function renderCardStars() {
    if (!dom.tcardStars || layout !== 'tcard') return;
    const ranks = activeRanks();
    const tiers = [...new Set(ranks.map((r) => r.type))];
    const rank = ranks[currentState.rank || 0];
    const tier = Math.max(0, tiers.indexOf(rank ? rank.type : tiers[0]));
    const stars = Math.max(1, Math.ceil(((tier + 1) * 5) / tiers.length));
    if (stars === cardStarsShown && dom.tcardStars.children.length) return;
    const gained = cardFx.has('stars') && cardStarsShown > 0 && stars > cardStarsShown;
    dom.tcardStars.innerHTML = [0, 1, 2, 3, 4].map((i) => {
      const cls = i < stars ? 'on' : '';
      const pop = gained && i >= cardStarsShown && i < stars ? ' pop' : '';
      return `<span class="${cls}${pop}">★</span>`;
    }).join('');
    cardStarsShown = stars;
  }

  // Trading card: the foil heats up with a win streak (spins faster, glows).
  function renderCardHeat() {
    if (layout !== 'tcard') return;
    const recent = currentState.recent || [];
    let streak = 0;
    for (let i = recent.length - 1; i >= 0 && recent[i] === 'w'; i--) streak++;
    dom.widget.style.setProperty('--heat', String(cardFx.has('heat') ? Math.min(streak, 5) : 0));
  }

  // Trading card: a glare sweeps across after a result or rank change.
  function playCardShine() {
    if (layout !== 'tcard' || !cardFx.has('shine') || !dom.tcardShine) return;
    dom.tcardShine.classList.remove('go');
    void dom.tcardShine.offsetWidth;
    dom.tcardShine.classList.add('go');
  }

  // Which card reactions are on (settings.cardFx); float is CSS-only via a body class.
  const CARD_FX = ['stars', 'flip', 'heat', 'shine', 'float'];
  let cardFx = new Set(['stars', 'flip', 'heat']);
  function applyCardFx(value) {
    const list = String(value || 'stars,flip,heat').split(',').map((s) => s.trim());
    cardFx = new Set(list.filter((id) => CARD_FX.includes(id)));
    CARD_FX.forEach((id) => document.body.classList.toggle(`card-fx-${id}`, cardFx.has(id)));
  }

  // Card back text shrinks to the card's width (wide and pixel fonts run long).
  function fitCardBack() {
    const room = dom.tcardBack.clientWidth * 0.86;
    dom.tcardBack.querySelectorAll('b, small').forEach((el) => {
      el.style.fontSize = '';
      const need = el.scrollWidth;
      if (room > 0 && need > room) {
        const base = parseFloat(getComputedStyle(el).fontSize);
        el.style.fontSize = `${Math.max(6, Math.floor((base * room) / need))}px`;
      }
    });
  }

  // Trading card: turn over to show the result, then turn back.
  let cardFlipTimers = [];
  function playCardFlip(kind) {
    playCardShine();
    if (layout !== 'tcard' || !cardFx.has('flip') || !dom.tcardBack) return;
    cardFlipTimers.forEach(clearTimeout);
    const recent = currentState.recent || [];
    let streak = 0;
    for (let i = recent.length - 1; i >= 0 && recent[i] === 'w'; i--) streak++;
    dom.tcardBack.innerHTML = kind === 'w'
      ? `<b class="w">+1 WIN</b><small>${streak >= 2 ? `${streak} IN A ROW` : 'NICE'}</small>`
      : '<b class="l">LOSS</b><small>NEXT ONE</small>';
    dom.widget.classList.remove('card-flip', 'card-back-on');
    void dom.widget.offsetWidth;
    dom.widget.classList.add('card-flip');
    // Faces swap while the card is edge-on (12.5% and 87.5% of the 1.6s flip).
    cardFlipTimers = [
      setTimeout(() => {
        dom.widget.classList.add('card-back-on');
        fitCardBack();
      }, 200),
      setTimeout(() => dom.widget.classList.remove('card-back-on'), 1400),
      setTimeout(() => dom.widget.classList.remove('card-flip'), 1600),
    ];
  }

  function renderLayoutExtras() {
    renderFormPips();
    renderLadder();
    renderRolesList();
    renderFeed();
    renderTug();
    renderFlip();
    renderCardStars();
    renderCardHeat();
    fitRankName();
    fitPreviewWidget();
  }

  const wrObserver = new MutationObserver(() => {
    updateWinRate();
    updateRankDelta();
    renderLayoutExtras();
  });
  [dom.wins, dom.losses, dom.rankValue].forEach((el) => wrObserver.observe(el, { childList: true, characterData: true, subtree: true }));
  updateWinRate();
  updateRankDelta();
  // Web fonts arrive after the first layout: re-fit names and the preview zoom then.
  if (document.fonts && document.fonts.addEventListener) {
    document.fonts.addEventListener('loadingdone', () => {
      fitRankName();
      fitPreviewWidget();
    });
  }

  function applyAppearEffect() {
    dom.widget.classList.remove('appear-slide', 'appear-fade', 'appear-bounce', 'appear-zoom');
    dom.widget.classList.add(`appear-${appearEffect || 'slide'}`);
  }

  function applyFillStyleClass() {
    document.body.classList.remove('fill-liquid', 'fill-solid', 'fill-glow', 'fill-bubble');
    document.body.classList.add(`fill-${fillStyle || 'liquid'}`);
  }

  function applyVesselIcon(kind, styleId) {
    const shapes = window.VesselShapes;
    if (!shapes) return;
    const shape = shapes.getSide ? shapes.getSide(styleId || 'classic', kind === 'wins' ? 'wins' : 'losses') : null;
    if (!shape || !shape.d) return;
    const isWins = kind === 'wins';
    const container = isWins ? dom.winsIconContainer : dom.lossesIconContainer;
    if (!container) return;
    const svg = container.querySelector('svg.stat-icon');
    const pathId = isWins ? 'trophyPath' : 'skullPath';
    const clipId = isWins ? 'winsShapeClip' : 'lossesShapeClip';
    const pathEl = document.getElementById(pathId);
    const clipPath = document.querySelector(`#${clipId} path`);
    if (!svg || !pathEl || !clipPath) return;
    const viewBox = shape.viewBox || '0 0 512 512';
    svg.setAttribute('viewBox', viewBox);
    pathEl.setAttribute('d', shape.d);
    clipPath.setAttribute('d', shape.d);
    const rule = shape.fillRule || 'nonzero';
    pathEl.setAttribute('fill-rule', rule);
    pathEl.setAttribute('clip-rule', rule);
    clipPath.setAttribute('clip-rule', rule);
    clipPath.setAttribute('fill-rule', rule);
    container.querySelectorAll('.vessel-outline').forEach((el) => {
      el.setAttribute('stroke-width', String(shape.stroke || 16));
    });
    container.querySelectorAll('.vessel-shell').forEach((el) => {
      el.setAttribute('fill-rule', rule);
    });
    const detail = document.getElementById(isWins ? 'winsDetail' : 'lossesDetail');
    if (detail) detail.setAttribute('d', shape.detail || '');
    if (shape.scale) container.style.setProperty('--vessel-scale', String(shape.scale));
    else container.style.removeProperty('--vessel-scale');
    const parts = String(viewBox).trim().split(/\s+/).map(Number);
    const vbX = parts[0] || 0;
    const vbY = parts[1] || 0;
    const vbW = parts[2] || 512;
    const vbH = parts[3] || 512;
    const liquid = isWins ? dom.winsLiquid : dom.lossesLiquid;
    if (liquid) {
      const prev = liquidY.has(liquid) ? fillRatioFromY(liquid, readLiquidY(liquid)) : 0;
      setLiquidGeom(liquid, vbX, vbY, vbW, vbH);
      const body = liquid.querySelector('.liquid-body');
      if (body) {
        body.setAttribute('x', String(vbX - Math.round(vbW * 0.12)));
        body.setAttribute('width', String(Math.round(vbW * 1.28)));
        body.setAttribute('height', String(Math.round(vbH * 1.2)));
      }
      // Keep current fill ratio after viewBox change.
      writeLiquidY(liquid, ratioToY(prev, liquid));
      // Scale wave paths to current viewBox width.
      updateWaveSurface(liquid, Math.round(vbW * 1.28), isWins ? 0 : 1.3);
    }
  }

  function fillRatioFromY(liquidEl, y) {
    const g = getLiquidGeom(liquidEl);
    if (!g.h) return 0;
    return Math.max(0, Math.min(1, 1 - (y - g.minY) / g.h));
  }

  function applyVesselIcons() {
    applyVesselIcon('wins', vesselStyleId);
    applyVesselIcon('losses', vesselStyleId);
    fixSvgClipRefs();
  }

  const PREVIEW_FILL_W = 0.82;
  const PREVIEW_FILL_H = 0.7;
  const PREVIEW_MAX_ZOOM = 2.6;
  // Runs synchronously (no animation frame): browsers pause frames in covered,
  // minimised or busy windows and throttle embedded frames, which left the
  // preview unsized after a layout switch.
  function fitPreviewWidget() {
    if (!isPreview || !dom.widget) return;
    const root = document.documentElement;
    root.style.setProperty('--preview-zoom', '1');
    // Natural size after zoom reset
    const w = Math.max(1, dom.widget.offsetWidth);
    const h = Math.max(1, dom.widget.offsetHeight);
    // Fill most of the preview box whatever the layout: compact layouts are
    // scaled up, wide ones down. The margin leaves room for rank effects.
    const zx = (window.innerWidth * PREVIEW_FILL_W) / w;
    const zy = (window.innerHeight * PREVIEW_FILL_H) / h;
    const zoom = Math.min(PREVIEW_MAX_ZOOM, zx, zy);
    root.style.setProperty('--preview-zoom', String(Number.isFinite(zoom) && zoom > 0 ? zoom : 1));
  }

  function applySettings(next) {
    if (!next || typeof next !== 'object') return;
    const prev = settings && typeof settings === 'object' ? settings : {};
    // Merge so empty remote fields (WS hello/reconnect) don't wipe local rankFx etc.
    const merged = { ...prev, ...next };
    for (const key of ['rankFx', 'fillStyle', 'emptyEffect', 'appearEffect', 'animDirection', 'font']) {
      if (next[key] == null || next[key] === '') {
        merged[key] = prev[key] || merged[key];
      }
    }
    if (prev.__previewLocal && isPreview) merged.__previewLocal = true;
    settings = merged;

    animDurationIn = (merged.animDurationIn || 500) / 1000;
    animStayTime = (merged.animStayTime || 10000) / 1000;
    animDurationOut = (merged.animDurationOut || 500) / 1000;
    reappearDelayMs = merged.hiddenTime || 0;
    animDirection = merged.animDirection || 'left';
    alwaysVisible = !!merged.alwaysVisible;
    document.body.classList.toggle('always-visible', alwaysVisible);
    fillLimit = merged.fillLimit || 10;
    fillDurationMs = merged.fillDurationMs || 650;
    emptyDurationMs = merged.emptyDurationMs || 1600;
    emptyEffect = merged.emptyEffect || 'drain';
    numberAnimMs = merged.numberAnimMs || 500;
    vesselWave = merged.vesselWave !== false;
    idlePulse = merged.idlePulse !== false;
    appearEffect = merged.appearEffect || 'slide';
    fillStyle = merged.fillStyle || 'liquid';
    rankFx = merged.rankFx || prev.rankFx || 'classic';
    vesselStyleId = merged.vesselStyle || 'classic';
    layout = LAYOUTS.includes(merged.layout) ? merged.layout : 'detailed';
    applyCardFx(merged.cardFx);

    if (typeof window.loadGoogleFont === 'function') {
      window.loadGoogleFont(merged.font || 'Arial, sans-serif');
    }

    const winsColor = merged.winsColor || '#00ff00';
    const lossesColor = merged.lossesColor || '#ff0000';
    const iconColor = merged.iconColor || '#ffffff';
    const sepColor = merged.separatorColor || 'rgba(255,255,255,0.55)';

    document.documentElement.style.setProperty('--bg-color', merged.bgColor || 'rgba(0,0,0,0.7)');
    applySkinTone(merged.bgColor || 'rgba(0,0,0,0.7)');
    document.documentElement.style.setProperty('--wins-color', winsColor);
    document.documentElement.style.setProperty('--losses-color', lossesColor);
    document.documentElement.style.setProperty('--rank-text-color', merged.rankTextColor || '#ffffff');
    document.documentElement.style.setProperty('--icon-color', iconColor);
    document.documentElement.style.setProperty('--separator-color', sepColor);
    document.documentElement.style.setProperty('--font', merged.font || 'Arial, sans-serif');
    document.documentElement.style.setProperty('--font-size', `${merged.fontSize || 16}px`);
    document.documentElement.style.setProperty('--animation-duration-in', `${animDurationIn}s`);
    document.documentElement.style.setProperty('--animation-duration-out', `${animDurationOut}s`);
    document.documentElement.style.setProperty('--fill-duration', `${fillDurationMs}ms`);
    document.documentElement.style.setProperty('--empty-duration', `${emptyDurationMs}ms`);
    document.body.classList.toggle('vessel-wave', vesselWave && fillStyle !== 'solid');
    document.body.classList.toggle('no-idle-pulse', !idlePulse);
    document.body.dataset.rankFx = rankFx;
    if (dom.rankSection) dom.rankSection.dataset.rankFx = rankFx;
    if (dom.rankAnimationWrapper) dom.rankAnimationWrapper.dataset.rankFx = rankFx;
    applyLayout();
    applyAppearEffect();
    applyFillStyleClass();
    applyVesselIcons();
    setLiquidPaint(dom.winsLiquid, winsColor);
    setLiquidPaint(dom.lossesLiquid, lossesColor);
    if (dom.winsIconContainer) dom.winsIconContainer.style.setProperty('--pour-color', winsColor);
    if (dom.lossesIconContainer) dom.lossesIconContainer.style.setProperty('--pour-color', lossesColor);
    if (sepEl) sepEl.style.color = sepColor;
    dom.widget.style.background = merged.bgImage
      ? `url('${merged.bgImage}') center/cover no-repeat`
      : (merged.bgColor || 'rgba(0,0,0,0.7)');
    ensureWaveLoop();
    fitPreviewWidget();
  }

  function preferPreviewLocalSettings() {
    return isPreview && settings && settings.__previewLocal;
  }

  function viewFromSnap(snap) {
    if (snap?.view && typeof snap.view.wins === 'number') {
      return {
        wins: snap.view.wins || 0,
        losses: snap.view.losses || 0,
        rank: snap.view.rank || 0,
        mode: snap.view.mode || snap.state?.mode || 'classic',
        role: snap.view.role || snap.state?.role || 'tank',
        game: snap.view.game || snap.state?.game || 'overwatch',
        startRank: typeof snap.view.startRank === 'number' ? snap.view.startRank : (snap.view.rank || 0),
        recent: Array.isArray(snap.view.recent) ? snap.view.recent : [],
        roles: snap.state?.roles || null,
      };
    }
    const st = snap?.state || {};
    const game = st.game || 'overwatch';
    const mode = st.mode || 'classic';
    const role = st.role || 'tank';
    if (game === 'apex') {
      return { wins: st.wins || 0, losses: st.losses || 0, rank: st.rank || 0, mode: 'classic', role, game };
    }
    if (mode === 'roles_shared' || mode === 'roles_rotate') {
      const r = (st.roles && st.roles[role]) || {};
      return { wins: st.wins || 0, losses: st.losses || 0, rank: r.rank || 0, mode, role, game };
    }
    if (mode === 'roles_split') {
      const r = (st.roles && st.roles[role]) || {};
      return { wins: r.wins || 0, losses: r.losses || 0, rank: r.rank || 0, mode, role, game };
    }
    return { wins: st.wins || 0, losses: st.losses || 0, rank: st.rank || 0, mode: 'classic', role, game };
  }

  function updateRoleBadge(view) {
    const mode = view?.mode || 'classic';
    const role = view?.role || 'tank';
    const game = view?.game || 'overwatch';
    document.body.classList.remove('mode-classic', 'mode-roles_shared', 'mode-roles_split', 'mode-roles_rotate', 'game-overwatch', 'game-apex');
    document.body.classList.add(`mode-${mode}`);
    document.body.classList.add(`game-${game}`);
    renderLayoutExtras(); // role/mode/game switches don't always change any text
    if (!dom.roleBadge || !dom.roleIcon) return;
    if (game === 'apex' || mode === 'classic') {
      dom.roleBadge.hidden = true;
      return;
    }
    dom.roleBadge.hidden = false;
    const url = ROLE_ICONS[role] || ROLE_ICONS.tank;
    dom.roleIcon.style.setProperty('--role-icon-url', `url("${url}")`);
    dom.roleIcon.setAttribute('aria-label', role);
  }

  function applyRemoteStateOnly(snap) {
    if (!snap) return;
    const v = viewFromSnap(snap);
    currentState = { ...v };
    previousState = { ...currentState };
    updateRoleBadge(v);
    dom.wins.textContent = currentState.wins;
    dom.losses.textContent = currentState.losses;
    setLiquidLevel(dom.winsLiquid, currentState.wins, { instant: true });
    setLiquidLevel(dom.lossesLiquid, currentState.losses, { instant: true });
    const info = getRankInfo(currentState.rank);
    setRankImages(info.img);
    dom.rankValue.textContent = info.display;
  }

  function fillRatio(value) {
    if (value <= 0) return 0;
    const rem = value % fillLimit;
    // After a completed cycle (and empty animation), vessel stays empty until the next point.
    if (rem === 0) return 0;
    return rem / fillLimit;
  }

  const LIQUID_H = 512;
  const liquidY = new WeakMap();
  const liquidGeom = new WeakMap(); // { minY, h } from current vessel viewBox
  const liquidAnim = new WeakMap(); // { raf, resolve }
  const vesselGen = new WeakMap();

  function getLiquidGeom(liquidEl) {
    return liquidGeom.get(liquidEl) || { minY: 0, h: LIQUID_H };
  }

  function emptyLiquidY(liquidEl) {
    const g = getLiquidGeom(liquidEl);
    return g.minY + g.h;
  }

  function setLiquidGeom(liquidEl, vbX, vbY, vbW, vbH) {
    if (!liquidEl) return;
    liquidGeom.set(liquidEl, {
      minY: vbY || 0,
      h: vbH > 0 ? vbH : LIQUID_H,
      minX: vbX || 0,
      w: vbW > 0 ? vbW : LIQUID_H,
    });
  }

  function readLiquidY(liquidEl) {
    if (liquidY.has(liquidEl)) return liquidY.get(liquidEl);
    const g = getLiquidGeom(liquidEl);
    const m = /translate\(\s*[^,\s)]+[,\s]+([^)]+)\)/.exec(liquidEl.getAttribute('transform') || '');
    return m ? parseFloat(m[1]) : (g.minY + g.h);
  }

  function writeLiquidY(liquidEl, y) {
    const g = getLiquidGeom(liquidEl);
    const fullY = g.minY;
    const emptyY = g.minY + g.h;
    const clamped = Math.max(fullY, Math.min(emptyY, y));
    liquidY.set(liquidEl, clamped);
    liquidEl.setAttribute('transform', `translate(0, ${clamped})`);
  }

  function cancelLiquidAnim(liquidEl) {
    const st = liquidAnim.get(liquidEl);
    if (!st) return;
    if (st.raf) cancelAnimationFrame(st.raf);
    liquidAnim.delete(liquidEl);
    // Always resolve so awaiters (pour/empty) never hang forever.
    if (typeof st.resolve === 'function') st.resolve();
  }

  function bumpVesselGen(iconContainer) {
    if (!iconContainer) return 0;
    const n = (vesselGen.get(iconContainer) || 0) + 1;
    vesselGen.set(iconContainer, n);
    return n;
  }

  function vesselAlive(iconContainer, gen) {
    return !!iconContainer && vesselGen.get(iconContainer) === gen;
  }

  const VESSEL_FX_CLASSES = [
    'is-emptying', 'pouring', 'wave-active',
    'vessel-drain-active', 'vessel-splash-active', 'vessel-burst-active',
    'vessel-pour-active', 'vessel-fade-active',
  ];

  function resetVesselVisual(iconContainer, liquidEl) {
    bumpVesselGen(iconContainer);
    if (iconContainer) {
      if (iconContainer._pourBlobTimer) {
        clearTimeout(iconContainer._pourBlobTimer);
        iconContainer._pourBlobTimer = 0;
      }
      iconContainer.classList.remove(...VESSEL_FX_CLASSES);
      const stream = iconContainer.querySelector('.pour-stream');
      if (stream) stream.classList.remove('active');
    }
    if (liquidEl) {
      cancelLiquidAnim(liquidEl);
      liquidEl.classList.remove('is-filling');
      liquidEl.style.transition = '';
      liquidEl.style.opacity = '';
      const meniscus = liquidEl.querySelector('.liquid-meniscus');
      const foam = liquidEl.querySelector('.liquid-foam');
      if (meniscus) { meniscus.style.opacity = ''; meniscus.style.display = ''; }
      if (foam) { foam.style.opacity = ''; foam.style.display = ''; }
    }
  }

  function easeOutCubic(t) {
    return 1 - Math.pow(1 - t, 3);
  }

  function easeInOutCubic(t) {
    return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
  }

  function animateLiquidY(liquidEl, targetY, durationMs, easeFn = easeOutCubic) {
    cancelLiquidAnim(liquidEl);
    const g = getLiquidGeom(liquidEl);
    const from = readLiquidY(liquidEl);
    const to = Math.max(g.minY, Math.min(g.minY + g.h, targetY));
    if (durationMs <= 0 || Math.abs(from - to) < 0.2) {
      writeLiquidY(liquidEl, to);
      return Promise.resolve();
    }
    const start = performance.now();
    return new Promise((resolve) => {
      const st = { raf: 0, resolve };
      function frame(now) {
        // Cancelled — resolve already called from cancelLiquidAnim.
        if (liquidAnim.get(liquidEl) !== st) return;
        const t = Math.min(1, (now - start) / durationMs);
        writeLiquidY(liquidEl, from + (to - from) * easeFn(t));
        if (t < 1) {
          st.raf = requestAnimationFrame(frame);
        } else {
          liquidAnim.delete(liquidEl);
          resolve();
        }
      }
      st.raf = requestAnimationFrame(frame);
      liquidAnim.set(liquidEl, st);
    });
  }

  function ratioToY(ratio, liquidEl) {
    const g = getLiquidGeom(liquidEl);
    return g.minY + g.h * (1 - Math.max(0, Math.min(1, ratio)));
  }

  function setLiquidLevel(liquidEl, value, opts = {}) {
    if (!liquidEl) return Promise.resolve();
    const { instant = false, full = false, duration = fillDurationMs } = opts;
    const ratio = full ? 1 : fillRatio(value);
    const y = ratioToY(ratio, liquidEl);
    liquidEl.classList.toggle('is-filling', !instant && ratio > 0);
    if (instant) {
      cancelLiquidAnim(liquidEl);
      writeLiquidY(liquidEl, y);
      return Promise.resolve();
    }
    return animateLiquidY(liquidEl, y, duration);
  }

  function animateNumber(element, from, to) {
    if (from === to) {
      element.dataset.animToken = String(Number(element.dataset.animToken || 0) + 1);
      element.textContent = to;
      return;
    }
    const isReset = to < from;
    const duration = isReset ? Math.max(numberAnimMs, 750) : numberAnimMs;
    // A newer count on the same number cancels this one (steps and safety timer).
    const token = (element.dataset.animToken = String(Number(element.dataset.animToken || 0) + 1));
    const current = () => element.dataset.animToken === token;
    let start = null;
    element.classList.add('value-change');
    element.textContent = from;
    function step(timestamp) {
      if (!current()) return;
      if (!start) start = timestamp;
      const progress = Math.min((timestamp - start) / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      if (isReset) {
        element.textContent = Math.max(0, Math.round(from - eased * (from - to)));
      } else {
        element.textContent = Math.round(from + eased * (to - from));
      }
      if (progress < 1) requestAnimationFrame(step);
      else finish();
    }
    // Always land on the final number, even if animation frames are paused.
    let done = false;
    const finish = () => {
      if (done || !current()) return;
      done = true;
      element.textContent = to;
      setTimeout(() => element.classList.remove('value-change'), 500);
    };
    requestAnimationFrame(step);
    setTimeout(finish, duration + 150);
  }

  function emptyEffectClass() {
    if (emptyEffect === 'splash') return 'vessel-splash-active';
    if (emptyEffect === 'burst') return 'vessel-burst-active';
    if (emptyEffect === 'pour') return 'vessel-pour-active';
    if (emptyEffect === 'fade') return 'vessel-fade-active';
    return 'vessel-drain-active';
  }

  function wavePath(width, t, amp, phase) {
    const y0 = 0;
    const left = -80;
    const right = left + width;
    let d = `M${left} ${y0}`;
    const steps = 16;
    for (let i = 1; i <= steps; i++) {
      const x = left + (right - left) * (i / steps);
      const y = y0 + Math.sin(t * 2.2 + i * 0.55 + phase) * amp
        + Math.sin(t * 3.1 + i * 1.1 + phase * 1.7) * (amp * 0.35);
      d += ` L${x.toFixed(1)} ${y.toFixed(1)}`;
    }
    d += ` V8 H${left} Z`;
    return d;
  }

  function foamPath(width, t, amp, phase) {
    const y0 = 0;
    const left = -80;
    const right = left + width;
    let d = `M${left} ${y0}`;
    const steps = 14;
    for (let i = 1; i <= steps; i++) {
      const x = left + (right - left) * (i / steps);
      const y = y0 + Math.sin(t * 2.6 + i * 0.7 + phase) * (amp * 0.7);
      d += ` L${x.toFixed(1)} ${y.toFixed(1)}`;
    }
    d += ` V6 H${left} Z`;
    return d;
  }


  function flattenLiquidSurface(liquidEl) {
    if (!liquidEl) return;
    const meniscus = liquidEl.querySelector('.liquid-meniscus');
    const foam = liquidEl.querySelector('.liquid-foam');
    if (meniscus) {
      meniscus.setAttribute('d', 'M-80 0 H800 V0 H-80 Z');
      meniscus.style.display = 'none';
      meniscus.style.opacity = '0';
    }
    if (foam) {
      foam.setAttribute('d', 'M-80 0 H800 V0 H-80 Z');
      foam.style.display = 'none';
      foam.style.opacity = '0';
    }
  }

  function updateWaveSurface(liquidEl, width, phase) {
    if (!liquidEl) return;
    const container = liquidEl.closest('.icon-fill-container');
    if (container?.classList.contains('is-emptying')) {
      flattenLiquidSurface(liquidEl);
      return;
    }
    const amp = (fillStyle === 'solid') ? 0 : (vesselWave ? 11 : 3);
    const meniscus = liquidEl.querySelector('.liquid-meniscus');
    const foam = liquidEl.querySelector('.liquid-foam');
    if (meniscus) {
      meniscus.style.display = '';
      meniscus.style.opacity = '';
      meniscus.setAttribute('d', wavePath(width, waveTime, amp, phase));
    }
    if (foam) {
      foam.style.display = '';
      foam.style.opacity = (fillStyle === 'liquid' || fillStyle === 'bubble') && vesselWave ? '1' : '0';
      foam.setAttribute('d', foamPath(width, waveTime + 0.4, amp, phase + 1));
    }
  }

  function ensureWaveLoop() {
    if (waveRaf) return;
    const tick = () => {
      waveTime += 0.055;
      const winsActive = dom.winsLiquid?.classList.contains('is-filling')
        || dom.winsIconContainer?.classList.contains('wave-active')
        || readLiquidY(dom.winsLiquid) < emptyLiquidY(dom.winsLiquid) - 0.5;
      const lossesActive = dom.lossesLiquid?.classList.contains('is-filling')
        || dom.lossesIconContainer?.classList.contains('wave-active')
        || readLiquidY(dom.lossesLiquid) < emptyLiquidY(dom.lossesLiquid) - 0.5;
      const winsW = getLiquidGeom(dom.winsLiquid).w || 740;
      const lossesW = getLiquidGeom(dom.lossesLiquid).w || 680;
      if (winsActive) updateWaveSurface(dom.winsLiquid, Math.round(winsW * 1.28), 0);
      if (lossesActive) updateWaveSurface(dom.lossesLiquid, Math.round(lossesW * 1.28), 1.3);
      waveRaf = requestAnimationFrame(tick);
    };
    waveRaf = requestAnimationFrame(tick);
  }

  function spawnBubbles(iconContainer, liquidEl) {
    if (fillStyle !== 'bubble' || !iconContainer) return;
    const fx = iconContainer.querySelector('.vessel-fx');
    if (!fx) return;
    const y = liquidEl ? readLiquidY(liquidEl) : emptyLiquidY(liquidEl);
    const fill = Math.max(0.05, Math.min(1, fillRatioFromY(liquidEl, y) || 0.05));
    const surfaceBottom = Math.max(14, Math.min(78, fill * 100 - 3));
    const count = fill < 0.15 ? 1 : 2;
    for (let i = 0; i < count; i++) {
      const b = document.createElement('span');
      b.className = 'bubble';
      const size = 3.5 + Math.random() * 2.5;
      b.style.left = `${28 + Math.random() * 40}%`;
      b.style.bottom = `${surfaceBottom + (Math.random() * 4 - 1)}%`;
      b.style.width = `${size}px`;
      b.style.height = `${size}px`;
      b.style.animationDuration = `${0.7 + Math.random() * 0.45}s`;
      b.style.animationDelay = `${Math.random() * 0.12}s`;
      fx.appendChild(b);
      setTimeout(() => b.remove(), 1400);
    }
  }

  function spawnPourBlobBurst(host, color, count = 10) {
    if (!host) return;
    for (let i = 0; i < count; i++) {
      const b = document.createElement('span');
      const roll = Math.random();
      b.className = 'pour-blob' + (roll > 0.72 ? ' fat' : roll < 0.28 ? ' thin' : '');
      b.style.setProperty('--pour-color', color);
      const bx = -(22 + Math.random() * 34);
      const by = 28 + Math.random() * 38;
      b.style.setProperty('--bx', `${bx}px`);
      b.style.setProperty('--by', `${by}px`);
      b.style.setProperty('--rot', `${18 + Math.random() * 32}deg`);
      b.style.setProperty('--delay', `${Math.random() * 0.12}s`);
      b.style.setProperty('--dur', `${0.5 + Math.random() * 0.45}s`);
      b.style.left = `${-2 + Math.random() * 8}px`;
      b.style.top = `${8 + Math.random() * 14}%`;
      host.appendChild(b);
      setTimeout(() => b.remove(), 1200);
    }
  }

  function spawnPourDroplets(iconContainer) {
    const host = iconContainer.querySelector('.pour-stream') || iconContainer.querySelector('.vessel-fx');
    if (!host) return;
    const color = getComputedStyle(iconContainer).getPropertyValue('--pour-color').trim() || '#39ff14';
    spawnPourBlobBurst(host, color, 12);
    for (let i = 0; i < 8; i++) {
      const d = document.createElement('span');
      d.className = 'pour-drop';
      d.style.setProperty('--pour-color', color);
      d.style.setProperty('--bx', `${-(30 + Math.random() * 28)}px`);
      d.style.setProperty('--by', `${36 + Math.random() * 30}px`);
      d.style.setProperty('--rot', `${22 + Math.random() * 40}deg`);
      d.style.setProperty('--delay', `${0.05 + Math.random() * 0.35}s`);
      d.style.setProperty('--dur', `${0.55 + Math.random() * 0.4}s`);
      d.style.left = `${Math.random() * 10}px`;
      d.style.top = `${10 + Math.random() * 20}%`;
      host.appendChild(d);
      setTimeout(() => d.remove(), 1400);
    }
  }

  async function pourOut(liquidEl, iconContainer, gen) {
    const stream = iconContainer.querySelector('.pour-stream');
    iconContainer.classList.add('pouring', 'is-emptying');
    flattenLiquidSurface(liquidEl);
    const color = getComputedStyle(iconContainer).getPropertyValue('--pour-color').trim() || '#39ff14';
    if (iconContainer._pourBlobTimer) clearTimeout(iconContainer._pourBlobTimer);
    if (stream) {
      stream.style.setProperty('--pour-color', color);
      stream.classList.remove('active');
      void stream.offsetWidth;
      stream.classList.add('active');
      const pulse = () => {
        if (!vesselAlive(iconContainer, gen) || !iconContainer.classList.contains('pouring')) return;
        spawnPourBlobBurst(stream, color, 7 + Math.floor(Math.random() * 5));
        iconContainer._pourBlobTimer = setTimeout(pulse, 110 + Math.random() * 90);
      };
      pulse();
    }
    spawnPourDroplets(iconContainer);
    await new Promise((r) => setTimeout(r, Math.min(180, emptyDurationMs * 0.1)));
    if (!vesselAlive(iconContainer, gen)) return;
    await animateLiquidY(liquidEl, emptyLiquidY(liquidEl), emptyDurationMs * 0.9, easeInOutCubic);
    if (iconContainer._pourBlobTimer) {
      clearTimeout(iconContainer._pourBlobTimer);
      iconContainer._pourBlobTimer = 0;
    }
    if (stream) stream.classList.remove('active');
    iconContainer.classList.remove('pouring');
  }

  function spawnSplashDroplets(iconContainer) {
    const fx = iconContainer.querySelector('.vessel-fx');
    if (!fx) return;
    const color = getComputedStyle(iconContainer).getPropertyValue('--pour-color').trim() || '#fff';
    for (let i = 0; i < 8; i++) {
      const d = document.createElement('span');
      d.className = 'splash-drop';
      d.style.background = color;
      const ang = -60 + Math.random() * 120;
      d.style.setProperty('--dx', `${Math.sin((ang * Math.PI) / 180) * (10 + Math.random() * 14)}px`);
      d.style.setProperty('--dy', `${-8 - Math.random() * 16}px`);
      d.style.left = '45%';
      d.style.top = '40%';
      fx.appendChild(d);
      setTimeout(() => d.remove(), 900);
    }
  }

  async function animateFill(liquidEl, iconContainer, from, to) {
    if (!liquidEl || !iconContainer) return;
    const gen = bumpVesselGen(iconContainer);
    if (from === to) {
      await setLiquidLevel(liquidEl, to, { instant: true });
      return;
    }

    await setLiquidLevel(liquidEl, from, { instant: true });
    if (!vesselAlive(iconContainer, gen)) return;
    const rising = to > from;
    iconContainer.classList.toggle('wave-active', vesselWave && rising && fillStyle !== 'solid');
    liquidEl.classList.add('is-filling');
    if (rising) spawnBubbles(iconContainer, liquidEl);

    const crossedThreshold = rising && to > 0 && to % fillLimit === 0;
    if (!crossedThreshold) {
      const anim = setLiquidLevel(liquidEl, to, { instant: false, duration: fillDurationMs });
      const bubbleTimer = rising && fillStyle === 'bubble'
        ? setInterval(() => {
          if (!vesselAlive(iconContainer, gen)) { clearInterval(bubbleTimer); return; }
          spawnBubbles(iconContainer, liquidEl);
        }, 280)
        : null;
      await anim;
      if (bubbleTimer) clearInterval(bubbleTimer);
      if (!vesselAlive(iconContainer, gen)) return;
      liquidEl.classList.remove('is-filling');
      setTimeout(() => {
        if (vesselAlive(iconContainer, gen)) iconContainer.classList.remove('wave-active');
      }, 180);
      return;
    }

    await setLiquidLevel(liquidEl, fillLimit, { instant: false, full: true, duration: fillDurationMs });
    if (!vesselAlive(iconContainer, gen)) return;
    const cls = emptyEffectClass();
    iconContainer.classList.add(cls, 'is-emptying');
    iconContainer.classList.remove('wave-active');
    flattenLiquidSurface(liquidEl);
    await new Promise((r) => setTimeout(r, Math.max(240, Math.floor(fillDurationMs * 0.35))));
    if (!vesselAlive(iconContainer, gen)) return;

    if (emptyEffect === 'fade') {
      liquidEl.style.opacity = '1';
      liquidEl.style.transition = `opacity ${emptyDurationMs}ms ease`;
      liquidEl.style.opacity = '0';
      await animateLiquidY(liquidEl, emptyLiquidY(liquidEl), emptyDurationMs, easeInOutCubic);
      if (vesselAlive(iconContainer, gen)) {
        liquidEl.style.transition = '';
        liquidEl.style.opacity = '1';
      }
    } else if (emptyEffect === 'pour') {
      await pourOut(liquidEl, iconContainer, gen);
    } else if (emptyEffect === 'splash') {
      spawnSplashDroplets(iconContainer);
      await animateLiquidY(liquidEl, emptyLiquidY(liquidEl), emptyDurationMs, easeInOutCubic);
    } else if (emptyEffect === 'burst') {
      spawnSplashDroplets(iconContainer);
      await animateLiquidY(liquidEl, emptyLiquidY(liquidEl), emptyDurationMs * 0.85, easeOutCubic);
    } else {
      await animateLiquidY(liquidEl, emptyLiquidY(liquidEl), emptyDurationMs, easeInOutCubic);
    }

    if (!vesselAlive(iconContainer, gen)) return;
    iconContainer.classList.remove(cls, 'wave-active', 'is-emptying', 'pouring');
    liquidEl.classList.remove('is-filling');
    const meniscus = liquidEl.querySelector('.liquid-meniscus');
    const foam = liquidEl.querySelector('.liquid-foam');
    if (meniscus) { meniscus.style.opacity = ''; meniscus.style.display = ''; }
    if (foam) { foam.style.opacity = ''; foam.style.display = ''; }
    writeLiquidY(liquidEl, emptyLiquidY(liquidEl));
  }

  function clearWidgetFx() {
    if (!dom.widget) return;
    dom.widget.classList.remove(
      'widget-fx-active', 'fx-up', 'fx-down',
      'widget-fx-classic', 'widget-fx-blaze', 'widget-fx-melt', 'widget-fx-frost',
      'widget-fx-neon', 'widget-fx-divine', 'widget-fx-arcane', 'widget-fx-toxic', 'widget-fx-paper',
      'widget-fx-vortex', 'widget-fx-shatter', 'widget-fx-smoke', 'widget-fx-lightning', 'widget-fx-ripple',
      'widget-fx-hologram', 'widget-fx-pixel', 'widget-fx-chrome', 'widget-fx-phantom', 'widget-fx-sonic', 'widget-fx-bloom', 'widget-fx-ink', 'widget-fx-prism'
    );
    if (dom.widgetFxLayer) dom.widgetFxLayer.innerHTML = '';
  }

  function setRankImages(src, underSrc) {
    if (dom.rankImage) dom.rankImage.src = src;
    if (dom.rankImageUnder) dom.rankImageUnder.src = underSrc || src;
  }

  function applyRankVisualInstant(absoluteIndex) {
    const info = getRankInfo(absoluteIndex);
    setRankImages(info.img);
    dom.rankValue.textContent = info.display;
    if (dom.rankImage) {
      dom.rankImage.style.opacity = '1';
      dom.rankImage.style.visibility = 'visible';
      dom.rankImage.style.filter = '';
      dom.rankImage.style.transform = '';
      dom.rankImage.style.clipPath = '';
      dom.rankImage.style.animation = 'none';
      void dom.rankImage.offsetWidth;
      dom.rankImage.style.animation = '';
    }
  }

  let rankAnimTimers = [];
  function clearRankAnimTimers() {
    for (const id of rankAnimTimers) clearTimeout(id);
    rankAnimTimers = [];
  }

  function abortRankTransition(finalRankIndex) {
    clearRankAnimTimers();
    clearWidgetFx();
    clearRankFx();
    isRankTransitionAnimating = false;
    if (typeof finalRankIndex === 'number') {
      currentState = { ...currentState, rank: finalRankIndex };
      previousState = { ...previousState, rank: finalRankIndex };
      applyRankVisualInstant(finalRankIndex);
    }
  }

  function clearRankFx() {
    if (dom.rankFxLayer) dom.rankFxLayer.innerHTML = '';
    if (dom.rankAnimationWrapper) {
      // Strip any leftover theme-* class from prior FX
      for (const cls of [...dom.rankAnimationWrapper.classList]) {
        if (cls.startsWith('theme-')) dom.rankAnimationWrapper.classList.remove(cls);
      }
      dom.rankAnimationWrapper.classList.remove(
        'rank-up-division', 'rank-down-division', 'rank-level-change', 'rank-idle-pulse', 'rank-idle-burst',
        'rank-up', 'rank-down', 'rank-level', 'dual-reveal'
      );
    }
    if (dom.rankSection) {
      dom.rankSection.classList.remove(
        'rank-up', 'rank-down', 'rank-level', 'rank-div-up', 'rank-div-down', 'rank-idle-pulse'
      );
    }
    if (dom.rankImageUnder) {
      dom.rankImageUnder.style.opacity = '0';
      dom.rankImageUnder.style.visibility = 'hidden';
      dom.rankImageUnder.style.filter = '';
      dom.rankImageUnder.style.transform = '';
      dom.rankImageUnder.style.animation = '';
    }
    if (dom.rankImage) {
      dom.rankImage.style.opacity = '1';
      dom.rankImage.style.filter = '';
      dom.rankImage.style.transform = '';
      dom.rankImage.style.clipPath = '';
      dom.rankImage.style.animation = '';
    }
  }

  function playWidgetDivisionFx(isUp) {
    if (!dom.widget) return;
    clearWidgetFx();
    const fx = rankFx || 'classic';
    dom.widget.classList.add('widget-fx-active', `widget-fx-${fx}`, isUp ? 'fx-up' : 'fx-down');
    if (dom.widgetFxLayer) {
      const edge = document.createElement('div');
      edge.className = `widget-fx-edge fx-${fx} ${isUp ? 'up' : 'down'}`;
      dom.widgetFxLayer.appendChild(edge);
    }
    setTimeout(clearWidgetFx, 2200);
  }

  function spawnRankParticles(isUp, isDivision) {
    const layer = dom.rankFxLayer;
    if (!layer) return;
    layer.innerHTML = '';
    const fx = rankFx || 'classic';
    const aura = document.createElement('span');
    aura.className = `rank-aura ${isUp ? 'up' : 'down'} ${isDivision ? 'division' : 'level'} fx-${fx}`;
    layer.appendChild(aura);

    const n = isDivision ? 14 : 6;
    if (fx === 'blaze') {
      const heat = document.createElement('span');
      heat.className = `rank-heat-glow ${isUp ? 'up' : 'down'}`;
      layer.appendChild(heat);
      setTimeout(() => heat.remove(), 2600);

      const flameCount = isDivision ? (isUp ? 8 : 14) : 7;
      for (let i = 0; i < flameCount; i++) {
        const f = document.createElement('span');
        const size = i % 3 === 0 ? 'lg' : (i % 3 === 1 ? 'md' : 'sm');
        f.className = `rank-flame flame-${size}`;
        f.style.left = `${12 + Math.random() * 76}%`;
        f.style.setProperty('--flick', `${-8 + Math.random() * 16}deg`);
        f.style.setProperty('--rise', `${18 + Math.random() * 22}px`);
        f.style.animationDelay = `${(isUp && isDivision ? 0.15 : 0) + Math.random() * 0.55}s`;
        f.style.animationDuration = `${0.85 + Math.random() * 0.45}s`;
        layer.appendChild(f);
        setTimeout(() => f.remove(), 2200);
      }

      const emberCount = isDivision ? 16 : 8;
      for (let i = 0; i < emberCount; i++) {
        const e = document.createElement('span');
        e.className = i % 4 === 0 ? 'rank-ember hot' : 'rank-ember';
        e.style.left = `${18 + Math.random() * 64}%`;
        e.style.bottom = `${6 + Math.random() * 36}%`;
        e.style.setProperty('--ex', `${-22 + Math.random() * 44}px`);
        e.style.setProperty('--ey', `${-28 - Math.random() * 36}px`);
        e.style.animationDelay = `${0.1 + Math.random() * 0.7}s`;
        e.style.animationDuration = `${0.9 + Math.random() * 0.6}s`;
        layer.appendChild(e);
        setTimeout(() => e.remove(), 2400);
      }

      const smokeCount = isDivision ? 7 : 3;
      for (let i = 0; i < smokeCount; i++) {
        const s = document.createElement('span');
        s.className = 'rank-smoke';
        s.style.left = `${20 + Math.random() * 60}%`;
        s.style.bottom = `${8 + Math.random() * 24}%`;
        s.style.setProperty('--sx', `${-16 + Math.random() * 32}px`);
        s.style.setProperty('--ss', `${0.7 + Math.random() * 0.9}`);
        s.style.animationDelay = `${0.2 + Math.random() * 0.6}s`;
        layer.appendChild(s);
        setTimeout(() => s.remove(), 2600);
      }

      if (isUp && isDivision) {
        for (let i = 0; i < 12; i++) {
          const a = document.createElement('span');
          a.className = 'rank-ash';
          a.style.left = `${18 + Math.random() * 64}%`;
          a.style.top = `${18 + Math.random() * 55}%`;
          a.style.setProperty('--ax', `${-18 + Math.random() * 36}px`);
          a.style.setProperty('--ay', `${-22 - Math.random() * 28}px`);
          a.style.animationDelay = `${0.28 + Math.random() * 0.55}s`;
          layer.appendChild(a);
          setTimeout(() => a.remove(), 2200);
        }
        for (let i = 0; i < 5; i++) {
          const c = document.createElement('span');
          c.className = 'rank-cinder';
          c.style.left = `${25 + Math.random() * 50}%`;
          c.style.top = `${30 + Math.random() * 40}%`;
          c.style.setProperty('--cx', `${-14 + Math.random() * 28}px`);
          c.style.setProperty('--cy', `${-10 - Math.random() * 20}px`);
          c.style.setProperty('--crot', `${-40 + Math.random() * 80}deg`);
          c.style.animationDelay = `${0.4 + Math.random() * 0.4}s`;
          layer.appendChild(c);
          setTimeout(() => c.remove(), 2300);
        }
      }
    } else if (fx === 'melt') {
      if (isUp && isDivision) {
        const press = document.createElement('span');
        press.className = 'rank-stamp-press';
        layer.appendChild(press);
        setTimeout(() => press.remove(), 2600);
        const flash = document.createElement('span');
        flash.className = 'rank-stamp-flash';
        layer.appendChild(flash);
        setTimeout(() => flash.remove(), 1800);
        const shock = document.createElement('span');
        shock.className = 'rank-stamp-shock';
        layer.appendChild(shock);
        setTimeout(() => shock.remove(), 1800);
        for (let i = 0; i < 14; i++) {
          const sp = document.createElement('span');
          sp.className = 'rank-spark';
          sp.style.left = `${30 + Math.random() * 40}%`;
          sp.style.top = `${48 + Math.random() * 18}%`;
          const side = Math.random() > 0.5 ? 1 : -1;
          sp.style.setProperty('--sx', `${side * (12 + Math.random() * 28)}px`);
          sp.style.setProperty('--sy', `${-6 - Math.random() * 18}px`);
          sp.style.animationDelay = `${0.95 + Math.random() * 0.12}s`;
          layer.appendChild(sp);
          setTimeout(() => sp.remove(), 1900);
        }
        for (let i = 0; i < 6; i++) {
          const s = document.createElement('span');
          s.className = 'rank-slag';
          s.style.left = `${28 + Math.random() * 44}%`;
          s.style.top = `${40 + Math.random() * 30}%`;
          s.style.setProperty('--sx', `${-18 + Math.random() * 36}px`);
          s.style.setProperty('--sy', `${6 + Math.random() * 16}px`);
          s.style.animationDelay = `${1.0 + Math.random() * 0.2}s`;
          layer.appendChild(s);
          setTimeout(() => s.remove(), 1900);
        }
      } else {
        for (let i = 0; i < (isDivision ? 18 : 9); i++) {
          const d = document.createElement('span');
          d.className = 'rank-drip';
          d.style.left = `${16 + Math.random() * 68}%`;
          d.style.top = `${28 + Math.random() * 42}%`;
          d.style.animationDelay = `${0.2 + Math.random() * 0.55}s`;
          d.style.height = `${11 + Math.random() * 12}px`;
          layer.appendChild(d);
          setTimeout(() => d.remove(), 1700);
        }
        for (let i = 0; i < 9; i++) {
          const s = document.createElement('span');
          s.className = 'rank-slag';
          s.style.left = `${22 + Math.random() * 56}%`;
          s.style.top = `${45 + Math.random() * 30}%`;
          s.style.setProperty('--sx', `${-16 + Math.random() * 32}px`);
          s.style.setProperty('--sy', `${12 + Math.random() * 20}px`);
          s.style.animationDelay = `${0.3 + Math.random() * 0.4}s`;
          layer.appendChild(s);
          setTimeout(() => s.remove(), 1500);
        }
      }
    } else if (fx === 'frost') {
      if (isUp && isDivision) {
        const shell = document.createElement('span');
        shell.className = 'rank-ice-shell';
        const maskSrc = (dom.rankImage && dom.rankImage.currentSrc) || (dom.rankImage && dom.rankImage.src) || '';
        if (maskSrc) {
          const url = `url("${maskSrc.replace(/\\/g, '/').replace(/"/g, '\\"')}")`;
          shell.style.setProperty('--rank-mask', url);
        }
        layer.appendChild(shell);
        setTimeout(() => shell.remove(), 1700);
      }
      for (let i = 0; i < (isDivision ? 16 : 8); i++) {
        const s = document.createElement('span');
        s.className = 'rank-frost-shard';
        s.style.left = `${12 + Math.random() * 76}%`;
        s.style.top = `${12 + Math.random() * 64}%`;
        const ang = Math.random() * Math.PI * 2;
        const dist = 16 + Math.random() * 28;
        s.style.setProperty('--sx', `${Math.cos(ang) * dist}px`);
        s.style.setProperty('--sy', `${Math.sin(ang) * dist}px`);
        s.style.setProperty('--rot', `${-50 + Math.random() * 100}deg`);
        s.style.animationDelay = `${0.42 + Math.random() * 0.35}s`;
        layer.appendChild(s);
        setTimeout(() => s.remove(), 1700);
      }
      for (let i = 0; i < 10; i++) {
        const sn = document.createElement('span');
        sn.className = 'rank-snow';
        sn.style.left = `${15 + Math.random() * 70}%`;
        sn.style.top = `${10 + Math.random() * 40}%`;
        sn.style.setProperty('--sx', `${-8 + Math.random() * 16}px`);
        sn.style.setProperty('--sy', `${14 + Math.random() * 22}px`);
        sn.style.animationDelay = `${Math.random() * 0.5}s`;
        layer.appendChild(sn);
        setTimeout(() => sn.remove(), 1500);
      }
    } else if (fx === 'paper' && isDivision) {
      const shade = document.createElement('span');
      shade.className = 'rank-page-shade';
      layer.appendChild(shade);
      setTimeout(() => shade.remove(), 1800);
    } else if (fx === 'vortex' && isDivision) {
      for (let i = 0; i < 3; i++) {
        const ring = document.createElement('span');
        ring.className = 'rank-vortex-ring';
        ring.style.animationDelay = `${i * 0.12}s`;
        layer.appendChild(ring);
        setTimeout(() => ring.remove(), 1800);
      }
    } else if (fx === 'shatter' && isDivision) {
      for (let i = 0; i < 14; i++) {
        const s = document.createElement('span');
        s.className = 'rank-glass-shard';
        s.style.left = `${15 + Math.random() * 70}%`;
        s.style.top = `${15 + Math.random() * 65}%`;
        const ang = Math.random() * Math.PI * 2;
        const dist = 18 + Math.random() * 30;
        s.style.setProperty('--sx', `${Math.cos(ang) * dist}px`);
        s.style.setProperty('--sy', `${Math.sin(ang) * dist}px`);
        s.style.setProperty('--rot', `${-60 + Math.random() * 120}deg`);
        s.style.animationDelay = `${0.35 + Math.random() * 0.25}s`;
        layer.appendChild(s);
        setTimeout(() => s.remove(), 1600);
      }
    } else if (fx === 'smoke' && isDivision) {
      for (let i = 0; i < 12; i++) {
        const p = document.createElement('span');
        p.className = 'rank-smoke-puff';
        p.style.left = `${20 + Math.random() * 60}%`;
        p.style.top = `${25 + Math.random() * 50}%`;
        p.style.setProperty('--sx', `${-14 + Math.random() * 28}px`);
        p.style.setProperty('--sy', `${-20 - Math.random() * 18}px`);
        p.style.animationDelay = `${Math.random() * 0.45}s`;
        layer.appendChild(p);
        setTimeout(() => p.remove(), 1600);
      }
    } else if (fx === 'lightning' && isDivision) {
      for (let i = 0; i < 5; i++) {
        const b = document.createElement('span');
        b.className = 'rank-bolt';
        b.style.left = `${20 + Math.random() * 60}%`;
        b.style.top = `${5 + Math.random() * 20}%`;
        b.style.setProperty('--rot', `${-25 + Math.random() * 50}deg`);
        b.style.animationDelay = `${i * 0.05}s`;
        layer.appendChild(b);
        setTimeout(() => b.remove(), 1200);
      }
    } else if (fx === 'ripple' && isDivision) {
      for (let i = 0; i < 3; i++) {
        const r = document.createElement('span');
        r.className = 'rank-ripple';
        r.style.animationDelay = `${0.15 + i * 0.18}s`;
        layer.appendChild(r);
        setTimeout(() => r.remove(), 1600);
      }
    } else if (fx === 'hologram' && isDivision) {
      const scan = document.createElement('span');
      scan.className = 'rank-holo-scan';
      layer.appendChild(scan);
      setTimeout(() => scan.remove(), 1800);
    } else if (fx === 'pixel' && isDivision) {
      const colors = ['#39ff14', '#00e5ff', '#ff6bcb', '#ffe566', '#ffffff'];
      for (let i = 0; i < 16; i++) {
        const p = document.createElement('span');
        p.className = 'rank-pixel';
        p.style.left = `${15 + Math.random() * 70}%`;
        p.style.top = `${15 + Math.random() * 70}%`;
        p.style.setProperty('--pxc', colors[i % colors.length]);
        p.style.setProperty('--sx', `${-16 + Math.random() * 32}px`);
        p.style.setProperty('--sy', `${-16 + Math.random() * 32}px`);
        p.style.animationDelay = `${Math.random() * 0.35}s`;
        layer.appendChild(p);
        setTimeout(() => p.remove(), 1400);
      }
    } else if (fx === 'chrome' && isDivision) {
      const g = document.createElement('span');
      g.className = 'rank-chrome-gleam';
      layer.appendChild(g);
      setTimeout(() => g.remove(), 1800);
    } else if (fx === 'phantom' && isDivision) {
      const echo = document.createElement('span');
      echo.className = 'rank-phantom-echo';
      const maskSrc = (dom.rankImage && (dom.rankImage.currentSrc || dom.rankImage.src)) || '';
      if (maskSrc) echo.style.setProperty('--rank-mask', `url("${maskSrc.replace(/\\/g, '/')}")`);
      layer.appendChild(echo);
      setTimeout(() => echo.remove(), 1700);
    } else if (fx === 'sonic' && isDivision) {
      for (let i = 0; i < 3; i++) {
        const r = document.createElement('span');
        r.className = 'rank-sonic-ring';
        r.style.animationDelay = `${0.2 + i * 0.12}s`;
        layer.appendChild(r);
        setTimeout(() => r.remove(), 1500);
      }
    } else if (fx === 'bloom' && isDivision) {
      for (let i = 0; i < 12; i++) {
        const p = document.createElement('span');
        p.className = 'rank-petal';
        p.style.left = `${20 + Math.random() * 60}%`;
        p.style.top = `${25 + Math.random() * 50}%`;
        p.style.setProperty('--sx', `${-18 + Math.random() * 36}px`);
        p.style.setProperty('--sy', `${-22 - Math.random() * 16}px`);
        p.style.setProperty('--rot', `${-40 + Math.random() * 80}deg`);
        p.style.animationDelay = `${0.2 + Math.random() * 0.4}s`;
        layer.appendChild(p);
        setTimeout(() => p.remove(), 1600);
      }
    } else if (fx === 'ink' && isDivision) {
      for (let i = 0; i < 6; i++) {
        const s = document.createElement('span');
        s.className = 'rank-ink-splash';
        s.style.left = `${25 + Math.random() * 50}%`;
        s.style.top = `${30 + Math.random() * 40}%`;
        s.style.setProperty('--sx', `${-10 + Math.random() * 20}px`);
        s.style.setProperty('--sy', `${6 + Math.random() * 12}px`);
        s.style.animationDelay = `${0.2 + Math.random() * 0.35}s`;
        layer.appendChild(s);
        setTimeout(() => s.remove(), 1500);
      }
    } else if (fx === 'prism' && isDivision) {
      for (let i = 0; i < 10; i++) {
        const s = document.createElement('span');
        s.className = 'rank-prism-shard';
        s.style.left = `${18 + Math.random() * 64}%`;
        s.style.top = `${18 + Math.random() * 60}%`;
        const ang = Math.random() * Math.PI * 2;
        s.style.setProperty('--sx', `${Math.cos(ang) * (14 + Math.random() * 20)}px`);
        s.style.setProperty('--sy', `${Math.sin(ang) * (14 + Math.random() * 20)}px`);
        s.style.setProperty('--rot', `${-50 + Math.random() * 100}deg`);
        s.style.animationDelay = `${0.25 + Math.random() * 0.3}s`;
        layer.appendChild(s);
        setTimeout(() => s.remove(), 1500);
      }
    } else if (fx === 'divine' && isDivision) {
      for (let i = 0; i < 8; i++) {
        const r = document.createElement('span');
        r.className = `rank-ray${isUp ? '' : ' hell'}`;
        r.style.setProperty('--rot', `${-40 + i * 11}deg`);
        r.style.animationDelay = `${i * 0.035}s`;
        layer.appendChild(r);
        setTimeout(() => r.remove(), 1700);
      }
    } else if (fx === 'arcane') {
      for (let i = 0; i < (isDivision ? 18 : 10); i++) {
        const m = document.createElement('span');
        m.className = 'rank-mote';
        m.style.left = `${20 + Math.random() * 60}%`;
        m.style.top = `${20 + Math.random() * 60}%`;
        const ang = Math.random() * Math.PI * 2;
        m.style.setProperty('--mx', `${Math.cos(ang) * (14 + Math.random() * 20)}px`);
        m.style.setProperty('--my', `${Math.sin(ang) * (14 + Math.random() * 20)}px`);
        m.style.animationDelay = `${Math.random() * 0.4}s`;
        m.style.width = `${3 + Math.random() * 3}px`;
        m.style.height = m.style.width;
        layer.appendChild(m);
        setTimeout(() => m.remove(), 1500);
      }
    } else if (fx === 'toxic') {
      for (let i = 0; i < n; i++) {
        const a = document.createElement('span');
        a.className = 'rank-acid';
        a.style.left = `${15 + Math.random() * 70}%`;
        a.style.top = `${15 + Math.random() * 70}%`;
        a.style.setProperty('--ax', `${-14 + Math.random() * 28}px`);
        a.style.setProperty('--ay', `${6 + Math.random() * 18}px`);
        a.style.animationDelay = `${Math.random() * 0.35}s`;
        layer.appendChild(a);
        setTimeout(() => a.remove(), 1300);
      }
    }
  }

  let idleBurstTimer = 0;
  function stopIdleSparks() {
    clearTimeout(idleBurstTimer);
    idleBurstTimer = 0;
    if (dom.rankAnimationWrapper) {
      dom.rankAnimationWrapper.classList.remove('rank-idle-burst');
    }
  }

  function triggerIdleBurst() {
    if (!idlePulse || isRankTransitionAnimating || !dom.rankAnimationWrapper) return;
    if (!dom.widget.classList.contains('visible') && !isPreview) return;
    const wrap = dom.rankAnimationWrapper;
    wrap.classList.remove('rank-idle-burst');
    void wrap.offsetWidth;
    wrap.classList.add('rank-idle-burst');
    setTimeout(() => wrap.classList.remove('rank-idle-burst'), 380);
  }

  function scheduleNextIdleBurst() {
    clearTimeout(idleBurstTimer);
    if (!idlePulse) return;
    const base = rankFx === 'neon' ? 2400 : 3400;
    const jitter = 900 + Math.random() * 2000;
    idleBurstTimer = setTimeout(() => {
      triggerIdleBurst();
      idleBurstTimer = setTimeout(scheduleNextIdleBurst, 500);
    }, base + jitter);
  }

  function setRankIdle(on) {
    if (!dom.rankAnimationWrapper) return;
    dom.rankAnimationWrapper.classList.toggle('rank-idle-pulse', !!on && idlePulse);
    if (dom.rankSection) dom.rankSection.classList.toggle('rank-idle-pulse', !!on && idlePulse);
    if (on && idlePulse) scheduleNextIdleBurst();
    else stopIdleSparks();
  }

  function animateRankTransition(fromAbsoluteIndex, toAbsoluteIndex) {
    return new Promise((resolve) => {
      if (fromAbsoluteIndex === toAbsoluteIndex) {
        applyRankVisualInstant(toAbsoluteIndex);
        resolve();
        return;
      }
      // Never leave a stuck lock — cancel in-flight FX and restart.
      if (isRankTransitionAnimating) {
        abortRankTransition(fromAbsoluteIndex);
        // abort parks currentState on `from`; the target is still where we are going.
        currentState = { ...currentState, rank: toAbsoluteIndex };
      }
      isRankTransitionAnimating = true;
      stopIdleSparks();
      clearRankAnimTimers();
      clearRankFx();

      const fromRankInfo = getRankInfo(fromAbsoluteIndex);
      const toRankInfo = getRankInfo(toAbsoluteIndex);
      const isRankUp = toAbsoluteIndex > fromAbsoluteIndex;
      const isDivisionChange = fromRankInfo.type !== toRankInfo.type;
      const fx = rankFx || 'classic';

      let animationClass = 'rank-level-change';
      let sectionClass = isRankUp ? 'rank-level rank-up' : 'rank-level rank-down';
      let animationDuration = 780;
      let swapAt = 0.46;
      if (!isDivisionChange && fx === 'blaze') {
        animationDuration = 900;
        swapAt = 0.55;
      }
      if (isDivisionChange) {
        animationClass = isRankUp ? 'rank-up-division' : 'rank-down-division';
        sectionClass = isRankUp ? 'rank-div-up rank-up' : 'rank-div-down rank-down';
        animationDuration = 2000;
        if (fx === 'neon') { animationDuration = 1550; swapAt = 0.38; }
        else if (fx === 'paper') { animationDuration = 1600; swapAt = 0.49; }
        else if (fx === 'blaze') { animationDuration = 2400; swapAt = 0.52; }
        else if (fx === 'melt') { animationDuration = 2400; swapAt = 0.48; }
        else if (fx === 'frost') { animationDuration = 2200; swapAt = 0.54; }
        else if (fx === 'divine') { animationDuration = 2100; swapAt = 0.50; }
        else if (fx === 'toxic') { animationDuration = 2200; swapAt = 0.54; }
        else if (fx === 'arcane') { animationDuration = 1900; swapAt = 0.40; }
        else if (fx === 'vortex') { animationDuration = 2000; swapAt = 0.50; }
        else if (fx === 'shatter') { animationDuration = 1900; swapAt = 0.48; }
        else if (fx === 'smoke') { animationDuration = 2100; swapAt = 0.50; }
        else if (fx === 'lightning') { animationDuration = 1600; swapAt = 0.40; }
        else if (fx === 'ripple') { animationDuration = 1800; swapAt = 0.50; }
        else if (fx === 'hologram') { animationDuration = 1900; swapAt = 0.50; }
        else if (fx === 'pixel') { animationDuration = 1500; swapAt = 0.50; }
        else if (fx === 'chrome') { animationDuration = 2000; swapAt = 0.50; }
        else if (fx === 'phantom') { animationDuration = 2000; swapAt = 0.50; }
        else if (fx === 'sonic') { animationDuration = 1700; swapAt = 0.50; }
        else if (fx === 'bloom') { animationDuration = 1900; swapAt = 0.48; }
        else if (fx === 'ink') { animationDuration = 1800; swapAt = 0.50; }
        else if (fx === 'prism') { animationDuration = 1900; swapAt = 0.50; }
        else { swapAt = 0.45; }
        playWidgetDivisionFx(isRankUp);
      }

      document.documentElement.style.setProperty('--rank-subtle-duration', `${Math.max(0.65, (isDivisionChange ? 0.85 : animationDuration / 1000))}s`);
      if (isDivisionChange) {
        document.documentElement.style.setProperty('--rank-epic-duration', `${animationDuration / 1000}s`);
      }

      setRankImages(fromRankInfo.img);
      if (dom.rankImage) {
        dom.rankImage.style.opacity = '1';
        dom.rankImage.style.visibility = 'visible';
        dom.rankImage.style.animation = 'none';
        dom.rankImage.style.filter = '';
        dom.rankImage.style.transform = '';
        dom.rankImage.style.clipPath = '';
        void dom.rankImage.offsetWidth;
        dom.rankImage.style.animation = '';
      }
      if (dom.rankImageUnder) {
        dom.rankImageUnder.style.display = 'none';
        dom.rankImageUnder.style.opacity = '0';
      }
      dom.rankValue.textContent = fromRankInfo.display;
      dom.rankAnimationWrapper.classList.add(
        animationClass,
        isRankUp ? 'rank-up' : 'rank-down',
        `theme-${fx}`
      );
      if (dom.rankSection) dom.rankSection.classList.add(...sectionClass.split(' '));
      spawnRankParticles(isRankUp, isDivisionChange);

      const swapTimer = setTimeout(() => {
        setRankImages(toRankInfo.img);
        dom.rankValue.textContent = toRankInfo.display;
      }, animationDuration * swapAt);
      rankAnimTimers.push(swapTimer);

      const doneTimer = setTimeout(() => {
        setRankImages(toRankInfo.img);
        if (dom.rankImage) {
          dom.rankImage.style.opacity = '1';
          dom.rankImage.style.filter = '';
          dom.rankImage.style.transform = '';
          dom.rankImage.style.clipPath = '';
          dom.rankImage.style.animation = 'none';
          void dom.rankImage.offsetWidth;
          dom.rankImage.style.animation = '';
        }
        if (dom.rankImageUnder) {
          dom.rankImageUnder.style.display = 'none';
          dom.rankImageUnder.style.opacity = '0';
        }
        dom.rankAnimationWrapper.classList.remove(`theme-${fx}`, 'dual-reveal');
        clearRankFx();
        isRankTransitionAnimating = false;
        if (dom.widget.classList.contains('visible') || isPreview) {
          setRankIdle(true);
        }
        previousState.rank = currentState.rank;
        resolve();
      }, animationDuration + 40);
      rankAnimTimers.push(doneTimer);
    });
  }

  let previewDemoToken = 0;

  async function restorePreviewHome(home) {
    abortRankTransition(home.rank);
    resetVesselVisual(dom.winsIconContainer, dom.winsLiquid);
    resetVesselVisual(dom.lossesIconContainer, dom.lossesLiquid);
    currentState = {
      wins: home.wins || 0,
      losses: home.losses || 0,
      rank: home.rank || 0,
      mode: home.mode || 'classic',
      role: home.role || 'tank',
      game: home.game || currentState.game || 'overwatch',
    };
    previousState = { ...currentState };
    updateRoleBadge(currentState);
    dom.wins.textContent = String(home.wins);
    dom.losses.textContent = String(home.losses);
    setLiquidLevel(dom.winsLiquid, home.wins, { instant: true });
    setLiquidLevel(dom.lossesLiquid, home.losses, { instant: true });
    applyRankVisualInstant(home.rank);
    if (dom.widget.classList.contains('visible') || isPreview) setRankIdle(true);
  }

  /** Level ↑↓ + division ↑↓, then back to starting rank — never climbs forever. */
  async function runRankShowcase(token) {
    const alive = () => token === previewDemoToken;
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
    const home = { ...currentState };

    abortRankTransition(home.rank);
    // Park near mid ladder so both up/down division FX work
    const step = currentState.game === 'apex' ? 4 : 5;
    const maxIdx = Math.max(step * 2, activeRanks().length - 1);
    let divBase = Math.floor(Math.max(step, Math.min(maxIdx - step * 2, home.rank)) / step) * step;
    if (divBase + step > maxIdx - step) {
      divBase = Math.max(0, Math.floor((maxIdx - step * 3) / step) * step);
    }
    const mid = divBase + Math.floor(step / 2);
    const lastInDiv = divBase + step - 1;
    const nextDiv = divBase + step;

    currentState = { ...home, rank: mid };
    previousState = { ...currentState };
    applyRankVisualInstant(mid);
    await sleep(180);
    if (!alive()) return;

    // Level up
    previousState = { ...currentState };
    currentState = { ...currentState, rank: mid + 1 };
    await animateRankTransition(previousState.rank, currentState.rank);
    if (!alive()) return;
    await sleep(350);
    if (!alive()) return;

    // Level down
    previousState = { ...currentState };
    currentState = { ...currentState, rank: mid };
    await animateRankTransition(previousState.rank, currentState.rank);
    if (!alive()) return;
    await sleep(350);
    if (!alive()) return;

    // Division up (from last rank of division)
    currentState = { ...home, rank: lastInDiv };
    previousState = { ...currentState };
    applyRankVisualInstant(lastInDiv);
    await sleep(120);
    if (!alive()) return;
    previousState = { ...currentState };
    currentState = { ...currentState, rank: nextDiv };
    await animateRankTransition(previousState.rank, currentState.rank);
    if (!alive()) return;
    await sleep(450);
    if (!alive()) return;

    // Division down (back)
    previousState = { ...currentState };
    currentState = { ...currentState, rank: lastInDiv };
    await animateRankTransition(previousState.rank, currentState.rank);
    if (!alive()) return;
    await sleep(300);
    if (!alive()) return;

    await restorePreviewHome(home);
  }

  async function runPreviewDemo(action) {
    if (!isPreview) return;
    const token = ++previewDemoToken;
    const alive = () => token === previewDemoToken;
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
    const home = { ...currentState };

    if (action === 'win') {
      previousState = { ...currentState };
      currentState = { ...currentState, wins: currentState.wins + 1 };
      await updateContentAnimations();
      return;
    }
    if (action === 'loss') {
      previousState = { ...currentState };
      currentState = { ...currentState, losses: currentState.losses + 1 };
      await updateContentAnimations();
      return;
    }
    if (action === 'rankStepUp' || action === 'rankStepDown') {
      const max = Math.max(0, activeRanks().length - 1);
      const next =
        action === 'rankStepUp'
          ? Math.min(max, currentState.rank + 1)
          : Math.max(0, currentState.rank - 1);
      if (next === currentState.rank) return;
      previousState = { ...currentState };
      currentState = { ...currentState, rank: next };
      await animateRankTransition(previousState.rank, currentState.rank);
      return;
    }
    if (action === 'resetStats') {
      abortRankTransition(currentState.rank);
      resetVesselVisual(dom.winsIconContainer, dom.winsLiquid);
      resetVesselVisual(dom.lossesIconContainer, dom.lossesLiquid);
      currentState = { ...currentState, wins: 0, losses: 0 };
      previousState = { ...currentState };
      dom.wins.textContent = '0';
      dom.losses.textContent = '0';
      setLiquidLevel(dom.winsLiquid, 0, { instant: true });
      setLiquidLevel(dom.lossesLiquid, 0, { instant: true });
      return;
    }
    if (action === 'rankUp' || action === 'rankUpEpic' || action === 'rankDown' || action === 'rankDownEpic' || action === 'rankShowcase') {
      await runRankShowcase(token);
      return;
    }
    if (action === 'fillCycle') {
      abortRankTransition(home.rank);
      resetVesselVisual(dom.winsIconContainer, dom.winsLiquid);
      resetVesselVisual(dom.lossesIconContainer, dom.lossesLiquid);
      const target = Math.max(0, fillLimit - 1);
      currentState = { ...home, wins: target };
      previousState = { ...currentState };
      setLiquidLevel(dom.winsLiquid, target, { instant: true });
      dom.wins.textContent = String(target);
      await sleep(120);
      if (!alive()) return;
      previousState = { ...currentState };
      currentState = { ...currentState, wins: currentState.wins + 1 };
      await updateContentAnimations();
      if (!alive()) return;
      await restorePreviewHome(home);
      return;
    }
    if (action === 'fullCycle') {
      abortRankTransition(home.rank);
      resetVesselVisual(dom.winsIconContainer, dom.winsLiquid);
      resetVesselVisual(dom.lossesIconContainer, dom.lossesLiquid);
      // Keep game/mode/role — dropping them makes Apex demos fall back to Overwatch ranks.
      currentState = { ...home, wins: 0, losses: 0 };
      previousState = { ...currentState };
      updateRoleBadge(currentState);
      dom.wins.textContent = '0';
      dom.losses.textContent = '0';
      setLiquidLevel(dom.winsLiquid, 0, { instant: true });
      setLiquidLevel(dom.lossesLiquid, 0, { instant: true });
      applyRankVisualInstant(home.rank);
      showWidget(false);
      await sleep(500);
      if (!alive()) return;

      previousState = { ...currentState };
      currentState = { ...currentState, wins: 3 };
      await updateContentAnimations();
      await sleep(400);
      if (!alive()) return;

      previousState = { ...currentState };
      currentState = { ...currentState, losses: 2 };
      await updateContentAnimations();
      await sleep(400);
      if (!alive()) return;

      await runRankShowcase(token);
      if (!alive()) return;

      // Fill-to-empty cycle then restore everything
      resetVesselVisual(dom.winsIconContainer, dom.winsLiquid);
      const target = Math.max(0, fillLimit - 1);
      currentState = { ...currentState, wins: target, losses: 2, rank: home.rank };
      previousState = { ...currentState };
      setLiquidLevel(dom.winsLiquid, target, { instant: true });
      dom.wins.textContent = String(target);
      await sleep(100);
      if (!alive()) return;
      previousState = { ...currentState };
      currentState = { ...currentState, wins: target + 1 };
      await updateContentAnimations();
      if (!alive()) return;
      await restorePreviewHome(home);
    }
  }

  function scheduleHide() {
    clearTimeout(hideWidgetTimeout);
    clearTimeout(reappearTimeout);
    // Preview stays always visible so streamers can tweak look live.
    if (isPreview || alwaysVisible || animStayTime <= 0) return;
    hideWidgetTimeout = setTimeout(() => {
      dom.widget.classList.remove('visible');
      dom.widget.classList.add('hide', `from-${animDirection}`);
      dom.rankAnimationWrapper.classList.remove('rank-idle-pulse');
      if (dom.rankSection) dom.rankSection.classList.remove('rank-idle-pulse');
      stopIdleSparks();
      if (reappearDelayMs > 0) {
        reappearTimeout = setTimeout(async () => {
          if (currentState.mode === 'roles_rotate') {
            try {
              const res = await fetch('/api/role/next', { method: 'POST' });
              if (res.ok) {
                const snap = await res.json();
                applyRemoteStateOnly(snap);
              }
            } catch {
              /* keep previous role if offline */
            }
          }
          showWidget(false);
        }, animDurationOut * 1000 + reappearDelayMs);
      }
    }, animStayTime * 1000);
  }

  function showWidget(animateContent) {
    clearTimeout(hideWidgetTimeout);
    clearTimeout(reappearTimeout);
    applyAppearEffect();
    dom.widget.classList.remove('hide', 'visible', 'from-left', 'from-right', 'from-top', 'from-bottom');
    if (appearEffect === 'slide') {
      dom.widget.classList.add(`from-${animDirection}`);
    }
    // Retrigger CSS appear animation cleanly.
    void dom.widget.offsetWidth;
    // Reveal on the next frame for a clean appear animation, but never later than
    // 60ms: frames can be paused (covered window, hidden OBS source), and waiting
    // on them left the widget invisible.
    const revealToken = ++showToken;
    const reveal = () => {
      if (revealToken !== showToken || dom.widget.classList.contains('hide')) return;
      dom.widget.classList.add('visible');
      if (appearEffect === 'slide') dom.widget.classList.add(`from-${animDirection}`);
    };
    requestAnimationFrame(reveal);
    setTimeout(reveal, 60);
    if (animateContent) {
      updateContentAnimations();
    } else {
      const info = getRankInfo(currentState.rank);
      setRankImages(info.img);
      dom.rankValue.textContent = info.display;
      dom.wins.textContent = currentState.wins;
      dom.losses.textContent = currentState.losses;
      setLiquidLevel(dom.winsLiquid, currentState.wins, { instant: true });
      setLiquidLevel(dom.lossesLiquid, currentState.losses, { instant: true });
      setRankIdle(true);
    }
    scheduleHide();
    fitPreviewWidget();
  }

  async function updateContentAnimations() {
    const shouldAnimateRank = previousState.rank !== currentState.rank;
    if (shouldAnimateRank) {
      animateRankTransition(previousState.rank, currentState.rank);
      playCardShine();
    } else {
      const info = getRankInfo(currentState.rank);
      setRankImages(info.img);
      dom.rankValue.textContent = info.display;
      setRankIdle(true);
    }
    animateNumber(dom.wins, previousState.wins, currentState.wins);
    animateNumber(dom.losses, previousState.losses, currentState.losses);
    if (currentState.wins > previousState.wins) playCardFlip('w');
    else if (currentState.losses > previousState.losses) playCardFlip('l');
    await Promise.all([
      animateFill(dom.winsLiquid, dom.winsIconContainer, previousState.wins, currentState.wins),
      animateFill(dom.lossesLiquid, dom.lossesIconContainer, previousState.losses, currentState.losses),
    ]);
  }

  function applySnapshot(snap, opts = {}) {
    const { animate = true, settingsOnly = false } = opts;
    if (snap.settings) applySettings(snap.settings);
    if (settingsOnly) {
      if (alwaysVisible) {
        clearTimeout(hideWidgetTimeout);
        clearTimeout(reappearTimeout);
        if (!dom.widget.classList.contains('visible')) showWidget(false);
      }
      if (dom.widget.classList.contains('visible') || isPreview) setRankIdle(idlePulse);
      return;
    }

    previousState = { ...currentState };
    currentState = viewFromSnap(snap);
    updateRoleBadge(currentState);

    const gameChanged = previousState.game !== currentState.game;
    const visible = dom.widget.classList.contains('visible');
    if (!visible) {
      showWidget(animate && !gameChanged);
      return;
    }
    if (animate && !gameChanged) {
      updateContentAnimations();
      scheduleHide();
    } else {
      dom.wins.textContent = currentState.wins;
      dom.losses.textContent = currentState.losses;
      setLiquidLevel(dom.winsLiquid, currentState.wins, { instant: true });
      setLiquidLevel(dom.lossesLiquid, currentState.losses, { instant: true });
      const info = getRankInfo(currentState.rank);
      setRankImages(info.img);
      dom.rankValue.textContent = info.display;
      setRankIdle(true);
      if (visible) scheduleHide();
    }
    fitPreviewWidget();
  }

  async function loadInitial() {
    const res = await fetch('/api/snapshot');
    const snap = await res.json();
    currentState = viewFromSnap(snap);
    previousState = { ...currentState };
    applySettings(snap.settings);
    updateRoleBadge(currentState);
    dom.wins.textContent = currentState.wins;
    dom.losses.textContent = currentState.losses;
    setLiquidLevel(dom.winsLiquid, currentState.wins, { instant: true });
    setLiquidLevel(dom.lossesLiquid, currentState.losses, { instant: true });
    const info = getRankInfo(currentState.rank);
    setRankImages(info.img);
    dom.rankValue.textContent = info.display;
    showWidget(false);
    fitPreviewWidget();
  }

  function connectWS() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${proto}://${location.host}/ws`);
    ws.addEventListener('message', (event) => {
      let msg;
      try {
        msg = JSON.parse(event.data);
      } catch {
        return;
      }
      if (msg.type === 'show') {
        // Preview keeps unsaved look from postMessage; don't let WS wipe rankFx.
        if (preferPreviewLocalSettings()) {
          if (msg.state || msg.view) applyRemoteStateOnly(msg);
          showWidget(false);
          return;
        }
        if (msg.settings) applySettings(msg.settings);
        if (msg.state || msg.view) {
          currentState = viewFromSnap(msg);
          previousState = { ...currentState };
          updateRoleBadge(currentState);
          dom.wins.textContent = currentState.wins;
          dom.losses.textContent = currentState.losses;
          setLiquidLevel(dom.winsLiquid, currentState.wins, { instant: true });
          setLiquidLevel(dom.lossesLiquid, currentState.losses, { instant: true });
          const info = getRankInfo(currentState.rank);
          setRankImages(info.img);
          dom.rankValue.textContent = info.display;
        }
        showWidget(false);
        return;
      }
      if (msg.type === 'settings') {
        if (preferPreviewLocalSettings()) return;
        applySnapshot(msg, { animate: false, settingsOnly: true });
        return;
      }
      // Preview with local unsaved look: sync stats/game only, keep local settings.
      if (isPreview && preferPreviewLocalSettings()) {
        if (msg.state || msg.view) applyRemoteStateOnly(msg);
        return;
      }
      if (msg.type === 'hello') {
        applySnapshot(msg, { animate: false });
        return;
      }
      applySnapshot(msg, { animate: !isPreview });
    });
    ws.addEventListener('close', () => {
      setTimeout(connectWS, 1500);
    });
  }

  if (isPreview) {
    window.addEventListener('message', (event) => {
      const data = event.data;
      if (!data || typeof data !== 'object') return;
      if (data.type === 'widget-preview-settings') {
        const next = { ...data.settings, __previewLocal: true };
        applySettings(next);
        showWidget(false);
        return;
      }
      if (data.type === 'widget-preview-state') {
        applyRemoteStateOnly(data);
        showWidget(false);
        return;
      }
      if (data.type === 'widget-preview-demo') {
        runPreviewDemo(data.action);
      }
    });
  }

  function applyOfflinePreviewDefaults() {
    currentState = {
      wins: 7,
      losses: 3,
      rank: 12,
      mode: 'classic',
      role: 'tank',
      game: 'overwatch',
      roleCycle: ['tank', 'support', 'damage'],
    };
    previousState = { ...currentState };
    applySettings({
      bgColor: 'rgba(0,0,0,0.72)',
      bgImage: '',
      winsColor: '#39ff14',
      lossesColor: '#ff3131',
      rankTextColor: '#ffffff',
      iconColor: '#ffffff',
      separatorColor: 'rgba(255,255,255,0.55)',
      font: "'Rajdhani', sans-serif",
      fontSize: 18,
      appearEffect: 'slide',
      fillStyle: 'liquid',
      rankFx: 'classic',
      animDirection: 'left',
      fillLimit: 10,
      fillDurationMs: 650,
      emptyDurationMs: 1600,
      emptyEffect: 'drain',
      vesselWave: true,
      idlePulse: true,
      skinId: 'default',
      __previewLocal: true,
    });
    updateRoleBadge(currentState);
    dom.wins.textContent = String(currentState.wins);
    dom.losses.textContent = String(currentState.losses);
    setLiquidLevel(dom.winsLiquid, currentState.wins, { instant: true });
    setLiquidLevel(dom.lossesLiquid, currentState.losses, { instant: true });
    const info = getRankInfo(currentState.rank);
    setRankImages(info.img);
    dom.rankValue.textContent = info.display;
    showWidget(false);
    fitPreviewWidget();
  }

  ensureWaveLoop();
  if (isPreview) {
    window.addEventListener('resize', fitPreviewWidget);
    if (typeof ResizeObserver !== 'undefined' && dom.widget) {
      const ro = new ResizeObserver(() => fitPreviewWidget());
      ro.observe(dom.widget);
    }
  }
  loadInitial()
    .then(connectWS)
    .catch((err) => {
      console.error(err);
      if (isPreview) {
        // Static demo / GitHub Pages: no local API — keep postMessage preview working.
        applyOfflinePreviewDefaults();
        return;
      }
      setTimeout(() => loadInitial().then(connectWS), 1500);
    });
});
