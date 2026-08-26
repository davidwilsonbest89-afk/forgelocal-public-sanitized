#!/usr/bin/env python3
import json
from pathlib import Path
p = Path(__file__).parent / 'gosec_after_r4a_final.json'
for i in json.loads(p.read_text()).get('Issues', []):
    if i.get('rule_id') == 'G104':
        print(i.get('file'), i.get('line'), i.get('details'))
