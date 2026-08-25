#!/usr/bin/env python3
import json
from pathlib import Path
p=Path('/home/ubuntu/forgelocal_self_validation_v6/V6_AXE_RESULTS_AFTER.json')
d=json.loads(p.read_text())
print('passes',len(d.get('passes',[])))
print('incomplete',len(d.get('incomplete',[])))
print('violations',len(d.get('violations',[])))
for v in d.get('violations',[]):
    print(v.get('id'),v.get('impact'),len(v.get('nodes',[])))
