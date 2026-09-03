#!/usr/bin/env python3
"""Prompt 87: platform-ops Postgres migrate + restart persistence check."""
from __future__ import annotations

import importlib.util
import json
import os
import subprocess
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
RUN = ROOT / "tmp" / "prompt87-stack"
COMPOSE = ROOT / "infra" / "docker" / "docker-compose.yml"
TENANT = "11111111-1111-1111-1111-111111111111"
DB = "nexora_platform_ops_service"
DSN = f"postgres://nexora:nexora@127.0.0.1:5432/{DB}?sslmode=disable"


def run(cmd, **kw):
    return subprocess.run(cmd, capture_output=True, text=True, **kw)


def http_json(method, url, body=None, token=None):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": TENANT, "X-Request-Id": "p87-pg"}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(url, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            raw = resp.read().decode("utf-8", "replace")
            obj = json.loads(raw) if raw.strip().startswith(("{", "[")) else {}
            return resp.status, obj, raw
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", "replace")
        try:
            obj = json.loads(raw)
        except json.JSONDecodeError:
            obj = {}
        return e.code, obj, raw


def wait_health(port=8110, seconds=30):
    deadline = time.time() + seconds
    while time.time() < deadline:
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{port}/health", timeout=2)
            return True
        except Exception:
            time.sleep(0.5)
    return False


def kill_port(port=8110):
    pidf = RUN / "platform-ops-service.pid"
    if pidf.exists():
        subprocess.run(["taskkill", "/PID", pidf.read_text().strip(), "/F", "/T"], capture_output=True)
        pidf.unlink(missing_ok=True)
    r = run(
        [
            "powershell",
            "-NoProfile",
            "-Command",
            f"(Get-NetTCPConnection -LocalPort {port} -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1 -ExpandProperty OwningProcess)",
        ]
    )
    pid = (r.stdout or "").strip()
    if pid.isdigit():
        subprocess.run(["taskkill", "/PID", pid, "/F", "/T"], capture_output=True)


def start_platform_ops(binary: Path):
    env = os.environ.copy()
    env.update(
        {
            "GOWORK": "off",
            "GOTOOLCHAIN": "auto",
            "OTP_DEV_MODE": "true",
            "RATE_LIMIT_PER_MINUTE": "0",
            "DATABASE_URL": DSN,
            "REDIS_URL": "",
            "KAFKA_BROKERS": "",
            "IDENTITY_URL": "http://127.0.0.1:8081",
            "HTTP_ADDR": ":8110",
        }
    )
    logf = open(RUN / "platform-ops-service.log", "a", encoding="utf-8")
    p = subprocess.Popen(
        [str(binary)],
        cwd=str(ROOT / "services" / "platform-ops-service"),
        env=env,
        stdout=logf,
        stderr=subprocess.STDOUT,
    )
    (RUN / "platform-ops-service.pid").write_text(str(p.pid), encoding="utf-8")
    return p


def main():
    RUN.mkdir(parents=True, exist_ok=True)
    compose = ["docker", "compose", "-f", str(COMPOSE)]
    print("== migrate platform-ops", flush=True)
    for _ in range(30):
        ping = run(compose + ["exec", "-T", "postgres", "pg_isready", "-U", "nexora", "-d", "nexora"])
        if ping.returncode == 0:
            break
        time.sleep(2)
    else:
        raise SystemExit("postgres not ready")

    run(compose + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", "nexora", "-c", f'DROP DATABASE IF EXISTS "{DB}" WITH (FORCE);'])
    r = run(
        compose
        + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", "nexora", "-v", "ON_ERROR_STOP=1", "-c", f'CREATE DATABASE "{DB}";']
    )
    if r.returncode != 0:
        raise SystemExit("create db failed: " + (r.stderr or r.stdout)[:400])

    for sql in sorted((ROOT / "services" / "platform-ops-service" / "migrations").glob("*.sql")):
        r = run(
            compose + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", DB, "-v", "ON_ERROR_STOP=1"],
            input=sql.read_text(encoding="utf-8"),
        )
        print("migrate", sql.name, "rc", r.returncode, flush=True)
        if r.returncode != 0:
            raise SystemExit(f"migrate {sql.name}: {(r.stderr or r.stdout)[:500]}")

    binary = RUN / ("platform-ops-service.exe" if os.name == "nt" else "platform-ops-service")
    cmd_dir = ROOT / "services" / "platform-ops-service" / "cmd" / "platform-ops-service"
    env = os.environ.copy()
    env["GOWORK"] = "off"
    env["GOTOOLCHAIN"] = "auto"
    print("== build platform-ops", flush=True)
    r = run(["go", "build", "-o", str(binary), "."], cwd=str(cmd_dir), env=env)
    if r.returncode != 0:
        raise SystemExit(r.stderr or r.stdout)

    kill_port(8110)
    time.sleep(1)
    start_platform_ops(binary)
    if not wait_health():
        raise SystemExit("platform-ops not healthy with DATABASE_URL")
    print("platform-ops up with DATABASE_URL", flush=True)

    spec = importlib.util.spec_from_file_location("gate", ROOT / "scripts" / "local" / "prompt87_live_gate.py")
    gate = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(gate)
    sa_tok, _ = gate.login("super_admin")

    slug = "p87-" + uuid.uuid4().hex[:8]
    st, tenant_body, raw = http_json(
        "POST",
        "http://127.0.0.1:8110/v1/platform/tenants",
        {"name": "P87 Persist Tenant", "slug": slug, "primaryCurrency": "TRY"},
        token=sa_tok,
    )
    print("create tenant", st, raw[:300], flush=True)
    if st not in (200, 201):
        raise SystemExit("create tenant failed")
    tid = tenant_body.get("id") or (tenant_body.get("tenant") or {}).get("id")

    st, company, raw = http_json(
        "POST",
        "http://127.0.0.1:8110/v1/platform/companies",
        {"legalName": "P87 Persist Co", "tradeName": "PersistCo", "countryCode": "TR", "primaryCurrency": "TRY"},
        token=sa_tok,
    )
    print("create company", st, raw[:300], flush=True)
    if st not in (200, 201):
        raise SystemExit("create company failed")
    cid = company.get("id") or (company.get("company") or {}).get("id")

    st, prop, raw = http_json(
        "POST",
        "http://127.0.0.1:8110/v1/platform/tenants/dual-control",
        {"action": "tenant_suspend", "tenantId": tid, "reason": "p87"},
        token=sa_tok,
    )
    print("dual-control", st, raw[:300], flush=True)
    if st not in (200, 201):
        raise SystemExit("dual-control propose failed")
    prop_id = prop.get("id") or (prop.get("proposal") or {}).get("id") or ""
    if not prop_id:
        raise SystemExit("dual-control proposal id missing")

    st, audit, raw = http_json("GET", "http://127.0.0.1:8110/v1/platform/audit", token=sa_tok)
    print("audit before restart", st, flush=True)
    if st != 200:
        raise SystemExit("audit before restart failed")

    print("== restart platform-ops", flush=True)
    kill_port(8110)
    time.sleep(1)
    start_platform_ops(binary)
    if not wait_health():
        raise SystemExit("platform-ops not healthy after restart")

    st, _, raw = http_json("GET", f"http://127.0.0.1:8110/v1/platform/tenants/{tid}", token=sa_tok)
    print("get tenant after restart", st, raw[:250], flush=True)
    if st != 200:
        raise SystemExit("tenant not persisted")
    st, _, raw = http_json("GET", f"http://127.0.0.1:8110/v1/platform/companies/{cid}", token=sa_tok)
    print("get company after restart", st, raw[:250], flush=True)
    if st != 200:
        raise SystemExit("company not persisted")
    if prop_id:
        # No dedicated GET; resolve after restart proves the proposal row persisted.
        st, _, raw = http_json(
            "POST",
            f"http://127.0.0.1:8110/v1/platform/tenants/dual-control/{prop_id}",
            {"decision": "approve", "approverId": "p87-approver"},
            token=sa_tok,
        )
        print("resolve dual-control after restart", st, raw[:250], flush=True)
        if st not in (200, 201, 409):
            raise SystemExit("dual-control proposal not persisted")
    st, _, raw = http_json("GET", "http://127.0.0.1:8110/v1/platform/audit", token=sa_tok)
    print("audit after restart", st, raw[:250], flush=True)
    if st != 200:
        raise SystemExit("audit after restart failed")
    print("PLATFORM_OPS_PERSIST_PASS", flush=True)


if __name__ == "__main__":
    main()
