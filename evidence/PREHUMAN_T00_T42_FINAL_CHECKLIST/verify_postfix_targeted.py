#!/usr/bin/env python3
import json
import pathlib

original = pathlib.Path('/home/ubuntu/forgelocal-final-check-20260824/PREHUMAN_FINAL_QUALITY_NORMALIZED.log').read_text()
line = next(item for item in original.splitlines() if item.startswith('golangci_new='))
targeted = json.loads(line.split('=', 1)[1])
postfix = json.loads(pathlib.Path('/home/ubuntu/forgelocal-postfix-final-check-20260824/POSTFIX_GOLANGCI.json').read_text())

def key(item):
    return (item.get('FromLinter'), item.get('Rule'), item.get('Pos', {}).get('Filename', ''), item.get('Text', ''))

current = {key(item) for item in postfix.get('Issues', [])}
remaining = []
for _, linter, filename, _line, text in targeted:
    matches = [row for row in current if row[0] == linter and row[2].endswith(filename) and row[3] == text]
    if matches:
        remaining.extend(matches)

print(f'original_targeted_count={len(targeted)}')
print(f'postfix_targeted_remaining_count={len(remaining)}')
print('postfix_targeted_remaining=' + json.dumps(sorted(remaining), ensure_ascii=False))
print('decision=' + ('ALL_13_TARGETED_FINDINGS_CLOSED' if not remaining else 'TARGETED_FINDINGS_REMAIN_OPEN'))
