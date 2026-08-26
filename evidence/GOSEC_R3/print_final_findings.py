#!/usr/bin/env python3
import json
from pathlib import Path
p = Path(__file__).parent / 'gosec_source_only_final_r3.json'
data = json.loads(p.read_text())
for i in data.get('Issues', []):
    print(f"{i.get('rule_id')}\t{i.get('severity')}\t{i.get('confidence')}\t{i.get('file')}\t{i.get('line')}\t{i.get('details')}")
