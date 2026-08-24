#!/usr/bin/env python3
import json, pathlib, sys
base=json.loads(pathlib.Path(sys.argv[1]).read_text())
head=json.loads(pathlib.Path(sys.argv[2]).read_text())
def collect(data):
 out=[]
 for r in data.get('Results',[]):
  target=r.get('Target','')
  for k in ('Vulnerabilities','Secrets','Misconfigurations'):
   for x in r.get(k) or []:
    out.append((target,k,x.get('VulnerabilityID') or x.get('ID') or x.get('RuleID'),x.get('PkgName') or x.get('Title',''),x.get('Severity','')))
 return sorted(out)
b=collect(base); h=collect(head)
print('baseline_findings=%d' % len(b))
print('head_findings=%d' % len(h))
print('new_findings=%d' % len(sorted(set(h)-set(b))))
print('resolved_findings=%d' % len(sorted(set(b)-set(h))))
print('new='+json.dumps(sorted(set(h)-set(b)),ensure_ascii=False))
print('resolved='+json.dumps(sorted(set(b)-set(h)),ensure_ascii=False))
