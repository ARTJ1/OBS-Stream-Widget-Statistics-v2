#!/usr/bin/env python3
"""Train the end-of-match banner CNN (win / loss / none) on frames from recorded streams.

Input: dataset produced by cmd/owdump (banner zone, 128x40 RGB, uint8) + labels CSV
(start,end,label in seconds; label win|loss|ignore). Frames outside labeled intervals
are "none". The model sees 96x32 images, normalized per image, and is trained with
color / brightness / occlusion augmentation so it learns the WORD SHAPE, not colors.

Pure numpy (no torch). Exports weights JSON for the Go inference (internal/owtracker).

  python scripts/ml/train_banner_cnn.py --data <prefix> --labels truth.csv --out model.json
  (several videos: --data a b --labels a.csv b.csv; unseen check: --test-data c --test-labels c.csv)
"""
from __future__ import annotations

import argparse
import csv
import json
import time

import numpy as np

SRC_W, SRC_H = 128, 40
W, H = 96, 32
CLASSES = ["none", "win", "loss"]
rng = np.random.default_rng(7)


# ---------------------------------------------------------------- data
def load(prefix: str):
    raw = np.fromfile(prefix + ".bin", dtype=np.uint8)
    n = raw.size // (SRC_W * SRC_H * 3)
    frames = raw[: n * SRC_W * SRC_H * 3].reshape(n, SRC_H, SRC_W, 3)
    times = np.array([float(r["t"]) for r in csv.DictReader(open(prefix + ".csv"))])[:n]
    return frames, times


def label_frames(times, labels_csv):
    y = np.zeros(len(times), dtype=np.int64)  # none
    keep = np.ones(len(times), dtype=bool)
    for r in csv.DictReader(open(labels_csv)):
        a, b, lab = float(r["start"]), float(r["end"]), r["label"]
        m = (times >= a) & (times <= b)
        if lab == "ignore":
            keep &= ~m
        else:
            y[m] = CLASSES.index(lab)
    return y, keep


def resize_batch(x):
    """128x40 -> 96x32 by bilinear sampling (matches Go nfnt bilinear closely enough)."""
    ys = (np.arange(H) + 0.5) * SRC_H / H - 0.5
    xs = (np.arange(W) + 0.5) * SRC_W / W - 0.5
    y0 = np.clip(np.floor(ys).astype(int), 0, SRC_H - 1)
    x0 = np.clip(np.floor(xs).astype(int), 0, SRC_W - 1)
    y1, x1 = np.clip(y0 + 1, 0, SRC_H - 1), np.clip(x0 + 1, 0, SRC_W - 1)
    wy, wx = (ys - y0)[None, :, None, None], (xs - x0)[None, None, :, None]
    x = x.astype(np.float32)
    top = x[:, y0][:, :, x0] * (1 - wx) + x[:, y0][:, :, x1] * wx
    bot = x[:, y1][:, :, x0] * (1 - wx) + x[:, y1][:, :, x1] * wx
    return top * (1 - wy) + bot * wy


def normalize(x):
    """Per-image standardization (NHWC float) -> NCHW. Same in Go."""
    m = x.mean(axis=(1, 2, 3), keepdims=True)
    s = x.std(axis=(1, 2, 3), keepdims=True) + 8.0
    return ((x - m) / s).transpose(0, 3, 1, 2).astype(np.float32)


def augment(x):
    """x: NHWC float 0..255 at 96x32."""
    n = len(x)
    # color: random channel permutation, sometimes grayscale -> color-agnostic
    perm = np.array([rng.permutation(3) for _ in range(n)])
    x = np.take_along_axis(x, perm[:, None, None, :], axis=3)
    g = rng.random(n) < 0.3
    x[g] = x[g].mean(axis=3, keepdims=True)
    # brightness / contrast (dim banners after emotional clicking)
    x = x * rng.uniform(0.25, 1.3, (n, 1, 1, 1)) + rng.uniform(-20, 20, (n, 1, 1, 1))
    # shift
    dx, dy = rng.integers(-4, 5, n), rng.integers(-2, 3, n)
    for i in range(n):
        x[i] = np.roll(np.roll(x[i], dy[i], axis=0), dx[i], axis=1)
    # occlusion: popups / player cards over the word
    occ = rng.random(n) < 0.35
    for i in np.where(occ)[0]:
        ow, oh = rng.integers(10, 40), rng.integers(8, 28)
        ox, oy = rng.integers(0, W - ow), rng.integers(0, H - oh)
        x[i, oy : oy + oh, ox : ox + ow] = rng.uniform(0, 255, 3)
    x += rng.normal(0, 6, x.shape)
    return np.clip(x, 0, 255)


# ---------------------------------------------------------------- model
def he(shape, fan_in):
    return (rng.standard_normal(shape) * np.sqrt(2.0 / fan_in)).astype(np.float32)


ARCH = [(3, 8), (8, 16), (16, 24)]  # 3x3 convs, each followed by ReLU + 2x2 maxpool
FC_HIDDEN = 64


def init_params():
    p = {}
    for i, (ci, co) in enumerate(ARCH):
        p[f"c{i}w"] = he((co, ci, 3, 3), ci * 9)
        p[f"c{i}b"] = np.zeros(co, np.float32)
    flat = ARCH[-1][1] * (H // 8) * (W // 8)
    p["f0w"], p["f0b"] = he((flat, FC_HIDDEN), flat), np.zeros(FC_HIDDEN, np.float32)
    p["f1w"], p["f1b"] = he((FC_HIDDEN, 3), FC_HIDDEN), np.zeros(3, np.float32)
    return p


def im2col(x):
    n, c, h, w = x.shape
    xp = np.pad(x, ((0, 0), (0, 0), (1, 1), (1, 1)))
    cols = np.empty((n, c, 3, 3, h, w), np.float32)
    for i in range(3):
        for j in range(3):
            cols[:, :, i, j] = xp[:, :, i : i + h, j : j + w]
    return cols.transpose(0, 4, 5, 1, 2, 3).reshape(n * h * w, c * 9)


def col2im(cols, shape):
    n, c, h, w = shape
    cols = cols.reshape(n, h, w, c, 3, 3).transpose(0, 3, 4, 5, 1, 2)
    xp = np.zeros((n, c, h + 2, w + 2), np.float32)
    for i in range(3):
        for j in range(3):
            xp[:, :, i : i + h, j : j + w] += cols[:, :, i, j]
    return xp[:, :, 1:-1, 1:-1]


def forward(p, x, train=False):
    cache = []
    for i in range(len(ARCH)):
        n, c, h, w = x.shape
        cols = im2col(x)
        wmat = p[f"c{i}w"].reshape(len(p[f"c{i}b"]), -1)
        z = (cols @ wmat.T + p[f"c{i}b"]).reshape(n, h, w, -1).transpose(0, 3, 1, 2)
        a = np.maximum(z, 0)
        r = a.reshape(n, a.shape[1], h // 2, 2, w // 2, 2)
        pooled = r.max(axis=(3, 5))
        cache.append((x.shape, cols, z, a, pooled))
        x = pooled
    flat = x.reshape(len(x), -1)
    h1 = np.maximum(flat @ p["f0w"] + p["f0b"], 0)
    logits = h1 @ p["f1w"] + p["f1b"]
    return logits, (cache, flat, h1)


def backward(p, logits, y, aux, weights):
    cache, flat, h1 = aux
    n = len(y)
    e = np.exp(logits - logits.max(1, keepdims=True))
    prob = e / e.sum(1, keepdims=True)
    loss = float(-(np.log(prob[np.arange(n), y] + 1e-9) * weights).sum() / weights.sum())
    d = prob.copy()
    d[np.arange(n), y] -= 1
    d *= (weights / weights.sum())[:, None]
    g = {"f1w": h1.T @ d, "f1b": d.sum(0)}
    dh = (d @ p["f1w"].T) * (h1 > 0)
    g["f0w"], g["f0b"] = flat.T @ dh, dh.sum(0)
    dx = (dh @ p["f0w"].T).reshape(cache[-1][4].shape)
    for i in reversed(range(len(ARCH))):
        xshape, cols, z, a, pooled = cache[i]
        nn_, c, h, w = a.shape
        r = a.reshape(nn_, c, h // 2, 2, w // 2, 2)
        mask = r == pooled[:, :, :, None, :, None]
        da = (mask * dx[:, :, :, None, :, None]).reshape(a.shape)
        dz = da * (z > 0)
        dzm = dz.transpose(0, 2, 3, 1).reshape(-1, c)
        g[f"c{i}w"] = (dzm.T @ cols).reshape(p[f"c{i}w"].shape)
        g[f"c{i}b"] = dzm.sum(0)
        if i > 0:
            dx = col2im(dzm @ p[f"c{i}w"].reshape(c, -1), xshape)
    return loss, g


def predict(p, x, bs=512):
    out = []
    for i in range(0, len(x), bs):
        logits, _ = forward(p, x[i : i + bs])
        out.append(logits.argmax(1))
    return np.concatenate(out)


# ---------------------------------------------------------------- main
def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", nargs="+", required=True, help="owdump prefixes (one per video)")
    ap.add_argument("--labels", nargs="+", required=True, help="labels CSV per --data, same order")
    ap.add_argument("--test-data", nargs="*", default=[], help="videos used only for the final check")
    ap.add_argument("--test-labels", nargs="*", default=[])
    ap.add_argument("--out", required=True)
    ap.add_argument("--split", type=float, default=0, help="seconds: train before, test after")
    ap.add_argument("--epochs", type=int, default=12)
    ap.add_argument("--extra", nargs="*", default=[], help="extra PNG crops: label=path (label win|loss|none)")
    ap.add_argument("--mine", type=int, default=0, help="hard-negative rounds (train split only)")
    ap.add_argument("--mine-epochs", type=int, default=4)
    args = ap.parse_args()

    if len(args.data) != len(args.labels) or len(args.test_data) != len(args.test_labels):
        ap.error("--data/--labels (and --test-data/--test-labels) must pair up")
    parts = []  # (small, y, train mask, test mask)
    for i, (prefix, lab) in enumerate(zip(args.data + args.test_data, args.labels + args.test_labels)):
        frames, times = load(prefix)
        yi, keep = label_frames(times, lab)
        si = np.concatenate([resize_batch(frames[j : j + 2048]) for j in range(0, len(frames), 2048)])
        del frames
        if i >= len(args.data):
            tr, te = np.zeros_like(keep), keep
        else:
            tr = keep & ((times < args.split) if args.split else np.ones_like(keep))
            te = keep & (times >= args.split) if args.split else np.zeros_like(keep)
        print(f"{prefix}: frames {len(times)}  win={(yi == 1).sum()} loss={(yi == 2).sum()}"
              f"{'  (test only)' if i >= len(args.data) else ''}")
        parts.append((si, yi, tr, te))
    small = np.concatenate([q[0] for q in parts]); y = np.concatenate([q[1] for q in parts])
    train_m = np.concatenate([q[2] for q in parts]); test_m = np.concatenate([q[3] for q in parts])
    del parts
    print(f"frames {len(y)}  train {train_m.sum()}  test {test_m.sum()}  "
          f"pos train win={((y == 1) & train_m).sum()} loss={((y == 2) & train_m).sum()}")

    xtr, ytr = small[train_m], y[train_m]
    if args.extra:
        from PIL import Image
        ex, ey = [], []
        for item in args.extra:
            lab, path = item.split("=", 1)
            im = Image.open(path).convert("RGB").resize((SRC_W, SRC_H), Image.BILINEAR)
            ex.append(np.asarray(im)); ey.append(CLASSES.index(lab))
        ex = resize_batch(np.stack(ex)); ey = np.array(ey)
        xtr, ytr = np.concatenate([xtr, ex]), np.concatenate([ytr, ey])
        print(f"extra samples: {len(ey)} (win={(ey == 1).sum()} loss={(ey == 2).sum()} none={(ey == 0).sum()})")
    # oversample banners so each epoch sees them often
    pos = np.where(ytr > 0)[0]
    neg = np.where(ytr == 0)[0]
    p = init_params()
    m = {k: np.zeros_like(v) for k, v in p.items()}
    v = {k: np.zeros_like(v_) for k, v_ in p.items()}
    lr, b1, b2, step, bs = 2e-3, 0.9, 0.999, 0, 128
    hard = np.array([], dtype=int)
    schedule = [args.epochs] + [args.mine_epochs] * args.mine
    total_ep = 0
    for rnd, n_ep in enumerate(schedule):
      if rnd > 0:
        # hard negatives: training-split "none" frames the model scores as banner
        pr = []
        for i in range(0, len(xtr), 2048):
            lo, _ = forward(p, normalize(xtr[i : i + 2048]))
            e = np.exp(lo - lo.max(1, keepdims=True)); pr.append(1 - (e / e.sum(1, keepdims=True))[:, 0])
        pb = np.concatenate(pr)
        new = np.where((ytr == 0) & (pb > 0.2))[0]
        hard = np.unique(np.concatenate([hard, new]))
        print(f"mining round {rnd}: {len(new)} new hard negatives, {len(hard)} total")
        lr = max(lr, 5e-4)
      for ep in range(n_ep):
          t0 = time.time()
          idx = np.concatenate([neg, np.repeat(pos, max(1, len(neg) // (4 * max(1, len(pos))))), np.repeat(hard, 30)])
          rng.shuffle(idx)
          losses = []
          for s in range(0, len(idx), bs):
              bi = idx[s : s + bs]
              xb = normalize(augment(xtr[bi].copy()))
              logits, aux = forward(p, xb, train=True)
              loss, g = backward(p, logits, ytr[bi], aux, np.ones(len(bi), np.float32))
              losses.append(loss)
              step += 1
              for k in p:
                  m[k] = b1 * m[k] + (1 - b1) * g[k]
                  v[k] = b2 * v[k] + (1 - b2) * g[k] ** 2
                  p[k] -= lr * (m[k] / (1 - b1**step)) / (np.sqrt(v[k] / (1 - b2**step)) + 1e-8)
          lr *= 0.8
          total_ep += 1
          print(f"epoch {total_ep}: loss {np.mean(losses):.4f}  ({time.time() - t0:.0f}s)")

    def report(name, mask):
        if mask.sum() == 0:
            return
        pred = predict(p, normalize(small[mask]))
        yt = y[mask]
        cm = np.zeros((3, 3), int)
        for a, b in zip(yt, pred):
            cm[a, b] += 1
        print(f"\n{name} confusion (rows=true none/win/loss, cols=pred):\n{cm}")

    report("TRAIN", train_m)
    report("TEST (unseen video / part)", test_m)

    model = {"w": W, "h": H, "classes": CLASSES, "arch": ARCH, "fc_hidden": FC_HIDDEN,
             "params": {k: v_.ravel().tolist() for k, v_ in p.items()},
             "shapes": {k: list(v_.shape) for k, v_ in p.items()}}
    json.dump(model, open(args.out, "w"))
    print("saved", args.out)


if __name__ == "__main__":
    main()
