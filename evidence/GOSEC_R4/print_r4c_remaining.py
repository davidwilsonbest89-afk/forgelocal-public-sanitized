#!/usr/bin/env python3
import json
from pathlib import Path
p = Path(__file__).parent / 'gosec_after_r4c_bounded.json'
for issue in json.loads(p.read_text()).get('Issues', []):
    if issue.get('rule_id') in {'G101','G107','G115','G302','G404'}:
        print(issue.get('rule_id'), issue.get('file'), issue.get('line'), issue.get('details'))
