#!/usr/bin/env python3
"""Prompt 80: fresh post-reboot customer → order; RBAC; tenant; SSE tickets; role APIs."""
from __future__ import annotations

import json
import os
import re
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

TENANT = "11111111-1111-1111-1111-111111111111"
WRONG = "22222222-2222-2222-2222-222222222222"
CUSTOMER = "http://127.0.0.1:3000"
WH = "http://127.0.0.1:8113"
CO = "http://127.0.0.1:8112"
ADMIN = "http://127.0.0.1:8114"
FIN = "http://127.0.0.1:8091"
SUP = "http://127.0.0.1:8117"
RT = "http://127.0.0.1:8115"
IDN = "http://127.0.0.1:8081"
BFF = "http://127.0.0.1:8111"
PLATFORM = "http://127.0.0.1:8110"
TRACK = None  # filled from docker inspect

PHONES = {
    "customer": "+905551112233",
    "warehouse": "+905551112234",
    "courier": "+905551112235",
    "finance": "+905551112236",
    "support": "+905551112237",
    "ops": "+905551112238",
    "admin": "+905551112239",
    "supplier": "+905551112240",
    "super": "+905551112241",
    "customer_b": "+905551119999",
}


def call(base, method, path, body=None, token=None, tenant=TENANT, timeout=30):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": tenant, "Accept": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(base + path, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode()
            obj = json.loads(raw) if raw.lstrip()[:1] in "{[" else {}
            print("HTTP", resp.status, method, base.split("://", 1)[-1] + path)
            return resp.status, obj, raw
    except urllib.error.HTTPError as e:
        raw = e.read().decode()[:500]
        print("HTTP", e.code, method, base.split("://", 1)[-1] + path, raw.replace("\n", " ")[:160])
        return e.code, {}, raw
    except Exception as e:
        print("ERR", method, base + path, type(e).__name__, e)
        return 0, {}, str(e)


def otp_code(phone):
    logs = subprocess.check_output(
        ["bash", "-lc", f"docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep {phone} | tail -1"],
        text=True,
    )
    m = re.search(r'"code":"(\d+)"', logs)
    return m.group(1) if m else ""


def jwt_roles(token: str):
    try:
        payload = token.split(".")[1]
        pad = "=" * (-len(payload) % 4)
        import base64

        data = json.loads(base64.urlsafe_b64decode(payload + pad))
        return data.get("roles") or []
    except Exception:
        return []


def login_identity(phone):
    st, start, _ = call(IDN, "POST", "/v1/identity/auth/otp/start", {"phone": phone, "tenantId": TENANT})
    time.sleep(0.3)
    code = otp_code(phone)
    st, sess, _ = call(IDN, "POST", "/v1/identity/auth/otp/verify", {"challengeId": start.get("challengeId"), "code": code})
    token = sess.get("accessToken") or ""
    print("login", phone[-4:], "http", st, "roles", jwt_roles(token))
    return token, sess.get("principalId") or ""


def login_customer(phone):
    st, start, _ = call(CUSTOMER, "POST", "/v1/customer/auth/otp/start", {"phone": phone})
    time.sleep(0.3)
    code = otp_code(phone)
    st, sess, _ = call(CUSTOMER, "POST", "/v1/customer/auth/otp/verify", {"challengeId": start.get("challengeId"), "code": code})
    token = sess.get("accessToken") or ""
    print("customer_login", phone[-4:], "http", st, "roles", jwt_roles(token))
    return token, sess.get("customerId") or sess.get("principalId") or ""


def curl_sse(url):
    try:
        out = subprocess.check_output(
            ["curl", "-sS", "-N", "--max-time", "2", "-D", "-", "-o", "/tmp/p80-sse.out", url],
            text=True,
            stderr=subprocess.STDOUT,
        )
    except subprocess.CalledProcessError as e:
        out = (e.output or "") + (e.stderr or "")
    m = re.search(r"HTTP/\S+\s+(\d+)", out)
    code = int(m.group(1)) if m else 0
    body = ""
    try:
        body = open("/tmp/p80-sse.out", encoding="utf-8", errors="replace").read()[:80]
    except Exception:
        pass
    return code, body


def tracking_ip():
    return subprocess.check_output(
        ["docker", "inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", "nexora-staging-tracking-service"],
        text=True,
    ).strip()


def main():
    st, _, _ = call(CUSTOMER, "GET", "/login")
    if st != 200:
        print("FAIL customer login page", st)
        return 1
    for base, name in ((BFF, "bff"), (RT, "realtime"), (WH, "warehouse_bff"), (CO, "courier_bff")):
        hs, _, _ = call(base, "GET", "/health")
        print("health", name, hs)

    cust_tok, principal = login_customer(PHONES["customer"])
    if not cust_tok:
        print("FAIL customer token")
        return 1

    st, home, _ = call(CUSTOMER, "GET", "/v1/customer/home?lat=41&lng=29", token=cust_tok)
    products = home.get("products") or []
    sku = str((products[0].get("id") or products[0].get("sku")) if products else "")
    print("catalog_sku", bool(sku), "product_count", len(products))
    if not sku:
        print("FAIL no catalog")
        return 1

    guest = "web-" + uuid.uuid4().hex
    raw = subprocess.check_output(
        [
            "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
            "-sS", "-H", "Content-Type: application/json", "-H", "X-Tenant-Id: " + TENANT,
            "-X", "POST", "http://cart-service:8080/v1/cart",
            "-d", json.dumps({"guestToken": guest, "currency": "TRY"}),
        ],
        text=True,
    )
    cart = json.loads(raw)
    cart_id = cart.get("ID") or cart.get("id") or cart.get("cartId")
    st_cart, _, _ = call(CUSTOMER, "POST", "/v1/customer/cart/items", {"cartId": cart_id, "sku": sku, "qty": 1, "unitMinor": 1500}, token=cust_tok)
    st_prev, prev, _ = call(CUSTOMER, "POST", "/v1/customer/checkout/preview", {"cartId": cart_id, "principalId": principal}, token=cust_tok)
    sid = prev.get("sessionId") or ""
    print("cart_http", st_cart, "preview_http", st_prev, "session", bool(sid))
    st, placed, _ = call(
        CUSTOMER,
        "POST",
        "/v1/customer/checkout/place",
        {
            "cartId": cart_id,
            "paymentMethod": "card",
            "sessionId": sid,
            "principalId": principal,
            "address": {"label": "Istanbul", "line1": "Istanbul", "city": "Istanbul", "country": "TR", "lat": 41.0082, "lng": 28.9784},
        },
        token=cust_tok,
    )
    oid = placed.get("orderId") or ""
    print("ORDER_ID", oid, "place_http", st)
    if st != 201 or not oid:
        print("FAIL place", st, placed)
        return 1
    open("/tmp/prompt80-order-id.txt", "w").write(oid)
    open("/tmp/prompt80-sse-order.txt", "w").write(oid)

    st_get, order, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)
    print("order_get", st_get, "status", order.get("status") or order.get("Status"))
    if st_get not in (200, 201):
        print("FAIL get order")
        return 1

    tip = tracking_ip()
    st_tr_direct, tr_direct, _ = call(f"http://{tip}:8080", "GET", f"/v1/tracking/orders/{oid}/timeline", token=None)
    st_track, track, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=cust_tok)
    print("tracking_direct", st_tr_direct, "bff_track", st_track, "track_status", track.get("status"))

    print("== rbac denials")
    denials = {}
    wh_tok, _ = login_identity(PHONES["warehouse"])
    co_tok, co_principal = login_identity(PHONES["courier"])
    fin_tok, _ = login_identity(PHONES["finance"])
    sup_tok, _ = login_identity(PHONES["supplier"])
    support_tok, _ = login_identity(PHONES["support"])
    ops_tok, _ = login_identity(PHONES["ops"])
    adm_tok, _ = login_identity(PHONES["admin"])
    super_tok, _ = login_identity(PHONES["super"])
    denials["customer_admin"] = call(ADMIN, "GET", "/v1/admin/dashboard", token=cust_tok)[0]
    denials["customer_finance"] = call(FIN, "GET", "/v1/ledger/journals", token=cust_tok)[0]
    denials["courier_finance"] = call(FIN, "GET", "/v1/ledger/journals", token=co_tok)[0]
    denials["supplier_admin"] = call(ADMIN, "GET", "/v1/admin/dashboard", token=sup_tok)[0]
    denials["support_super"] = call(PLATFORM, "GET", "/v1/platform/admin/stats", token=support_tok)[0]
    denials["finance_admin"] = call(ADMIN, "GET", "/v1/admin/dashboard", token=fin_tok)[0]
    denials["warehouse_finance"] = call(FIN, "GET", "/v1/ledger/journals", token=wh_tok)[0]
    print("DENIALS", denials)
    for k, v in denials.items():
        if v not in (401, 403):
            print("FAIL denial", k, v)

    print("== legitimate APIs (no warehouse/courier mutation yet)")
    legit = {
        "customer_home": call(CUSTOMER, "GET", "/v1/customer/home?lat=41&lng=29", token=cust_tok)[0],
        "finance_journals": call(FIN, "GET", "/v1/ledger/journals", token=fin_tok)[0],
        "support_order": call(ADMIN, "GET", f"/v1/admin/orders/{oid}", token=support_tok)[0],
        "ops_dashboard": call(ADMIN, "GET", "/v1/admin/dashboard", token=ops_tok)[0],
        "admin_order": call(ADMIN, "GET", f"/v1/admin/orders/{oid}", token=adm_tok)[0],
        "supplier_list": call(SUP, "GET", "/v1/supplier/suppliers", token=sup_tok)[0],
        "super_platform": call(PLATFORM, "GET", "/v1/platform/admin/stats", token=super_tok)[0],
    }
    print("LEGIT", legit)

    print("== tenant")
    wrong_order = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=cust_tok, tenant=WRONG)[0]
    right_order = call(BFF, "GET", f"/v1/customer/orders/{oid}", token=cust_tok)[0]
    print("wrong_tenant_order", wrong_order, "right_order", right_order)

    print("== sse tickets (pre-event)")
    st_unauth, _ = curl_sse(RT + "/v1/realtime/sse?topic=" + urllib.parse.quote("order:" + oid))
    print("unauth_sse", st_unauth)
    st_ticket, ticket_body, _ = call(CUSTOMER, "POST", f"/v1/customer/orders/{oid}/realtime-ticket", {}, token=cust_tok)
    ticket = ticket_body.get("ticket") or ""
    print("own_ticket_http", st_ticket, "has_ticket", bool(ticket))
    st_ok, body_ok = 0, ""
    st_cross = 0
    st_wrong_tenant_sse = 0
    if ticket:
        st_ok, body_ok = curl_sse(
            RT + "/v1/realtime/sse?topic=" + urllib.parse.quote("order:" + oid) + "&ticket=" + urllib.parse.quote(ticket)
        )
        print("own_order_sse", st_ok, "body", body_ok.replace("\n", " ")[:80])
        st_cross, _ = curl_sse(
            RT + "/v1/realtime/sse?topic=" + urllib.parse.quote("order:00000000-0000-0000-0000-000000000000") + "&ticket=" + urllib.parse.quote(ticket)
        )
        print("cross_topic_sse", st_cross)
        st_wrong_tenant_sse, _ = curl_sse(
            RT + "/v1/realtime/sse?topic=" + urllib.parse.quote("order:" + oid) + "&ticket=" + urllib.parse.quote(ticket)
        )
        # ticket is tenant-bound; request with wrong tenant header is gateway-side. Mint ticket as wrong tenant:
    st_bad_tix, bad_tix, _ = call(CUSTOMER, "POST", f"/v1/customer/orders/{oid}/realtime-ticket", {}, token=cust_tok, tenant=WRONG)
    print("wrong_tenant_ticket", st_bad_tix)

    b_tok, _ = login_customer(PHONES["customer_b"])
    st_b_get = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=b_tok)[0]
    st_b_tix = call(CUSTOMER, "POST", f"/v1/customer/orders/{oid}/realtime-ticket", {}, token=b_tok)[0]
    print("customer_b_order", st_b_get, "customer_b_ticket", st_b_tix)

    open("/tmp/prompt80-tokens.json", "w").write(json.dumps({
        "oid": oid, "place_http": st, "order_get": st_get,
        "denials": denials, "legit": legit,
        "wrong_tenant_order": wrong_order, "right_order": right_order,
        "unauth_sse": st_unauth, "own_sse": st_ok, "cross_sse": st_cross,
        "wrong_tenant_ticket": st_bad_tix, "customer_b_order": st_b_get, "customer_b_ticket": st_b_tix,
        "tracking_direct": st_tr_direct, "bff_track": st_track,
        "wh": bool(wh_tok), "co": bool(co_tok), "co_principal": co_principal,
    }))
    print("EVIDENCE_ORDER", oid)
    bad = [k for k, v in denials.items() if v not in (401, 403)]
    if bad:
        print("RBAC_FAIL", bad)
        return 1
    if st_unauth not in (401, 403):
        print("SSE_UNAUTH_FAIL", st_unauth)
        return 1
    if st_ticket != 200 or not ticket:
        print("SSE_TICKET_FAIL")
        return 1
    if st_ok != 200:
        print("SSE_OWN_FAIL", st_ok)
        return 1
    if st_cross not in (401, 403):
        print("SSE_CROSS_FAIL", st_cross)
        return 1
    if st_b_get not in (401, 403, 404) or st_b_tix not in (401, 403, 404):
        print("SSE_OTHER_CUSTOMER_FAIL", st_b_get, st_b_tix)
        return 1
    if wrong_order not in (401, 403, 404):
        print("TENANT_FAIL", wrong_order)
        return 1
    if st_bad_tix not in (401, 403, 404):
        print("SSE_WRONG_TENANT_TICKET_FAIL", st_bad_tix)
        return 1
    print("PROMPT80_EVIDENCE_PRE_SSE_OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
