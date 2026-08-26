#!/usr/bin/env python3
"""Train tiny banner MLP (win/loss/none) and export banner_mlp.json for pure-Go inference."""

from __future__ import annotations

import json
import math
import os
import random
import struct
import sys
import zlib
from pathlib import Path

try:
    from PIL import Image, ImageEnhance, ImageFilter, ImageOps
except ImportError:
    print("Install: pip install pillow numpy", file=sys.stderr)
    raise

import numpy as np

ROOT = Path(__file__).resolve().parents[2]
ASSETS = ROOT / "internal" / "owtracker" / "assets"
TEMPLATES = ASSETS / "templates"
MODEL_OUT = ROOT / "internal" / "owtracker" / "ml" / "assets" / "models" / "banner_mlp.json"
DATA_ML = ROOT / "data" / "ml" / "train"
DEBUG_CROPS = ROOT / "data" / "ow_debug_crops"

INPUT = 48
H1, H2 = 128, 64
CLASSES = ["none", "win", "loss"]
EPOCHS = 40
BATCH = 32
LR = 0.002


def load_rgb(path: Path, size: int = INPUT) -> np.ndarray:
    img = Image.open(path).convert("RGB")
    img = img.resize((size, size), Image.Resampling.LANCZOS)
    arr = np.asarray(img, dtype=np.float32) / 255.0
    return arr.transpose(2, 0, 1).reshape(-1)


def augment(img: Image.Image) -> Image.Image:
    img = img.copy()
    if random.random() < 0.5:
        img = ImageEnhance.Brightness(img).enhance(random.uniform(0.7, 1.35))
    if random.random() < 0.5:
        img = ImageEnhance.Contrast(img).enhance(random.uniform(0.75, 1.4))
    if random.random() < 0.4:
        img = img.filter(ImageFilter.GaussianBlur(radius=random.uniform(0.3, 1.2)))
    if random.random() < 0.3:
        img = ImageOps.autocontrast(img)
    return img


def synthetic_none(n: int) -> list[np.ndarray]:
    out = []
    for _ in range(n):
        base = random.randint(15, 80)
        img = Image.new("RGB", (INPUT, INPUT), (base, base, base + random.randint(-10, 10)))
        pix = img.load()
        for _ in range(random.randint(50, 400)):
            x, y = random.randint(0, INPUT - 1), random.randint(0, INPUT - 1)
            c = random.randint(0, 255)
            pix[x, y] = (c, c // 2, c // 3)
        out.append(load_rgb_from_pil(img))
    return out


def load_rgb_from_pil(img: Image.Image) -> np.ndarray:
    img = img.resize((INPUT, INPUT), Image.Resampling.LANCZOS)
    arr = np.asarray(img, dtype=np.float32) / 255.0
    return arr.transpose(2, 0, 1).reshape(-1)


def gather_class(label: str, per_seed: int = 120) -> list[np.ndarray]:
    samples: list[np.ndarray] = []
    dirs = [DATA_ML / "train" / label, DATA_ML / "auto" / label]
    if label == "win":
        dirs.append(TEMPLATES)
        seeds = [TEMPLATES / "win.png"]
    elif label == "loss":
        seeds = [TEMPLATES / "loss.png"]
    else:
        seeds = []

    for d in dirs:
        if d.is_dir():
            aug_n = 32 if "auto" in d.parts or "train" in d.parts else 8
            for p in list(d.glob("*.png")) + list(d.glob("*.jpg")):
                try:
                    img = Image.open(p).convert("RGB")
                    for _ in range(aug_n):
                        samples.append(load_rgb_from_pil(augment(img)))
                except OSError:
                    pass

    for seed in seeds:
        if seed.is_file():
            img = Image.open(seed).convert("RGB")
            for _ in range(per_seed):
                samples.append(load_rgb_from_pil(augment(img)))
    return samples


def relu(x: np.ndarray) -> np.ndarray:
    return np.maximum(x, 0)


def softmax(x: np.ndarray) -> np.ndarray:
    x = x - x.max(axis=1, keepdims=True)
    e = np.exp(x)
    return e / e.sum(axis=1, keepdims=True)


def train(X: np.ndarray, y: np.ndarray):
    n, d = X.shape
    c = len(CLASSES)
    rng = np.random.default_rng(42)
    W1 = rng.normal(0, 0.05, (H1, d)).astype(np.float32)
    b1 = np.zeros(H1, dtype=np.float32)
    W2 = rng.normal(0, 0.05, (H2, H1)).astype(np.float32)
    b2 = np.zeros(H2, dtype=np.float32)
    W3 = rng.normal(0, 0.05, (c, H2)).astype(np.float32)
    b3 = np.zeros(c, dtype=np.float32)

    Y = np.zeros((n, c), dtype=np.float32)
    Y[np.arange(n), y] = 1.0

    for epoch in range(EPOCHS):
        idx = rng.permutation(n)
        total = 0.0
        for start in range(0, n, BATCH):
            batch = idx[start : start + BATCH]
            xb = X[batch]
            yb = Y[batch]
            h1 = relu(xb @ W1.T + b1)
            h2 = relu(h1 @ W2.T + b2)
            logits = h2 @ W3.T + b3
            prob = softmax(logits)
            loss = -np.mean(np.sum(yb * np.log(prob + 1e-7), axis=1))
            total += loss
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
        if (epoch + 1) % 10 == 0:
            print(f"epoch {epoch+1}/{EPOCHS} loss={total:.4f}")

    return W1, b1, W2, b2, W3, b3


def main():
    print("Gathering training samples...")
    xs, ys = [], []
    for ci, label in enumerate(CLASSES):
        if label == "none":
            vecs = synthetic_none(400)
        else:
            vecs = gather_class(label)
        print(f"  {label}: {len(vecs)} samples")
        xs.extend(vecs)
        ys.extend([ci] * len(vecs))

    X = np.stack(xs).astype(np.float32)
    y = np.array(ys, dtype=np.int64)
    print(f"Training on {len(y)} samples...")
    W1, b1, W2, b2, W3, b3 = train(X, y)

    MODEL_OUT.parent.mkdir(parents=True, exist_ok=True)
    out = {
        "inputSize": INPUT,
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
    MODEL_OUT.write_text(json.dumps(out), encoding="utf-8")
    data_copy = ROOT / "data" / "banner_mlp.json"
    data_copy.write_text(json.dumps(out), encoding="utf-8")
    print(f"Saved {MODEL_OUT}")
    print(f"Saved {data_copy}")


if __name__ == "__main__":
    main()
