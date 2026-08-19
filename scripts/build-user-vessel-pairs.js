/**
 * Import user skull/trophy icons → white-on-transparent PNG + SVG path (no full-frame square).
 * Rebuilds fang/hard/grim/... pairs in vessel-shapes.js from these assets.
 */
const fs = require('fs');
const path = require('path');
const sharp = require(path.join(process.env.TEMP, 'vessel-trace', 'node_modules', 'sharp'));
const { PNG } = require(path.join(process.env.TEMP, 'vessel-trace', 'node_modules', 'pngjs'));
const ImageTracer = require(path.join(process.env.TEMP, 'vessel-trace', 'node_modules', 'imagetracerjs'));

const ROOT = path.join(__dirname, '..');
const ASSETS = path.join(
  process.env.USERPROFILE,
  '.cursor',
  'projects',
  'g-OBS-Stream-Widget-Statistics-v2',
  'assets',
);
const OUT_DIR = path.join(ROOT, 'web', 'overlay', 'assets', 'vessels');
const SHAPES_JS = path.join(ROOT, 'web', 'overlay', 'vessel-shapes.js');

/** Unique source files (partial name match) → output id. */
const SOURCES = {
  'trophy-classic': 'trophy_113624',
  'trophy-block': 'trophy_icon_206885',
  'trophy-wing': 'trophy_icon_178444',
  'trophy-fill': 'trophy_fill_icon_159388',
  'trophy-sharp': 'trophy_icon_136082',
  'trophy-cup': 'cup_trophy_icon_188030',
  'trophy-award': 'award_trophy_cup_winner_icon_208753',
  'trophy-medal': 'achievements_winner_trophy_award_icon_153858',
  'skull-death': 'death-skull_37988',
  'skull-detail': 'skull_icon_234382',
  'skull-glare': 'skull_icon_205843',
  'skull-simple': 'skull_icon_187601',
  'skull-emoji': 'skull-4adc7a05',
  'skull-solid': 'skull_icon_160954',
  'skull-angry': 'skull_icon_234796',
  'skull-tech': 'skull_icon_212853-48d3b7fc',
  'skull-free': 'free-skull-icon',
  'skull-bones': 'png-transparent-human-skull',
};

function find(partial) {
  const files = fs.readdirSync(ASSETS);
  const hit = files.find((f) => f.includes(partial));
  if (!hit) throw new Error(`not found: ${partial}`);
  return path.join(ASSETS, hit);
}

function loadExistingThemePairs() {
  const src = fs.readFileSync(SHAPES_JS, 'utf8');
  const m = src.match(/const PAIRS = (\{[\s\S]*?\});\s*\n\s*function list/);
  if (!m) throw new Error('cannot parse existing VesselShapes PAIRS');
  const all = JSON.parse(m[1]);
  const keep = {};
  for (const id of ['classic', 'royal', 'rage', 'inferno', 'toxic', 'nova']) {
    if (!all[id]) throw new Error(`missing theme pair ${id}`);
    keep[id] = all[id];
  }
  return keep;
}

async function decodeToRgba(filePath) {
  const { data, info } = await sharp(filePath)
    .ensureAlpha()
    .raw()
    .toBuffer({ resolveWithObject: true });
  return { data: Buffer.from(data), width: info.width, height: info.height };
}

function detectMode(rgba) {
  let bright = 0;
  let dark = 0;
  const { data, width, height } = rgba;
  for (let i = 0; i < width * height; i++) {
    const o = i * 4;
    if (data[o + 3] < 40) continue;
    const lum = (data[o] + data[o + 1] + data[o + 2]) / 3;
    if (lum > 140) bright++;
    else if (lum < 120) dark++;
  }
  return bright > dark ? 'bright' : 'dark';
}

function toTraceBitmap(rgba) {
  const mode = detectMode(rgba);
  const { width, height, data } = rgba;
  const out = new PNG({ width, height, fill: true });
  for (let i = 0; i < width * height; i++) {
    const o = i * 4;
    const a = data[o + 3];
    const lum = (data[o] + data[o + 1] + data[o + 2]) / 3;
    let on = false;
    if (a >= 40) on = mode === 'bright' ? lum > 55 : lum < 200;
    // ImageTracer uses RGB only: white = ink, black = empty (no full-frame square).
    out.data[o] = on ? 255 : 0;
    out.data[o + 1] = on ? 255 : 0;
    out.data[o + 2] = on ? 255 : 0;
    out.data[o + 3] = 255;
  }
  return out;
}

function bounds(bmp) {
  let minX = bmp.width;
  let minY = bmp.height;
  let maxX = -1;
  let maxY = -1;
  for (let y = 0; y < bmp.height; y++) {
    for (let x = 0; x < bmp.width; x++) {
      if (bmp.data[(y * bmp.width + x) * 4] < 128) continue;
      if (x < minX) minX = x;
      if (y < minY) minY = y;
      if (x > maxX) maxX = x;
      if (y > maxY) maxY = y;
    }
  }
  if (maxX < 0) throw new Error('empty silhouette');
  return { minX, minY, maxX, maxY };
}

function closeGaps(bmp, radius) {
  if (radius <= 0) return bmp;
  const { width: w, height: h, data } = bmp;
  const ink = (x, y) => x >= 0 && y >= 0 && x < w && y < h && data[(y * w + x) * 4] > 128;
  const dil = new Uint8Array(w * h);
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      let on = false;
      for (let dy = -radius; dy <= radius && !on; dy++) {
        for (let dx = -radius; dx <= radius && !on; dx++) if (ink(x + dx, y + dy)) on = true;
      }
      dil[y * w + x] = on ? 1 : 0;
    }
  }
  const out = new PNG({ width: w, height: h, fill: true });
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      let on = true;
      for (let dy = -radius; dy <= radius && on; dy++) {
        for (let dx = -radius; dx <= radius && on; dx++) {
          const xx = x + dx;
          const yy = y + dy;
          if (xx < 0 || yy < 0 || xx >= w || yy >= h || !dil[yy * w + xx]) on = false;
        }
      }
      if (ink(x, y)) on = true;
      const o = (y * w + x) * 4;
      out.data[o] = out.data[o + 1] = out.data[o + 2] = on ? 255 : 0;
      out.data[o + 3] = 255;
    }
  }
  return out;
}

function squarePad(bmp, size = 512) {
  const b = bounds(bmp);
  const bw = b.maxX - b.minX + 1;
  const bh = b.maxY - b.minY + 1;
  const pad = Math.max(4, Math.round(Math.max(bw, bh) * 0.08));
  const side = Math.max(bw, bh) + pad * 2;
  const canvas = new PNG({ width: side, height: side, fill: true });
  for (let i = 0; i < side * side; i++) {
    const o = i * 4;
    canvas.data[o] = canvas.data[o + 1] = canvas.data[o + 2] = 0;
    canvas.data[o + 3] = 255;
  }
  const ox = Math.floor((side - bw) / 2);
  const oy = Math.floor((side - bh) / 2);
  for (let y = 0; y < bh; y++) {
    for (let x = 0; x < bw; x++) {
      const src = ((b.minY + y) * bmp.width + (b.minX + x)) * 4;
      const dst = ((oy + y) * side + (ox + x)) * 4;
      const on = bmp.data[src] > 128;
      canvas.data[dst] = canvas.data[dst + 1] = canvas.data[dst + 2] = on ? 255 : 0;
      canvas.data[dst + 3] = 255;
    }
  }
  const out = new PNG({ width: size, height: size, fill: true });
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const sx = Math.min(side - 1, Math.floor((x * side) / size));
      const sy = Math.min(side - 1, Math.floor((y * side) / size));
      const s = (sy * side + sx) * 4;
      const d = (y * size + x) * 4;
      const on = canvas.data[s] > 128;
      out.data[d] = out.data[d + 1] = out.data[d + 2] = on ? 255 : 0;
      out.data[d + 3] = 255;
    }
  }
  return out;
}

/** White-on-transparent mask for CSS/SVG <image> masks (no black square bg). */
function toAlphaMask(bmp) {
  const out = new PNG({ width: bmp.width, height: bmp.height, fill: true });
  for (let i = 0; i < bmp.width * bmp.height; i++) {
    const o = i * 4;
    const on = bmp.data[o] > 128;
    out.data[o] = out.data[o + 1] = out.data[o + 2] = 255;
    out.data[o + 3] = on ? 255 : 0;
  }
  return out;
}

function preview(bmp, dest) {
  const p = new PNG({ width: bmp.width, height: bmp.height, fill: true });
  for (let i = 0; i < bmp.width * bmp.height; i++) {
    const o = i * 4;
    const on = bmp.data[o] > 128;
    p.data[o] = p.data[o + 1] = p.data[o + 2] = on ? 0 : 255;
    p.data[o + 3] = 255;
  }
  fs.writeFileSync(dest, PNG.sync.write(p));
}

function pathLenApprox(d) {
  return (String(d).match(/[MmLlCcQqZzAaSsTtHhVv]/g) || []).length;
}

/** Prefer white fill paths; drop full-frame rectangle (the "square"). */
function extractInkPath(svg) {
  const pathTags = [...svg.matchAll(/<path\b([^>]*)>/g)];
  const parsed = pathTags.map((m) => {
    const attrs = m[1];
    const d = (attrs.match(/\bd="([^"]+)"/) || [])[1];
    const fill = ((attrs.match(/\bfill="([^"]+)"/) || [])[1] || '').toLowerCase();
    return { d, fill };
  }).filter((p) => p.d);

  const isWhite = (f) => /255\s*,\s*255\s*,\s*255|#fff|#ffffff|white/.test(f);
  const isBlack = (f) => /rgb\(\s*0\s*,\s*0\s*,\s*0\s*\)|#000|#000000|black/.test(f);
  const isFullFrame = (d) =>
    /M\s*0\s+0\s+L\s*512\s+0\s+L\s*512\s+512\s+L\s*0\s+512/i.test(d) ||
    /M\s*0\s+0\s+L\s*\d+\s+0\s+L\s*\d+\s+\d+\s+L\s*0\s+\d+/i.test(d.split(/Z/i)[0] || '');

  let whites = parsed.filter((p) => isWhite(p.fill));
  if (!whites.length) {
    // Fallback: longest non-full-frame path
    whites = parsed.filter((p) => !isFullFrame(p.d));
  }
  whites = whites.filter((p) => !isFullFrame(p.d));
  if (!whites.length) throw new Error('no ink path (only square?)');

  whites.sort((a, b) => b.d.length - a.d.length);
  // One compound path already has holes; if multiple white paths, join them.
  const main = whites[0].d;
  const extras = whites.slice(1).filter((p) => p.d.length > main.length * 0.05).map((p) => p.d);
  return [main, ...extras].join(' ');
}

function trace(bmp) {
  const imageData = {
    width: bmp.width,
    height: bmp.height,
    data: new Uint8ClampedArray(bmp.data),
  };
  const svg = ImageTracer.imagedataToSVG(imageData, {
    ltres: 0.6,
    qtres: 0.6,
    pathomit: 4,
    colorsampling: 0,
    numberofcolors: 2,
    blurradius: 0,
    blurdelta: 20,
    strokewidth: 0,
    linefilter: true,
    scale: 1,
    roundcoords: 1,
    viewbox: true,
    desc: false,
  });
  const d = extractInkPath(svg);
  // Clean SVG for assets: only the ink path, transparent bg.
  const cleanSvg =
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">` +
    `<path fill="#fff" fill-rule="evenodd" d="${d}"/>` +
    `</svg>`;
  return { svg: cleanSvg, d, cmds: pathLenApprox(d) };
}

function tracedShape(entry) {
  return {
    viewBox: '0 0 512 512',
    d: entry.d,
    mask: entry.png, // white-on-transparent PNG — used for paint (no square)
    stroke: 12,
    fillRule: 'evenodd',
  };
}

function cleanJunk() {
  if (!fs.existsSync(OUT_DIR)) return;
  for (const f of fs.readdirSync(OUT_DIR)) {
    if (f === '_clean') {
      fs.rmSync(path.join(OUT_DIR, '_clean'), { recursive: true, force: true });
      continue;
    }
    if (f.endsWith('-preview.png') || f === 'user-catalog.json' || f === 'manifest.json') {
      fs.unlinkSync(path.join(OUT_DIR, f));
    }
  }
}

async function main() {
  fs.mkdirSync(OUT_DIR, { recursive: true });
  cleanJunk();
  const catalog = {};

  for (const [name, partial] of Object.entries(SOURCES)) {
    try {
      const src = find(partial);
      const rgba = await decodeToRgba(src);
      // Skip nearly-solid canvases (e.g. opaque black JPEG with no silhouette).
      const mode = detectMode(rgba);
      let ink = 0;
      for (let i = 0; i < rgba.width * rgba.height; i++) {
        const o = i * 4;
        if (rgba.data[o + 3] < 40) continue;
        const lum = (rgba.data[o] + rgba.data[o + 1] + rgba.data[o + 2]) / 3;
        if (mode === 'bright' ? lum > 55 : lum < 200) ink++;
      }
      const inkPct = ink / (rgba.width * rgba.height);
      if (inkPct > 0.92 || inkPct < 0.02) {
        console.warn('skip', name, `ink=${(inkPct * 100).toFixed(1)}% (looks like solid square)`);
        continue;
      }

      let bmp = toTraceBitmap(rgba);
      bmp = closeGaps(bmp, name.startsWith('trophy') ? 2 : 0);
      bmp = squarePad(bmp, 512);
      const alpha = toAlphaMask(bmp);
      fs.writeFileSync(path.join(OUT_DIR, `${name}.png`), PNG.sync.write(alpha));
      preview(bmp, path.join(OUT_DIR, `${name}-preview.png`));
      const { svg, d, cmds } = trace(bmp);
      fs.writeFileSync(path.join(OUT_DIR, `${name}.svg`), svg);
      catalog[name] = { viewBox: '0 0 512 512', d, cmds, png: `assets/vessels/${name}.png` };
      console.log('ok', name, `cmds≈${cmds}`, `path=${d.length}`);
    } catch (e) {
      console.warn('fail', name, e.message);
    }
  }

  fs.writeFileSync(path.join(OUT_DIR, 'user-catalog.json'), JSON.stringify(catalog, null, 2));

  const need = (id) => {
    if (!catalog[id]) throw new Error(`required icon missing: ${id}`);
    return tracedShape(catalog[id]);
  };

  // Many cup+skull pairs from user art.
  const userPairs = {
    fang: {
      label: 'Fang',
      labelRu: 'Клык',
      wins: need('trophy-classic'),
      losses: catalog['skull-angry'] ? tracedShape(catalog['skull-angry']) : need('skull-glare'),
    },
    hard: {
      label: 'Hard',
      labelRu: 'Хард',
      wins: need('trophy-block'),
      losses: need('skull-detail'),
    },
    grim: {
      label: 'Grim',
      labelRu: 'Мрачный',
      wins: need('trophy-wing'),
      losses: need('skull-death'),
    },
    solid: {
      label: 'Solid',
      labelRu: 'Солид',
      wins: catalog['trophy-fill'] ? tracedShape(catalog['trophy-fill']) : need('trophy-classic'),
      losses: need('skull-solid'),
    },
    emoji: {
      label: 'Emoji',
      labelRu: 'Эмодзи',
      wins: need('trophy-block'),
      losses: need('skull-emoji'),
    },
    simple: {
      label: 'Simple',
      labelRu: 'Простой',
      wins: need('trophy-wing'),
      losses: need('skull-simple'),
    },
  };

  if (catalog['skull-tech'] && catalog['trophy-sharp']) {
    userPairs.tech = {
      label: 'Tech',
      labelRu: 'Тех',
      wins: tracedShape(catalog['trophy-sharp']),
      losses: tracedShape(catalog['skull-tech']),
    };
  }
  if (catalog['skull-free'] && (catalog['trophy-cup'] || catalog['trophy-award'])) {
    userPairs.free = {
      label: 'Free',
      labelRu: 'Свободный',
      wins: tracedShape(catalog['trophy-cup'] || catalog['trophy-award'] || catalog['trophy-medal']),
      losses: tracedShape(catalog['skull-free']),
    };
  }
  if (catalog['skull-bones'] && catalog['trophy-award']) {
    userPairs.cross = {
      label: 'Cross',
      labelRu: 'Крест',
      wins: tracedShape(catalog['trophy-award'] || catalog['trophy-medal']),
      losses: tracedShape(catalog['skull-bones']),
    };
  }

  const pairs = { ...loadExistingThemePairs(), ...userPairs };

  const src = `/* Paired vessel icon sets. Theme = FA6; fang/hard/grim/... = user icons (SVG path, no square bg). */
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
  console.log('pairs:', Object.keys(pairs).join(', '));
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
