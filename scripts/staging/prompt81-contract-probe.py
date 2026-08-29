#!/usr/bin/env python3
"""Dump real response shapes from the phone-staging stack for contract auditing.

Prints top-level keys (and the first element's keys for lists) so DTO drift between
Go services and the web clients can be spotted without guessing. Never prints OTP
codes or tokens.
"""
from __future__ import annotations

import json
import re
import subprocess
import time
import urllib.error
import urllib.request

TENANT = "11111111-1111-1111-1111-111111111111"
CUSTOMER = "http://127.0.0.1:3000"
IDENTITY = "http://127.0.0.1:8081"
PHONE = "+905551112233"


def call(base, method, path, body=None, token=None, tenant=TENANT):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": tenant}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(base + path, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=25) as resp:
            raw = resp.read().decode()
            return resp.status, (json.loads(raw) if raw.lstrip()[:1] in "{[" else {})
    except urllib.error.HTTPError as e:
        return e.code, {}


def otp_code(phone):
    logs = subprocess.check_output(
        ["bash", "-lc", f"docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep {phone} | tail -1"],
        text=True,
    )
    m = re.search(r'"code":"(\d+)"', logs)
    return m.group(1) if m else ""


def login():
    _, start = call(CUSTOMER, "POST", "/v1/customer/auth/otp/start", {"phone": PHONE})
    time.sleep(0.3)
    st, sess = call(
        CUSTOMER,
        "POST",
        "/v1/customer/auth/otp/verify",
        {"challengeId": start.get("challengeId"), "code": otp_code(PHONE)},
    )
    print("login_http", st, "session_keys", sorted(sess.keys()))
    return sess.get("accessToken") or "", sess.get("customerId") or sess.get("principalId") or ""


def shape(label, status, obj):
    if isinstance(obj, dict):
        print(label, status, "keys", sorted(obj.keys()))
        for k, v in obj.items():
            if isinstance(v, list) and v and isinstance(v[0], dict):
                print("   ", label + "." + k + "[0] keys", sorted(v[0].keys()))
    elif isinstance(obj, list):
        print(label, status, "list len", len(obj), "first keys", sorted(obj[0].keys()) if obj and isinstance(obj[0], dict) else None)
    else:
        print(label, status, type(obj).__name__)


def main():
    token, principal = login()
    if not token:
        print("FAIL login")
        return 1

    st, home = call(CUSTOMER, "GET", f"/v1/customer/home?lat=41&lng=29&customerId={principal}", token=token)
    shape("home", st, home)
    products = home.get("products") or []
    if products:
        print("home.product[0]", json.dumps(products[0], sort_keys=True)[:400])

    st, search = call(CUSTOMER, "GET", "/v1/customer/home?lat=41&lng=29&q=milk", token=token)
    shape("search", st, search)

    print("PROMPT81_CONTRACT_PROBE_DONE")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
