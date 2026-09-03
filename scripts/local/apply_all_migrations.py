#!/usr/bin/env python3
"""Apply every services/*/migrations/*.sql onto the local compose Postgres."""
from __future__ import annotations

import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
COMPOSE = ["docker", "compose", "-f", str(ROOT / "infra" / "docker" / "docker-compose.yml")]


def run(args, **kw):
    return subprocess.run(args, capture_output=True, text=True, **kw)


def main() -> int:
    print("wait postgres", flush=True)
    for _ in range(40):
        r = run(COMPOSE + ["exec", "-T", "postgres", "pg_isready", "-U", "nexora", "-d", "nexora"])
        if r.returncode == 0:
            print("postgres ready", flush=True)
            break
        time.sleep(2)
    else:
        print("FAIL postgres not ready")
        return 1
    r = run(COMPOSE + ["exec", "-T", "redis", "redis-cli", "PING"])
    print("redis", (r.stdout or "").strip(), flush=True)
    fails = []
    ok = 0
    for mig in sorted((ROOT / "services").glob("*/migrations")):
        svc = mig.parent.name
        db = "nexora_" + svc.replace("-", "_")
        print("==>", svc, "->", db, flush=True)
        exists = run(
            COMPOSE
            + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", "nexora", "-Atc",
               "SELECT 1 FROM pg_database WHERE datname='%s'" % db]
        )
        if "1" not in (exists.stdout or ""):
            cr = run(
                COMPOSE
                + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", "nexora", "-v", "ON_ERROR_STOP=1",
                   "-c", 'CREATE DATABASE "%s";' % db]
            )
            if cr.returncode != 0:
                print("FAIL create", db, cr.stderr or cr.stdout)
                fails.append(svc + ":create")
                continue
        sqls = sorted(mig.glob("*.sql"))
        if not sqls:
            print("SKIP no sql", svc)
            continue
        bad = False
        for sql in sqls:
            r = run(
                COMPOSE + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", db, "-v", "ON_ERROR_STOP=1"],
                input=sql.read_text(encoding="utf-8"),
            )
            if r.returncode != 0:
                print("FAIL", sql.name, (r.stderr or r.stdout)[:500])
                fails.append("%s:%s" % (svc, sql.name))
                bad = True
                break
            print("  OK", sql.name, flush=True)
        if not bad:
            ok += 1
    print("MIG_OK", ok, "MIG_FAIL", len(fails), flush=True)
    for f in fails:
        print(" -", f)
    if fails:
        return 1
    print("MIGRATION_SMOKE_OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
