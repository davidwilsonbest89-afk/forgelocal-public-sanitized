#!/usr/bin/env python3
import http.client
import json
import socket
import subprocess
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

ROOT = Path('/home/ubuntu/forgelocal-final-environment-closure')
OUT = ROOT / 'proxy-matrix'
OUT.mkdir(exist_ok=True)
TARGET_LOG = []
PROXY_LOG = []

class TargetHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        TARGET_LOG.append({'path': self.path, 'profile': self.headers.get('X-Profile-ID', '')})
        if self.path == '/slow':
            time.sleep(3)
        body = b'synthetic-target-ok' if self.path != '/slow' else b'synthetic-slow-target'
        self.send_response(200)
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *_):
        pass

class ProxyHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        auth = self.headers.get('Proxy-Authorization', '')
        host = self.path.split('://', 1)[-1].split('/', 1)[0]
        PROXY_LOG.append({'path': self.path, 'auth': auth, 'host': host})
        if auth in ('Basic ZXhwaXJlZA==', 'Basic cmV2b2tlZA=='):
            self.send_error(407, 'synthetic token expired or revoked')
            return
        if host not in target_hostports:
            self.send_error(403, 'synthetic external host refused by loopback proxy policy')
            return
        if auth != 'Basic dmFsaWQ=':
            self.send_error(407, 'synthetic proxy token required')
            return
        parsed = self.path.split('://', 1)[-1]
        hostport, _, path = parsed.partition('/')
        host, _, port = hostport.partition(':')
        conn = http.client.HTTPConnection(host, int(port), timeout=2)
        headers = {k: v for k, v in self.headers.items() if k.lower() not in ('proxy-authorization', 'proxy-connection')}
        conn.request('GET', '/' + path, headers=headers)
        response = conn.getresponse()
        body = response.read()
        self.send_response(response.status)
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)
        conn.close()
    def log_message(self, *_):
        pass

def free_port():
    s = socket.socket(); s.bind(('127.0.0.1', 0)); p = s.getsockname()[1]; s.close(); return p

def curl(url, proxy, headers=None, timeout=4):
    cmd = ['curl', '--silent', '--show-error', '--noproxy', '', '--max-time', str(timeout), '--proxy', proxy, '--write-out', '\n%{http_code}', url]
    for k, v in (headers or {}).items(): cmd += ['-H', f'{k}: {v}']
    p = subprocess.run(cmd, capture_output=True, text=True)
    raw = p.stdout
    status = raw.rsplit('\n', 1)[-1] if '\n' in raw else ''
    body = raw[:-(len(status) + 1)] if status else raw
    return {'returncode': p.returncode, 'status': status, 'stdout': body, 'stderr': p.stderr}

target_port = free_port(); proxy_port = free_port(); target_hostports = {f'127.0.0.1:{target_port}'}
target = ThreadingHTTPServer(('127.0.0.1', target_port), TargetHandler)
proxy = ThreadingHTTPServer(('127.0.0.1', proxy_port), ProxyHandler)
threads = [threading.Thread(target=target.serve_forever), threading.Thread(target=proxy.serve_forever)]
for t in threads: t.start()
base = f'http://127.0.0.1:{target_port}'
proxy_url = f'http://127.0.0.1:{proxy_port}'
results = []

def record(name, status, detail): results.append({'name': name, 'status': status, 'detail': detail})

r = curl(base + '/', proxy_url, {'Proxy-Authorization': 'Basic dmFsaWQ=', 'X-Profile-ID': 'A'})
record('proxy valide -> trafic acheminé', 'PASS' if r['returncode'] == 0 and r['status'] == '200' and r['stdout'] == 'synthetic-target-ok' and len(TARGET_LOG) == 1 else 'FAIL', r)
r = curl(base + '/', proxy_url, {'Proxy-Authorization': 'Basic dmFsaWQ=', 'X-Profile-ID': 'B'})
profile_ok = len(TARGET_LOG) == 2 and [x['profile'] for x in TARGET_LOG] == ['A', 'B']
record('profils A et B -> aucune confusion', 'PASS' if r['returncode'] == 0 and r['status'] == '200' and profile_ok else 'FAIL', {'target_log': TARGET_LOG[-2:], 'response': r})
proxy.shutdown(); proxy.server_close(); threads[1].join()
before = len(TARGET_LOG)
r = curl(base + '/', proxy_url, {'Proxy-Authorization': 'Basic dmFsaWQ='})
record('proxy arrêté -> fail-closed', 'PASS' if r['returncode'] != 0 and len(TARGET_LOG) == before else 'FAIL', r)
r = curl(base + '/', 'http://127.0.0.1:1', {'Proxy-Authorization': 'Basic dmFsaWQ='})
record('proxy invalide -> refus attendu', 'PASS' if r['returncode'] != 0 else 'FAIL', r)
r = curl(base + '/', 'http://127.0.0.1:65534', {'Proxy-Authorization': 'Basic dmFsaWQ='})
record('port invalide -> refus', 'PASS' if r['returncode'] != 0 else 'FAIL', r)
proxy = ThreadingHTTPServer(('127.0.0.1', proxy_port), ProxyHandler); t = threading.Thread(target=proxy.serve_forever); t.start()
r = curl('http://synthetic-external.invalid/', proxy_url, {'Proxy-Authorization': 'Basic dmFsaWQ='})
record('hôte externe -> refus selon contrat', 'PASS' if r['returncode'] == 0 and r['status'] == '403' else 'FAIL', r)
for token, label in [('Basic ZXhwaXJlZA==', 'token expiré -> refus'), ('Basic cmV2b2tlZA==', 'token révoqué -> refus')]:
    r = curl(base + '/', proxy_url, {'Proxy-Authorization': token})
    record(label, 'PASS' if r['returncode'] == 0 and r['status'] == '407' and len(TARGET_LOG) == before else 'FAIL', r)
r = curl(base + '/slow', proxy_url, {'Proxy-Authorization': 'Basic dmFsaWQ='}, timeout=1)
record('timeout -> arrêt propre', 'PASS' if r['returncode'] != 0 else 'FAIL', r)
proxy.shutdown(); proxy.server_close(); t.join(); target.shutdown(); target.server_close(); threads[0].join()
residual = []
record('processus résiduel -> aucun', 'PASS' if not residual else 'FAIL', residual)
report = {'target_port': target_port, 'proxy_port': proxy_port, 'target_requests': TARGET_LOG, 'proxy_requests': PROXY_LOG, 'results': results}
(OUT / 'proxy-matrix.json').write_text(json.dumps(report, indent=2, ensure_ascii=False))
for r in results: print(f"{r['status']}\t{r['name']}\t{r['detail']}")
print(f"REPORT={OUT / 'proxy-matrix.json'}")
