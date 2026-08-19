/**
 * Normalize user-provided skull/trophy PNGs into white-on-transparent masks
 * and optional SVG path traces for vessel fill.
 */
const fs = require('fs');
const path = require('path');
const { PNG } = require(path.join(process.env.TEMP, 'vessel-trace', 'node_modules', 'pngjs'));
const ImageTracer = require(path.join(process.env.TEMP, 'vessel-trace', 'node_modules', 'imagetracerjs'));

const ASSETS = path.join(
  process.env.USERPROFILE,
  '.cursor',
  'projects',
  'g-OBS-Stream-Widget-Statistics-v2',
  'assets',
);
const OUT_DIR = path.join(__dirname, '..', 'web', 'overlay', 'assets', 'vessels');

function find(partial) {
  const files = fs.readdirSync(ASSETS);
  const hit = files.find((f) => f.includes(partial) && f.endsWith('.png'));
  if (!hit) throw new Error(`not found: ${partial}`);
  return path.join(ASSETS, hit);
}

/** Pick distinct single-icon sources (user-provided). */
const SOURCES = {
  'wins-angular.png': 'trophy_icon_136082',
  'wins-block.png': 'trophy_fill_icon_159388',
  'wins-wing.png': 'trophy_icon_178444',
  'wins-cup.png': 'cup_trophy_icon_188030',
  'wins-award.png': 'award_trophy_cup_winner_icon_208753',
  'wins-classic-alt.png': 'achievements_winner_trophy_award_icon_153858',
  'losses-angry.png': 'skull_icon_234796',
  'losses-fang.png': 'death-skull_37988',
  'losses-hard.png': 'skull_icon_187601',
  'losses-glare.png': 'skull_icon_205843',
  'losses-tech.png': 'skull_icon_212853-48d3b7fc',
  'losses-solid.png': 'skull_icon_160954',
};

function toRGBA(png) {
  return {
    width: png.width,
    height: png.height,
    data: new Uint8ClampedArray(png.data),
  };
}

function normalizeSilhouette(png) {
  // Make opaque dark/light ink into white silhouette on transparent.
  const out = new PNG({ width: png.width, height: png.height });
  for (let i = 0; i < png.width * png.height; i++) {
    const o = i * 4;
    const r = png.data[o];
    const g = png.data[o + 1];
    const b = png.data[o + 2];
    const a = png.data[o + 3];
    const lum = (r + g + b) / 3;
    // Treat near-black transparent OR near-white as background depending on majority.
    // Prefer: any visible non-near-background pixel becomes white.
    const isTransparent = a < 16;
    const isNearBlack = lum < 18;
    const isNearWhite = lum > 240;
    // Icons are usually white on black OR black on transparent.
    let on = false;
    if (!isTransparent) {
      if (isNearBlack && !isNearWhite) {
        // black ink on light/transparent -> ink
        on = a > 40;
      } else {
        // white/gray ink on black -> ink if bright enough
        on = lum > 40 && a > 40;
      }
    }
    // Heuristic refinement: if image is mostly black canvas with light icon
    // handled by lum>40. If black icon on transparent, a>40 && lum low still on.
    if (!isTransparent && a > 40 && (lum < 200 || lum > 40)) {
      // For black-on-transparent: lum low, a high -> on
      // For white-on-black: lum high -> on; black bg lum low a high -> off if lum < 30
    }
    if (!isTransparent) {
      if (lum >= 30) on = true; // light icon
      else if (a > 40 && lum < 30) on = true; // dark icon on transparent
      // black background of white icons is opaque black — exclude
      if (lum < 30 && a > 200 && r === g && g === b) {
        // could be black bg OR black icon. Check neighbors later — for now:
        // if average image luminance is low, bright pixels are icon.
      }
    } else on = false;

    out.data[o] = 255;
    out.data[o + 1] = 255;
    out.data[o + 2] = 255;
    out.data[o + 3] = on ? 255 : 0;
  }

  // Second pass: decide mode by mean luminance of opaque pixels
  let sum = 0;
  let count = 0;
  for (let i = 0; i < png.width * png.height; i++) {
    const o = i * 4;
    if (png.data[o + 3] > 40) {
      sum += (png.data[o] + png.data[o + 1] + png.data[o + 2]) / 3;
      count++;
    }
  }
  const mean = count ? sum / count : 0;
  const whiteOnBlack = mean < 90; // overall dark image => white icons

  for (let i = 0; i < png.width * png.height; i++) {
    const o = i * 4;
    const r = png.data[o];
    const g = png.data[o + 1];
    const b = png.data[o + 2];
    const a = png.data[o + 3];
    const lum = (r + g + b) / 3;
    let on = false;
    if (a < 16) on = false;
    else if (whiteOnBlack) on = lum > 50;
    else on = a > 40 && lum < 220; // black icon on light/transparent
    out.data[o] = 255;
    out.data[o + 1] = 255;
    out.data[o + 2] = 255;
    out.data[o + 3] = on ? 255 : 0;
  }
  return out;
}

function extractPathFromSvg(svg) {
  const paths = [...svg.matchAll(/d="([^"]+)"/g)].map((m) => m[1]);
  if (!paths.length) return null;
  // Merge into one path for clip
  return paths.join(' ');
}

function main() {
  fs.mkdirSync(OUT_DIR, { recursive: true });
  const manifest = {};

  for (const [outName, partial] of Object.entries(SOURCES)) {
    const src = find(partial);
    const png = PNG.sync.read(fs.readFileSync(src));
    const sil = normalizeSilhouette(png);
    const outPngPath = path.join(OUT_DIR, outName);
    fs.writeFileSync(outPngPath, PNG.sync.write(sil));

    // Trace to SVG path (optional, for path-mode)
    const imageData = toRGBA(sil);
    // imagetracer wants ImageData-like; use imagedataToSVG if available
    let svg = '';
    try {
      svg = ImageTracer.imagedataToSVG(imageData, {
        ltres: 1,
        qtres: 1,
        pathomit: 8,
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
    } catch (e) {
      console.warn('trace fail', outName, e.message);
    }

    let d = svg ? extractPathFromSvg(svg) : null;
    if (svg) {
      fs.writeFileSync(path.join(OUT_DIR, outName.replace(/\.png$/, '.svg')), svg);
    }

    manifest[outName.replace(/\.png$/, '')] = {
      png: `assets/vessels/${outName}`,
      svg: d ? `assets/vessels/${outName.replace(/\.png$/, '.svg')}` : null,
      viewBox: `0 0 ${sil.width} ${sil.height}`,
      d: d,
      w: sil.width,
      h: sil.height,
    };
    console.log('ok', outName, sil.width + 'x' + sil.height, d ? `path=${d.length}` : 'no-path');
  }

  fs.writeFileSync(path.join(OUT_DIR, 'manifest.json'), JSON.stringify(manifest, null, 2));
  console.log('wrote', OUT_DIR);
}

main();
