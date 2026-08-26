#!/usr/bin/env python3
"""Train banner TEXT net on binary white-on-black glyph bands (language/font agnostic)."""

from __future__ import annotations

import json
import math
import random
import sys
from pathlib import Path

try:
    from PIL import Image, ImageEnhance, ImageFilter
except ImportError:
    print("pip install pillow numpy", file=sys.stderr)
    raise

import numpy as np

ROOT = Path(__file__).resolve().parents[2]
TEMPLATES = ROOT / "internal" / "owtracker" / "assets" / "templates"
DATA_ML = ROOT / "data" / "ml" / "train"
OUT = ROOT / "internal" / "owtracker" / "ml" / "assets" / "models" / "banner_text_mlp.json"

W, H = 128, 32
H1, H2 = 256, 64
CLASSES = ["none", "win", "loss"]
EPOCHS = 45
BATCH = 32
LR = 0.002


def is_gold(r, g, b):
    return r > 130 and g > 95 and b < 150 and r > b + 15


def is_red(r, g, b):
    return r > 110 and r > g + 18 and b < 120


def extract_band(img: Image.Image) -> Image.Image:
    w, h = img.size
    band_h = max(24, min(h, h // 2))
    step = max(2, band_h // 4)
    best_y, best_score = 0, 0
    px = img.load()
    for y in range(0, max(1, h - band_h), step):
        score = 0
        for yy in range(y, min(y + band_h, h)):
            for x in range(w):
                r, g, b = px[x, yy][:3]
                if is_gold(r, g, b) or is_red(r, g, b):
                    score += 1
        if score > best_score:
            best_score, best_y = score, y
    return img.crop((0, best_y, w, min(h, best_y + band_h)))


def to_binary(img: Image.Image) -> Image.Image:
    img = extract_band(img.convert("RGB"))
    w, h = img.size
    out = Image.new("L", (w, h), 0)
    src = img.load()
    dst = out.load()
    for y in range(h):
        for x in range(w):
            r, g, b = src[x, y][:3]
            if is_gold(r, g, b) or is_red(r, g, b):
                dst[x, y] = 255
    return out.resize((W, H), Image.Resampling.LANCZOS).point(lambda v: 255 if v > 127 else 0)


def augment_bin(img: Image.Image) -> Image.Image:
    img = img.copy()
    if random.random() < 0.5:
        img = ImageEnhance.Brightness(img).enhance(random.uniform(0.85, 1.2))
    if random.random() < 0.4:
        img = img.filter(ImageFilter.GaussianBlur(radius=random.uniform(0.2, 0.9)))
    if random.random() < 0.3:
        # slight horizontal shift
        canvas = Image.new("L", (W, H), 0)
        ox = random.randint(-4, 4)
        canvas.paste(img, (ox, 0))
        img = canvas
    return img


def vec_from_bin(img: Image.Image) -> np.ndarray:
    arr = np.asarray(img, dtype=np.float32) / 255.0
    return arr.reshape(-1)


def synthetic_none(n: int) -> list[np.ndarray]:
    out = []
    for _ in range(n):
        img = Image.new("L", (W, H), 0)
        px = img.load()
        for _ in range(random.randint(20, 200)):
            x, y = random.randint(0, W - 1), random.randint(0, H - 1)
            px[x, y] = random.choice([0, 255])
        out.append(vec_from_bin(img))
    return out


def gather(label: str) -> list[np.ndarray]:
    samples = []
    dirs = [DATA_ML / "train" / label, DATA_ML / "auto" / label, DATA_ML / label]
    if label == "win":
        dirs.append(TEMPLATES)
    paths = []
    for d in dirs:
        if d.is_dir():
            paths.extend(d.glob("*.png"))
            paths.extend(d.glob("*.jpg"))
    if label == "win":
        paths.append(TEMPLATES / "win.png")
    elif label == "loss":
        paths.append(TEMPLATES / "loss.png")
    seen = set()
    for p in paths:
        rp = str(p.resolve())
        if rp in seen:
            continue
        seen.add(rp)
        try:
            bin_img = to_binary(Image.open(p))
            aug_n = 40 if "train" in p.parts or "auto" in p.parts else 24
            for _ in range(aug_n):
                samples.append(vec_from_bin(augment_bin(bin_img)))
        except OSError:
            pass
    return samples


def relu(x):
    return np.maximum(x, 0)


def softmax(x):
    x = x - x.max(axis=1, keepdims=True)
    e = np.exp(x)
    return e / e.sum(axis=1, keepdims=True)


def train(X, y):
    n, d = X.shape
    c = len(CLASSES)
    rng = np.random.default_rng(42)
    W1 = rng.normal(0, 0.04, (H1, d)).astype(np.float32)
    b1 = np.zeros(H1, dtype=np.float32)
    W2 = rng.normal(0, 0.04, (H2, H1)).astype(np.float32)
    b2 = np.zeros(H2, dtype=np.float32)
    W3 = rng.normal(0, 0.04, (c, H2)).astype(np.float32)
    b3 = np.zeros(c, dtype=np.float32)
    Y = np.zeros((n, c), dtype=np.float32)
    Y[np.arange(n), y] = 1.0
    for epoch in range(EPOCHS):
        idx = rng.permutation(n)
        loss_sum = 0.0
        for start in range(0, n, BATCH):
            batch = idx[start : start + BATCH]
            xb, yb = X[batch], Y[batch]
            h1 = relu(xb @ W1.T + b1)
            h2 = relu(h1 @ W2.T + b2)
            logits = h2 @ W3.T + b3
            prob = softmax(logits)
            loss_sum += -np.mean(np.sum(yb * np.log(prob + 1e-7), axis=1))
            grad = (prob - yb) / len(batch)
            dW3 = grad.T @ h2
            db3 = grad.sum(axis=0)
            dh2 = grad @ W3
            dh2[h2 <= 0] = 0
            dW2 = dh2.T @ h1
            db2 = dh2.sum(axis=0)
            dh1 = dh2 @ W2
            dh1[h1 <= 0] = 0
            dW1 = dh1.T @ xb
            db1 = dh1.sum(axis=0)
            W3 -= LR * dW3
            b3 -= LR * db3
            W2 -= LR * dW2
            b2 -= LR * db2
            W1 -= LR * dW1
            b1 -= LR * db1
        if (epoch + 1) % 15 == 0:
            print(f"epoch {epoch+1}/{EPOCHS} loss={loss_sum:.4f}")
    return W1, b1, W2, b2, W3, b3


def main():
    xs, ys = [], []
    for ci, label in enumerate(CLASSES):
        vecs = synthetic_none(300) if label == "none" else gather(label)
        print(f"  {label}: {len(vecs)} binary samples")
        xs.extend(vecs)
        ys.extend([ci] * len(vecs))
    X = np.stack(xs).astype(np.float32)
    y = np.array(ys, dtype=np.int64)
    print(f"Training text net on {len(y)} samples ({W}x{H} binary)...")
    W1, b1, W2, b2, W3, b3 = train(X, y)
    OUT.parent.mkdir(parents=True, exist_ok=True)
    out = {
        "inputSize": W * H,
        "hidden1": H1,
        "hidden2": H2,
        "classes": CLASSES,
        "w1": W1.reshape(-1).tolist(),
        "b1": b1.tolist(),
        "w2": W2.reshape(-1).tolist(),
        "b2": b2.tolist(),
        "w3": W3.reshape(-1).tolist(),
        "b3": b3.tolist(),
    }
    OUT.write_text(json.dumps(out), encoding="utf-8")
    (ROOT / "data" / "banner_text_mlp.json").write_text(json.dumps(out), encoding="utf-8")
    print(f"Saved {OUT}")


if __name__ == "__main__":
    main()
