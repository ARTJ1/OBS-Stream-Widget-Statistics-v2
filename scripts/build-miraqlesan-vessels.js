/**
 * Trace the miraqlesan line-art pair (assets/vessel-src/*.webp) into vessel paths:
 *   d      = filled silhouette (liquid clip + outline)
 *   detail = the drawing's lines, thickened so they still read at ~40px
 * Writes the "miraqlesan" pair into web/overlay/vessel-shapes.js.
 * Deps: same TEMP/vessel-trace node_modules as the other vessel scripts.
 */
const fs = require('fs');
const path = require('path');
const NM = path.join(process.env.TEMP, 'vessel-trace', 'node_modules');
const sharp = require(path.join(NM, 'sharp'));
const { PNG } = require(path.join(NM, 'pngjs'));
const ImageTracer = require(path.join(NM, 'imagetracerjs'));

const ROOT = path.join(__dirname, '..');
const SRC = path.join(ROOT, 'assets', 'vessel-src');
const OUT_DIR = path.join(ROOT, 'web', 'overlay', 'assets', 'vessels');
const SHAPES_JS = path.join(ROOT, 'web', 'overlay', 'vessel-shapes.js');
const S = 512;
const LINE = Number(process.env.LINE || 2); // extra px added around each source line

function dilate(m, w, h, r) {
  if (r <= 0) return m;
  const out = new Uint8Array(w * h);
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      if (!m[y * w + x]) continue;
      for (let dy = -r; dy <= r; dy++) {
        for (let dx = -r; dx <= r; dx++) {
          if (dx * dx + dy * dy > r * r + r) continue;
          const xx = x + dx;
          const yy = y + dy;
          if (xx >= 0 && yy >= 0 && xx < w && yy < h) out[yy * w + xx] = 1;
        }
      }
    }
  }
  return out;
}

function erode(m, w, h, r) {
  const inv = m.map((v) => (v ? 0 : 1));
  return dilate(inv, w, h, r).map((v) => (v ? 0 : 1));
}

/** Everything not reachable from the border without crossing ink. */
function silhouette(ink, w, h) {
  const bg = new Uint8Array(w * h);
  const st = [];
  for (let x = 0; x < w; x++) st.push(x, (h - 1) * w + x);
  for (let y = 0; y < h; y++) st.push(y * w, y * w + w - 1);
  while (st.length) {
    const i = st.pop();
    if (bg[i] || ink[i]) continue;
    bg[i] = 1;
    const x = i % w;
    if (x > 0) st.push(i - 1);
    if (x < w - 1) st.push(i + 1);
    if (i >= w) st.push(i - w);
    if (i < w * (h - 1)) st.push(i + w);
  }
  return bg.map((v) => (v ? 0 : 1));
}

/** Crop to bounds, centre in a square with a small pad, resample to S×S. */
function toSquare(masks, w, h) {
  const ref = masks[0];
  let minX = w, minY = h, maxX = -1, maxY = -1;
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      if (!ref[y * w + x]) continue;
      if (x < minX) minX = x;
      if (x > maxX) maxX = x;
      if (y < minY) minY = y;
      if (y > maxY) maxY = y;
    }
  }
  const bw = maxX - minX + 1;
  const bh = maxY - minY + 1;
  const side = Math.round(Math.max(bw, bh) * 1.03);
  const ox = Math.floor((side - bw) / 2) - minX;
  const oy = Math.floor((side - bh) / 2) - minY;
  return masks.map((m) => {
    const out = new Uint8Array(S * S);
    for (let y = 0; y < S; y++) {
      for (let x = 0; x < S; x++) {
        const sx = Math.floor((x * side) / S) - ox;
        const sy = Math.floor((y * side) / S) - oy;
        if (sx >= 0 && sy >= 0 && sx < w && sy < h) out[y * S + x] = m[sy * w + sx];
      }
    }
    return out;
  });
}

function trace(mask) {
  const data = new Uint8ClampedArray(S * S * 4);
  for (let i = 0; i < S * S; i++) {
    const v = mask[i] ? 255 : 0;
    data[i * 4] = data[i * 4 + 1] = data[i * 4 + 2] = v;
    data[i * 4 + 3] = 255;
  }
  const svg = ImageTracer.imagedataToSVG({ width: S, height: S, data }, {
    ltres: 0.5, qtres: 0.5, pathomit: 10, colorsampling: 0, numberofcolors: 2,
    blurradius: 0, strokewidth: 0, linefilter: true, roundcoords: 1, viewbox: true, desc: false,
  });
  return [...svg.matchAll(/<path\b([^>]*)>/g)]
    .map((m) => ({
      d: (m[1].match(/\bd="([^"]+)"/) || [])[1],
      fill: ((m[1].match(/\bfill="([^"]+)"/) || [])[1] || '').replace(/\s/g, ''),
    }))
    .filter((p) => p.d && p.fill.includes('255,255,255'))
    .map((p) => p.d.trim())
    .join(' ');
}

function writePng(mask, dest, preview) {
  const p = new PNG({ width: S, height: S, fill: true });
  for (let i = 0; i < S * S; i++) {
    const o = i * 4;
    if (preview) {
      p.data[o] = p.data[o + 1] = p.data[o + 2] = mask[i] ? 0 : 255;
      p.data[o + 3] = 255;
    } else {
      p.data[o] = p.data[o + 1] = p.data[o + 2] = 255;
      p.data[o + 3] = mask[i] ? 255 : 0;
    }
  }
  fs.writeFileSync(dest, PNG.sync.write(p));
}

async function build(file, name, scale) {
  const { data, info } = await sharp(path.join(SRC, file)).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
  const w = info.width;
  const h = info.height;
  const ink = new Uint8Array(w * h);
  for (let i = 0; i < w * h; i++) ink[i] = data[i * 4 + 3] >= 60 ? 1 : 0;
  // Close small gaps in the outer line before flood-filling the outside.
  const sil = erode(silhouette(dilate(ink, w, h, 2), w, h), w, h, 1);
  const lines = dilate(ink, w, h, LINE).map((v, i) => v & sil[i]);
  const [silSq, linesSq] = toSquare([sil, lines], w, h);
  writePng(silSq, path.join(OUT_DIR, `${name}.png`), false);
  writePng(silSq.map((v, i) => v & (linesSq[i] ? 0 : 1)), path.join(OUT_DIR, `${name}-preview.png`), true);
  const d = trace(silSq);
  const detail = trace(linesSq);
  fs.writeFileSync(path.join(OUT_DIR, `${name}.svg`),
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">` +
    `<path fill="#fff" fill-rule="evenodd" d="${d}"/>` +
    `<path fill="#000" fill-rule="evenodd" d="${detail}"/></svg>`);
  return { viewBox: '0 0 512 512', d, detail, stroke: 10, fillRule: 'evenodd', scale };
}

async function main() {
  const pair = {
    label: 'miraqlesan',
    labelRu: 'miraqlesan',
    wins: await build('miraqlesan-happy.webp', 'miraqlesan-happy', 1.3),
    losses: await build('miraqlesan-melt.webp', 'miraqlesan-melt', 1.45),
  };
  let src = fs.readFileSync(SHAPES_JS, 'utf8');
  const body = JSON.stringify(pair, null, 2).replace(/\n/g, '\n  ');
  // Replace an existing entry (always the last one) or append it.
  src = src.replace(/,\n  "miraqlesan": \{[\s\S]*?\n  \}\n\};/, '\n};');
  src = src.replace(/\n\};\n\n  function list/, `,\n  "miraqlesan": ${body}\n};\n\n  function list`);
  fs.writeFileSync(SHAPES_JS, src);
  console.log('miraqlesan pair written', pair.wins.detail.length, pair.losses.detail.length);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
