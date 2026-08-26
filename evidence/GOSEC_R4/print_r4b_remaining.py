#!/usr/bin/env python3
import json
from pathlib import Path
p = Path(__file__).parent / 'gosec_after_r4b_root.json'
for issue in json.loads(p.read_text()).get('Issues', []):
    if issue.get('rule_id') in {'G304','G305','G703'}:
        print(issue.get('rule_id'), issue.get('file'), issue.get('line'), issue.get('details'))
