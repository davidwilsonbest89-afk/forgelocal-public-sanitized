#!/usr/bin/env python3
import json
from collections import Counter, defaultdict
from pathlib import Path
p = Path('/home/ubuntu/forgelocal-final-environment-closure/independent-results/shellcheck.json')
data = json.loads(p.read_text() or '[]')
by_file = defaultdict(list)
for item in data:
    by_file[item.get('file','')].append(item)
for f in sorted(by_file):
    items = by_file[f]
    print(f'=== {f} ({len(items)}) ===')
    for x in items:
        print(f"line={x.get('line')} col={x.get('column')} SC{x.get('code')} {x.get('level')}: {x.get('message')}")
print('COUNTS_BY_LEVEL', Counter(x.get('level') for x in data))
print('COUNTS_BY_CODE', Counter(f"SC{x.get('code')}" for x in data))
