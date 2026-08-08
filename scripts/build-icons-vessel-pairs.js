/**
 * Build vessel-shapes.js user pairs from real SVGs in web/overlay/assets/icons.
 * Same path+clipPath paint path as Classic — liquid fill works.
 */
const fs = require('fs');
const path = require('path');

const ROOT = path.join(__dirname, '..');
const ICONS = path.join(ROOT, 'web', 'overlay', 'assets', 'icons');
const SHAPES_JS = path.join(ROOT, 'web', 'overlay', 'vessel-shapes.js');

function loadThemePairs() {
  const src = fs.readFileSync(SHAPES_JS, 'utf8');
  const m = src.match(/const PAIRS = (\{[\s\S]*?\});\s*\n\s*function list/);
  if (!m) throw new Error('cannot parse VesselShapes');
  const all = JSON.parse(m[1]);
  const keep = {};
  for (const id of ['classic', 'royal', 'rage', 'inferno', 'toxic', 'nova']) {
    if (!all[id]) throw new Error(`missing ${id}`);
    // Drop any leftover mask fields from theme pairs
    keep[id] = {
      label: all[id].label,
      labelRu: all[id].labelRu,
      wins: { viewBox: all[id].wins.viewBox, d: all[id].wins.d, stroke: all[id].wins.stroke },
      losses: { viewBox: all[id].losses.viewBox, d: all[id].losses.d, stroke: all[id].losses.stroke },
    };
    if (all[id].wins.fillRule) keep[id].wins.fillRule = all[id].wins.fillRule;
    if (all[id].losses.fillRule) keep[id].losses.fillRule = all[id].losses.fillRule;
  }
  return keep;
}

function isBlankPath(d, fill) {
  if (!d) return true;
  if (/fill\s*=\s*["']none["']/i.test(fill || '')) return true;
  // Empty viewBox rects used as hit areas
  if (/^M\s*0\s+0\s+h?\s*\d+\s+v?\s*\d+\s+H?\s*0\s*z?$/i.test(d.replace(/,/g, ' ').replace(/\s+/g, ' ').trim())) {
    return true;
  }
  if (/M0 0h\d+v\d+H0z/i.test(d.replace(/\s+/g, ''))) return true;
  if (/M 0 0 L \d+ 0 L \d+ \d+ L 0 \d+/i.test(d) && d.length < 80) return true;
  return false;
}

function circleToPath(cx, cy, r) {
  // Two semicircles → full circle subpath (evenodd hole-friendly).
  return `M ${cx - r} ${cy} a ${r} ${r} 0 1 0 ${r * 2} 0 a ${r} ${r} 0 1 0 ${-r * 2} 0`;
}

function polygonToPath(points) {
  const pts = String(points)
    .trim()
    .split(/[\s,]+/)
    .map(Number)
    .filter((n) => !Number.isNaN(n));
  if (pts.length < 6) return null;
  let d = `M ${pts[0]} ${pts[1]}`;
  for (let i = 2; i < pts.length; i += 2) d += ` L ${pts[i]} ${pts[i + 1]}`;
  return `${d} Z`;
}

function extractShape(fileName) {
  const file = path.join(ICONS, fileName);
  if (!fs.existsSync(file)) throw new Error(`missing icon ${fileName}`);
  const svg = fs.readFileSync(file, 'utf8');
  const vbMatch = svg.match(/viewBox\s*=\s*["']([^"']+)["']/i);
  if (!vbMatch) throw new Error(`no viewBox in ${fileName}`);
  const viewBox = vbMatch[1].trim().replace(/,/g, ' ').replace(/\s+/g, ' ');

  const ds = [];
  let holeShapes = 0;
  const pathRe = /<path\b([^>]*)>/gi;
  let m;
  while ((m = pathRe.exec(svg))) {
    const attrs = m[1];
    const d = (attrs.match(/\bd\s*=\s*["']([^"']+)["']/) || [])[1];
    const fill = (attrs.match(/\bfill\s*=\s*["']([^"']+)["']/) || [])[1] || '';
    if (isBlankPath(d, fill) || fill.toLowerCase() === 'none') continue;
    ds.push(d.trim());
  }
  const circleRe = /<circle\b([^>]*)\/?>/gi;
  while ((m = circleRe.exec(svg))) {
    const attrs = m[1];
    const fill = (attrs.match(/\bfill\s*=\s*["']([^"']+)["']/) || [])[1] || '';
    if (fill.toLowerCase() === 'none') continue;
    const cx = Number((attrs.match(/\bcx\s*=\s*["']([^"']+)["']/) || [])[1]);
    const cy = Number((attrs.match(/\bcy\s*=\s*["']([^"']+)["']/) || [])[1]);
    const r = Number((attrs.match(/\br\s*=\s*["']([^"']+)["']/) || [])[1]);
    if ([cx, cy, r].some((n) => Number.isNaN(n)) || r <= 0) continue;
    ds.push(circleToPath(cx, cy, r));
    holeShapes += 1;
  }
  const polyRe = /<polygon\b([^>]*)\/?>/gi;
  while ((m = polyRe.exec(svg))) {
    const attrs = m[1];
    const fill = (attrs.match(/\bfill\s*=\s*["']([^"']+)["']/) || [])[1] || '';
    if (fill.toLowerCase() === 'none') continue;
    const points = (attrs.match(/\bpoints\s*=\s*["']([^"']+)["']/) || [])[1];
    const d = polygonToPath(points);
    if (d) {
      ds.push(d);
      holeShapes += 1;
    }
  }

  if (!ds.length) throw new Error(`no drawable path in ${fileName}`);

  const d = ds.join(' ');
  let outViewBox = viewBox;

  // Inkscape medal SVG: path lives in ~24×24 coords, viewBox is 6.35 with transforms.
  // We take raw `d` (transforms ignored) → force a fitting viewBox.
  if (/achievements_winner/i.test(fileName)) {
    outViewBox = '0 0 24 24';
  } else if (/transform\s*=\s*["']matrix/i.test(svg)) {
    const wh = viewBox.split(/\s+/).map(Number).slice(2);
    if (Math.max(wh[0] || 0, wh[1] || 0) < 12) outViewBox = '0 0 24 24';
  }

  const parts = outViewBox.split(/\s+/).map(Number);
  const vbW = parts[2] || 512;
  const vbH = parts[3] || 512;
  const stroke = Math.max(1, Math.round(Math.max(vbW, vbH) * 0.03));
  // evenodd only when extra circles/polygons are eye/detail cutouts — not for multi-part trophies
  const fillRule = holeShapes > 0 ? 'evenodd' : undefined;
  const shape = { viewBox: outViewBox, d, stroke };
  if (fillRule) shape.fillRule = fillRule;
  return shape;
}

function pair(label, labelRu, winsFile, lossesFile) {
  return {
    label,
    labelRu,
    wins: extractShape(winsFile),
    losses: extractShape(lossesFile),
  };
}

function main() {
  const userPairs = {
    fang: pair('Fang', 'Клык', 'trophy_113624.svg', 'skull_icon_205843.svg'),
    hard: pair('Hard', 'Хард', 'trophy_icon_206885.svg', 'skull_icon_234382.svg'),
    grim: pair('Grim', 'Мрачный', 'trophy_icon_178444.svg', '08dbb3cd-dc9f-408d-893d-6a9d9cb0024c.svg'),
    solid: pair('Solid', 'Солид', 'trophy_fill_icon_159388.svg', 'skull_icon_160954.svg'),
    sharp: pair('Sharp', 'Острый', 'trophy_icon_136082.svg', 'skull_icon_187601.svg'),
    award: pair('Award', 'Награда', 'award_trophy_cup_winner_icon_208753.svg', 'skull_icon_212853.svg'),
    medal: pair('Medal', 'Медаль', 'achievements_winner_trophy_award_icon_153858.svg', 'skull2.svg'),
  };

  // Optional extras if present
  if (fs.existsSync(path.join(ICONS, 'trophy.svg')) && fs.existsSync(path.join(ICONS, 'skull.svg'))) {
    // Same as classic — skip to avoid clutter
  }

  const pairs = { ...loadThemePairs(), ...userPairs };

  const src = `/* Paired vessel icons. Theme = FA6; fang/hard/... = paths from assets/icons/*.svg */
(function (global) {
  const PAIRS = ${JSON.stringify(pairs, null, 2)};

  function list() {
    return Object.keys(PAIRS).map((id) => ({
      id,
      label: PAIRS[id].label,
      labelRu: PAIRS[id].labelRu,
    }));
  }

  function get(id) {
    return PAIRS[id] || PAIRS.classic;
  }

  function getSide(id, side) {
    const pair = get(id);
    return side === 'losses' ? pair.losses : pair.wins;
  }

  global.VesselShapes = { PAIRS, list, get, getSide };
})(typeof window !== 'undefined' ? window : globalThis);
`;

  fs.writeFileSync(SHAPES_JS, src);
  for (const [id, p] of Object.entries(userPairs)) {
    console.log(
      'ok',
      id,
      `wins=${p.wins.viewBox} d=${p.wins.d.length}`,
      `losses=${p.losses.viewBox} d=${p.losses.d.length}`,
    );
  }
  console.log('pairs:', Object.keys(pairs).join(', '));
}

main();
