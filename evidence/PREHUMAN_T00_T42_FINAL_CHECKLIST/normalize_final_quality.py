#!/usr/bin/env python3
import json
import pathlib

out = pathlib.Path('/home/ubuntu/forgelocal-final-check-20260824')

def relpath(value):
    value = str(value or '')
    for marker in ('/cmd/', '/internal/', '/pkg/', '/forge-dashboard/'):
        index = value.find(marker)
        if index >= 0:
            return value[index + 1:]
    return value

def load_static(path):
    rows = []
    for line in pathlib.Path(path).read_text(errors='replace').splitlines():
        if not line.strip():
            continue
        try:
            item = json.loads(line)
        except Exception:
            continue
        location = item.get('location') or {}
        rows.append(('staticcheck', item.get('code', ''), relpath(location.get('file')), location.get('line', 0), item.get('message', '')))
    return rows

def load_golangci(path):
    payload = json.loads(pathlib.Path(path).read_text())
    rows = []
    for item in payload.get('Issues') or []:
        position = item.get('Pos') or {}
        rows.append(('golangci', item.get('FromLinter', ''), relpath(position.get('Filename')), position.get('Line', 0), item.get('Text', '')))
    return rows

for name, loader, extension in (
    ('staticcheck', load_static, 'jsonl'),
    ('golangci', load_golangci, 'json'),
):
    baseline = loader(out / f'PREHUMAN_FINAL_{name.upper()}_BASELINE.{extension}')
    head = loader(out / f'PREHUMAN_FINAL_{name.upper()}_HEAD.{extension}')
    baseline_set = set(baseline)
    head_set = set(head)
    print(f'{name}_baseline_count={len(baseline)}')
    print(f'{name}_head_count={len(head)}')
    print(f'{name}_new_count={len(head_set - baseline_set)}')
    print(f'{name}_resolved_count={len(baseline_set - head_set)}')
    print(f'{name}_new={json.dumps(sorted(head_set - baseline_set), ensure_ascii=False)}')
    print(f'{name}_resolved={json.dumps(sorted(baseline_set - head_set), ensure_ascii=False)}')
