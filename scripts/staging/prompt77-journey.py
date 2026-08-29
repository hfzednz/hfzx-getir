#!/usr/bin/env python3
"""Prompt 77: real order + warehouse/courier lifecycle + live SSE (not GET 200)."""
from __future__ import annotations

import json
import os
import re
import subprocess
import threading
import time
import urllib.error
import urllib.request
import uuid

TENANT = "11111111-1111-1111-1111-111111111111"
WRONG_TENANT = "22222222-2222-2222-2222-222222222222"
CUSTOMER = "http://127.0.0.1:3000"
WAREHOUSE = "http://127.0.0.1:8113"
COURIER = "http://127.0.0.1:8112"
ADMIN = "http://127.0.0.1:8114"
FINANCE = "http://127.0.0.1:8091"
SUPPLIER = "http://127.0.0.1:8117"
REALTIME = "http://127.0.0.1:8115"
IDENTITY = "http://127.0.0.1:8081"
PHONE = "+905551112233"


def call(base, method, path, body=None, token=None, tenant=TENANT, timeout=30):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": tenant, "Accept": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(base + path, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode()
            obj = json.loads(raw) if raw.lstrip().startswith("{") else {}
            print(method, base.split("://", 1)[-1], path, resp.status)
            return resp.status, obj, raw
    except urllib.error.HTTPError as e:
        raw = e.read().decode()[:800]
        print(method, base.split("://", 1)[-1], path, "FAIL", e.code, raw.replace("\n", " ")[:240])
        return e.code, {}, raw
    except Exception as e:
        print(method, base.split("://", 1)[-1], path, "ERR", type(e).__name__, e)
        return 0, {}, str(e)


def otp_code():
    logs = subprocess.check_output(
        [
            "bash",
            "-lc",
            "docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep +905551112233 | tail -1",
        ],
        text=True,
    )
    m = re.search(r'"code":"(\d+)"', logs)
    return m.group(1) if m else ""


def login_customer():
    st, start, _ = call(CUSTOMER, "POST", "/v1/customer/auth/otp/start", {"phone": PHONE})
    cid = start.get("challengeId")
    code = otp_code()
    st, sess, _ = call(
        CUSTOMER,
        "POST",
        "/v1/customer/auth/otp/verify",
        {"challengeId": cid, "code": code},
    )
    token = sess.get("accessToken") or sess.get("AccessToken") or ""
    principal = sess.get("customerId") or sess.get("principalId") or sess.get("PrincipalID") or ""
    print("customer_otp", bool(token), "principal", bool(principal), "otp_len", len(code), "start", bool(cid))
    return token, principal


def login_identity():
    st, start, _ = call(
        IDENTITY,
        "POST",
        "/v1/identity/auth/otp/start",
        {"phone": PHONE, "tenantId": TENANT},
    )
    cid = start.get("challengeId")
    code = otp_code()
    st, sess, _ = call(
        IDENTITY,
        "POST",
        "/v1/identity/auth/otp/verify",
        {"challengeId": cid, "code": code},
    )
    token = sess.get("accessToken") or sess.get("AccessToken") or ""
    print("identity_otp", bool(token), "start", bool(cid), "otp_len", len(code))
    return token


def ensure_order(token, principal):
    oid = os.environ.get("PROMPT77_ORDER_ID", "").strip()
    if oid:
        st, order, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token)
        if st == 200:
            print("reuse_order", oid, order.get("status"))
            return oid, order
        print("reuse_order_miss", oid, st)
    guest = "web-" + uuid.uuid4().hex
    raw = subprocess.check_output(
        [
            "docker",
            "run",
            "--rm",
            "--network",
            "nexora-phone-staging",
            "curlimages/curl:8.5.0",
            "-sS",
            "-H",
            "Content-Type: application/json",
            "-H",
            "X-Tenant-Id: " + TENANT,
            "-X",
            "POST",
            "http://cart-service:8080/v1/cart",
            "-d",
            json.dumps({"guestToken": guest, "currency": "TRY"}),
        ],
        text=True,
    )
    created = json.loads(raw)
    cart_id = created.get("ID") or created.get("id") or created.get("cartId") or ""
    st, home, _ = call(CUSTOMER, "GET", "/v1/customer/home?lat=41&lng=29", token=token)
    products = home.get("products") or []
    sku = ""
    if products:
        p0 = products[0]
        sku = str(p0.get("id") or p0.get("productId") or p0.get("sku") or "")
    sku = os.environ.get("PROMPT77_SKU", sku)
    print("place_sku", sku, "cart", cart_id, "home_products", len(products))
    call(
        CUSTOMER,
        "POST",
        "/v1/customer/cart/items",
        {"cartId": cart_id, "sku": sku, "qty": 1, "unitMinor": 1500},
        token=token,
    )
    st, prev, _ = call(
        CUSTOMER,
        "POST",
        "/v1/customer/checkout/preview",
        {"cartId": cart_id, "principalId": principal},
        token=token,
    )
    sid = prev.get("sessionId") or ""
    st, placed, _ = call(
        CUSTOMER,
        "POST",
        "/v1/customer/checkout/place",
        {
            "cartId": cart_id,
            "paymentMethod": "card",
            "sessionId": sid,
            "principalId": principal,
            "address": {
                "label": "Istanbul",
                "line1": "Istanbul",
                "city": "Istanbul",
                "country": "TR",
                "lat": 41.0082,
                "lng": 28.9784,
            },
        },
        token=token,
    )
    oid = placed.get("orderId") or ""
    st, order, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token) if oid else (0, {}, "")
    print("placed", oid, "status", order.get("status"), "place_http", st)
    return oid, order


class SseListener:
    def __init__(self, url):
        self.url = url
        self.chunks = []
        self.opened = False
        self.error = ""
        self._stop = False

    def start(self):
        t = threading.Thread(target=self._run, daemon=True)
        t.start()
        return t

    def _run(self):
        req = urllib.request.Request(self.url, headers={"Accept": "text/event-stream"})
        try:
            with urllib.request.urlopen(req, timeout=25) as resp:
                self.opened = True
                resp.fp._sock.settimeout(20)
                while not self._stop:
                    line = resp.readline()
                    if not line:
                        break
                    self.chunks.append(line.decode("utf-8", "replace"))
        except Exception as e:
            self.error = f"{type(e).__name__}: {e}"

    def text(self):
        return "".join(self.chunks)


def main():
    print("== health")
    for base, path in [
        (CUSTOMER, "/health"),
        (WAREHOUSE, "/health"),
        (COURIER, "/health"),
        (ADMIN, "/health"),
        (FINANCE, "/health"),
        (SUPPLIER, "/health"),
        (REALTIME, "/health"),
        (IDENTITY, "/health"),
        ("http://127.0.0.1:8110", "/health"),
    ]:
        call(base, "GET", path)

    token, principal = login_customer()
    staff = login_identity()
    oid, order = ensure_order(token, principal)
    if not oid:
        print("FAIL no order")
        return 1

    call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=token)

    sse_gw = SseListener(f"{REALTIME}/v1/realtime/sse?topic=order:{oid}")
    sse_web = SseListener(f"{CUSTOMER}/v1/realtime/sse?topic=order:{oid}")
    sse_gw.start()
    sse_web.start()
    time.sleep(1.5)
    print("sse_open_gateway", sse_gw.opened, "sse_open_web", sse_web.opened)

    print("== warehouse lifecycle")
    call(WAREHOUSE, "POST", f"/v1/warehouse/tasks/{oid}/pick", {})
    call(WAREHOUSE, "POST", f"/v1/warehouse/tasks/{oid}/pack", {})
    call(WAREHOUSE, "POST", f"/v1/warehouse/tasks/{oid}/ready", {})
    print("== courier lifecycle")
    call(COURIER, "POST", f"/v1/courier/offers/{oid}", {"courierId": principal or "courier-1", "accept": True})
    call(COURIER, "POST", f"/v1/courier/offers/{oid}/enroute", {})

    time.sleep(2)
    gw_txt = sse_gw.text()
    web_txt = sse_web.text()
    sse_event = ("data:" in gw_txt) or ("data:" in web_txt)
    print("sse_gateway_bytes", len(gw_txt), "sse_web_bytes", len(web_txt))
    print("sse_gateway_open", sse_gw.opened, "err", sse_gw.error)
    print("sse_web_open", sse_web.opened, "err", sse_web.error)
    print("sse_event_received", sse_event)
    if gw_txt:
        print("sse_gateway_sample", gw_txt[:300].replace("\n", " | "))
    if web_txt:
        print("sse_web_sample", web_txt[:300].replace("\n", " | "))

    st, order2, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token)
    st, track2, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=token)
    print("order_after", order2.get("status"), "track_after", track2.get("status"))

    print("== admin/support/ops/finance/supplier")
    call(ADMIN, "GET", "/v1/admin/dashboard")
    call(ADMIN, "GET", f"/v1/admin/orders/{oid}")
    call(ADMIN, "GET", "/v1/admin/orders")
    call(FINANCE, "GET", "/v1/ledger/journals")
    call(SUPPLIER, "GET", "/v1/supplier/suppliers")
    call(CUSTOMER, "GET", "/v1/admin/dashboard", token=token)
    call("http://127.0.0.1:3002", "GET", "/login")
    call("http://127.0.0.1:3001", "GET", "/login")
    call("http://127.0.0.1:3003", "GET", "/login")
    call("http://127.0.0.1:3004", "GET", "/login")
    call("http://127.0.0.1:3005", "GET", "/login")
    call("http://127.0.0.1:3006", "GET", "/login")
    call("http://127.0.0.1:3100", "GET", "/login")
    call("http://127.0.0.1:3200", "GET", "/login")

    print("== tenant isolation")
    st_wrong, _, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token, tenant=WRONG_TENANT)
    print("wrong_tenant_get_order", st_wrong)

    print("== rbac notes")
    st_unauth, _, _ = call(ADMIN, "GET", "/v1/admin/orders")
    print("admin_orders_without_jwt", st_unauth, "(dest-mode bff-admin is tenant-header only)")
    print("staff_token_present", bool(staff))

    print("ORDER_ID", oid)
    print("SSE_PASS" if sse_event else "SSE_FAIL")
    return 0 if oid and sse_event else 1


if __name__ == "__main__":
    raise SystemExit(main())
