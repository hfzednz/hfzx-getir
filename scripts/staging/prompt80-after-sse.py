#!/usr/bin/env python3
"""Prompt 80: pack/ready/courier + tracking after browser pick. No OTP digits printed."""
from __future__ import annotations

import json
import re
import subprocess
import time
import urllib.error
import urllib.request

TENANT = "11111111-1111-1111-1111-111111111111"
CUSTOMER = "http://127.0.0.1:3000"
WH = "http://127.0.0.1:8113"
CO = "http://127.0.0.1:8112"
IDN = "http://127.0.0.1:8081"

PHONES = {
    "customer": "+905551112233",
    "warehouse": "+905551112234",
    "courier": "+905551112235",
}


def call(base, method, path, body=None, token=None, timeout=30):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": TENANT, "Accept": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(base + path, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode()
            obj = json.loads(raw) if raw.lstrip()[:1] in "{[" else {}
            print("HTTP", resp.status, method, path)
            return resp.status, obj
    except urllib.error.HTTPError as e:
        raw = e.read().decode()[:400]
        print("HTTP", e.code, method, path, raw.replace("\n", " ")[:160])
        return e.code, {}


def otp_code(phone):
    logs = subprocess.check_output(
        ["bash", "-lc", f"docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep {phone} | tail -1"],
        text=True,
    )
    m = re.search(r'"code":"(\d+)"', logs)
    return m.group(1) if m else ""


def login_identity(phone):
    _, start = call(IDN, "POST", "/v1/identity/auth/otp/start", {"phone": phone, "tenantId": TENANT})
    time.sleep(0.3)
    _, sess = call(IDN, "POST", "/v1/identity/auth/otp/verify", {"challengeId": start.get("challengeId"), "code": otp_code(phone)})
    return sess.get("accessToken") or "", sess.get("principalId") or ""


def login_customer(phone):
    _, start = call(CUSTOMER, "POST", "/v1/customer/auth/otp/start", {"phone": phone})
    time.sleep(0.3)
    _, sess = call(CUSTOMER, "POST", "/v1/customer/auth/otp/verify", {"challengeId": start.get("challengeId"), "code": otp_code(phone)})
    return sess.get("accessToken") or "", sess.get("customerId") or sess.get("principalId") or ""


def main():
    oid = open("/tmp/prompt80-order-id.txt").read().strip()
    print("ORDER", oid)
    wh_tok, _ = login_identity(PHONES["warehouse"])
    co_tok, co_principal = login_identity(PHONES["courier"])
    cust_tok, _ = login_customer(PHONES["customer"])

    st, o = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("status_after_pick", o.get("status"))

    st_pack, pack = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/pack", {}, token=wh_tok)
    print("pack", st_pack, pack.get("status"))
    st, o = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("status_after_pack", o.get("status"))

    st_ready, ready = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/ready", {}, token=wh_tok)
    print("ready", st_ready, ready.get("status"))
    st, o = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("status_after_ready", o.get("status"))

    st_acc, acc = call(CO, "POST", f"/v1/courier/offers/{oid}", {"courierId": co_principal or "c1", "accept": True}, token=co_tok)
    print("accept", st_acc, acc.get("status"))
    st, o = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("status_after_accept", o.get("status"))

    st_en, en = call(CO, "POST", f"/v1/courier/offers/{oid}/enroute", {}, token=co_tok)
    print("enroute", st_en, en.get("status"))
    st, o = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("status_after_enroute", o.get("status"))

    st_done, done = call(CO, "POST", f"/v1/courier/offers/{oid}/complete", {}, token=co_tok)
    print("complete", st_done, done.get("status"))
    st, o = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("status_after_complete", o.get("status"))

    st_tr, tr = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=cust_tok)
    print("track_final", st_tr, tr.get("status"), tr.get("orderId"))

    tip = subprocess.check_output(
        ["docker", "inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", "nexora-staging-tracking-service"],
        text=True,
    ).strip()
    st_d, direct = call(f"http://{tip}:8080", "GET", f"/v1/tracking/orders/{oid}/timeline")
    print("tracking_service_direct", st_d, list(direct)[:8] if isinstance(direct, dict) else type(direct))

    ok = (
        st_pack == 200
        and st_ready == 200
        and st_acc == 200
        and st_en == 200
        and st_done == 200
        and st_tr == 200
        and bool(tr.get("status"))
    )
    print("PROMPT80_AFTER_SSE", "OK" if ok else "FAIL")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
