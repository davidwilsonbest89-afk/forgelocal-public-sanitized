#!/usr/bin/env python3
import json, pathlib, sys
out=pathlib.Path(sys.argv[1])
def findings(name):
    data=json.loads((out/name).read_text())
    result=[]
    for item in data.get('Issues',[]):
        file=str(item.get('file',''))
        marker='/forgelocal-prehuman-fresh-20260824/'
        if marker in file:
            file=file.split(marker,1)[1]
        for marker2 in ('/cmd/', '/internal/', '/pkg/'):
            if marker2 in file:
                file=file[file.index(marker2)+1:]
                break
        result.append((item.get('rule_id'),file,item.get('line'),item.get('details')))
    return sorted(result, key=lambda x: tuple('' if v is None else str(v) for v in x))
base=findings('PREHUMAN_FINAL_GOSEC_BASELINE.json')
head=findings('PREHUMAN_FINAL_GOSEC_HEAD.json')
new=sorted(set(head)-set(base))
resolved=sorted(set(base)-set(head))
print(f'baseline_count={len(base)}')
print(f'head_count={len(head)}')
print(f'new_count={len(new)}')
print(f'resolved_count={len(resolved)}')
print('new_findings='+json.dumps(new, ensure_ascii=False))
print('resolved_findings='+json.dumps(resolved, ensure_ascii=False))
