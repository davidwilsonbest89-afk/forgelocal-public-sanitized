#!/usr/bin/env python3
"""Inject complete WebGL fingerprints into BrowseForge fingerprint pools.

Uses camoufox.webgl.sample_webgl() to generate full WebGL configs.
Run after generate-fingerprints.js as part of the build process.

Requirements: pip install camoufox
Usage: python scripts/inject-webgl.py
"""

import json, os, sys

try:
    from camoufox.webgl import sample_webgl
except ImportError:
    print("ERROR: pip install camoufox")
    sys.exit(1)

DATA_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "data")
OS_MAP = {"windows": "win", "macos": "mac", "linux": "lin"}

def main():
    files = sorted(f for f in os.listdir(DATA_DIR) if f.startswith("fingerprints-") and f.endswith(".json"))
    for filename in files:
        os_code = OS_MAP.get(filename.replace(".json", "").split("-")[-1])
        if not os_code:
            continue
        filepath = os.path.join(DATA_DIR, filename)
        with open(filepath) as f:
            pool = json.load(f)
        for fp in pool:
            webgl = sample_webgl(os_code)
            fp.pop("webGl:renderer", None)
            fp.pop("webGl:vendor", None)
            fp.update(webgl)
        with open(filepath, "w") as f:
            json.dump(pool, f, separators=(",", ":"))
        print(f"  ✅ {filename}: {len(pool)} fingerprints")

if __name__ == "__main__":
    print("Injecting WebGL via camoufox.webgl.sample_webgl()...")
    main()
    print("Done.")
