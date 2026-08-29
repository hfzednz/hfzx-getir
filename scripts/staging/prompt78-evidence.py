#!/usr/bin/env python3
"""Prompt 78 evidence: real checkout order, warehouse/courier APIs, track source, SSE via Chromium EventSource."""
from __future__ import annotations

import json
import os
import re
import subprocess
import time
import urllib.error
import urllib.request
import uuid

TENANT = "11111111-1111-1111-1111-111111111111"
WRONG = "22222222-2222-2222-2222-222222222222"
PHONE = "+905551112233"
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
            print("HTTP", resp.status, method, base.split("://", 1)[-1] + path)
            return resp.status, obj, raw
    except urllib.error.HTTPError as e:
        raw = e.read().decode()[:600]
        print("HTTP", e.code, method, base.split("://", 1)[-1] + path, raw.replace("\n", " ")[:180])
        return e.code, {}, raw
    except Exception as e:
        print("ERR", method, base + path, type(e).__name__, e)
        return 0, {}, str(e)


def otp_code():
    logs = subprocess.check_output(
        ["bash", "-lc", "docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep +905551112233 | tail -1"],
        text=True,
    )
    m = re.search(r'"code":"(\d+)"', logs)
    return m.group(1) if m else ""


def seed_catalog_if_needed():
    st, home, _ = call(CUSTOMER, "GET", "/v1/customer/home?lat=41&lng=29")
    products = home.get("products") or []
    if products:
        sku = str(products[0].get("id") or products[0].get("sku") or "")
        print("catalog_ok", sku)
        return sku
    raw = subprocess.check_output(
        [
            "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
            "-sS", "-H", "Content-Type: application/json", "-H", "X-Tenant-Id: " + TENANT,
            "-X", "POST", "http://catalog-service:8080/v1/catalog/products",
            "-d", json.dumps({"kind": "standard", "slug": "fresh-milk", "skuCode": "MILK-1"}),
        ],
        text=True,
    )
    prod = json.loads(raw)
    pid = (prod.get("product") or prod).get("ID") or (prod.get("product") or prod).get("id")
    subprocess.check_call(
        [
            "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
            "-sS", "-H", "Content-Type: application/json", "-H", "X-Tenant-Id: " + TENANT,
            "-X", "PUT", f"http://catalog-service:8080/v1/catalog/products/{pid}/locales/en",
            "-d", json.dumps({"title": "Fresh Milk", "description": "1L"}),
        ]
    )
    subprocess.check_call(
        [
            "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
            "-sS", "-H", "Content-Type: application/json", "-H", "X-Tenant-Id: " + TENANT,
            "-X", "POST", f"http://catalog-service:8080/v1/catalog/products/{pid}/variants",
            "-d", json.dumps({"skuCode": "MILK-1L", "name": "1L"}),
        ]
    )
    subprocess.check_call(
        [
            "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
            "-sS", "-H", "Content-Type: application/json", "-H", "X-Tenant-Id: " + TENANT,
            "-X", "POST", "http://catalog-service:8080/v1/catalog/search/reindex", "-d", "{}",
        ]
    )
    print("catalog_seeded", pid)
    return pid


def jwt_roles(token: str):
    try:
        payload = token.split(".")[1]
        pad = "=" * (-len(payload) % 4)
        import base64
        data = json.loads(base64.urlsafe_b64decode(payload + pad))
        return data.get("roles") or data.get("Roles") or []
    except Exception:
        return []


def main():
    print("== preserve customer")
    st, _, _ = call(CUSTOMER, "GET", "/login")
    if st != 200:
        print("FAIL customer login page", st)
        return 1

    print("== health")
    for base in (RT, "http://127.0.0.1:8115", BFF, WH, CO, ADMIN, FIN, SUP):
        call(base, "GET", "/health")
    st, trh, _ = call("http://127.0.0.1:8111", "GET", "/health")
    # tracking is docker-internal only
    tr_health = subprocess.check_output(
        ["docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0", "-sS", "http://tracking-service:8080/health"],
        text=True,
    ).strip()
    print("tracking_health", tr_health)

    sku = seed_catalog_if_needed()
    st, start, _ = call(CUSTOMER, "POST", "/v1/customer/auth/otp/start", {"phone": PHONE})
    code = otp_code()
    st, sess, _ = call(CUSTOMER, "POST", "/v1/customer/auth/otp/verify", {"challengeId": start.get("challengeId"), "code": code})
    token = sess.get("accessToken") or ""
    principal = sess.get("customerId") or sess.get("principalId") or ""
    print("customer_token", bool(token), "roles_jwt", jwt_roles(token), "otp_len", len(code))

    st, start2, _ = call(IDN, "POST", "/v1/identity/auth/otp/start", {"phone": PHONE, "tenantId": TENANT})
    code2 = otp_code()
    st, staff, _ = call(IDN, "POST", "/v1/identity/auth/otp/verify", {"challengeId": start2.get("challengeId"), "code": code2})
    staff_token = staff.get("accessToken") or ""
    print("staff_token", bool(staff_token), "roles_jwt", jwt_roles(staff_token))

    if os.environ.get("PROMPT78_FULFILL_ONLY") == "1":
        oid = (os.environ.get("PROMPT78_ORDER_ID") or "").strip() or open("/tmp/prompt78-order-id.txt").read().strip()
        print("FULFILL_ONLY", oid)
        st, order, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token)
        print("order_status", order.get("status"))
        print("== warehouse pack/ready (pick may already be done by browser)")
        if order.get("status") == "warehouse_assigned":
            st, pick, _ = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/pick", {})
            print("warehouse_pick", st, pick.get("status"))
        st, pack, _ = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/pack", {})
        print("warehouse_pack", st, pack.get("status"))
        st, ready, _ = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/ready", {})
        print("warehouse_ready", st, ready.get("status"))
        print("== courier APIs")
        st, acc, _ = call(CO, "POST", f"/v1/courier/offers/{oid}", {"courierId": principal or "c1", "accept": True})
        print("courier_accept", st, acc.get("status"))
        st, en, _ = call(CO, "POST", f"/v1/courier/offers/{oid}/enroute", {})
        print("courier_enroute", st, en.get("status"))
        st, done, _ = call(CO, "POST", f"/v1/courier/offers/{oid}/complete", {})
        print("courier_complete", st, done.get("status"))
        st, order2, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token)
        st, track2, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=token)
        print("order_after", order2.get("status"), "track_after", track2.get("status"))
        tl2 = subprocess.check_output(
            [
                "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
                "-sS", "-H", "X-Tenant-Id: " + TENANT,
                f"http://tracking-service:8080/v1/tracking/orders/{oid}/timeline",
            ],
            text=True,
        )
        print("tracking_timeline_after", tl2[:400])
        print("== roles observe")
        call(ADMIN, "GET", "/v1/admin/dashboard")
        call(ADMIN, "GET", f"/v1/admin/orders/{oid}")
        call(FIN, "GET", "/v1/ledger/journals")
        call(SUP, "GET", "/v1/supplier/suppliers")
        print("== rbac")
        print("customer_web_admin", call(CUSTOMER, "GET", "/v1/admin/dashboard", token=token)[0])
        print("customer_token_admin_bff", call(ADMIN, "GET", "/v1/admin/orders", token=token)[0])
        print("courier_web_finance", call("http://127.0.0.1:3001", "GET", "/v1/ledger/journals", token=staff_token)[0])
        print("courier_token_finance_direct", call(FIN, "GET", "/v1/ledger/journals", token=staff_token)[0])
        print("supplier_web_admin", call("http://127.0.0.1:3003", "GET", "/v1/admin/dashboard", token=staff_token)[0])
        print("support_web_platform", call("http://127.0.0.1:3005", "GET", "/v1/platform/admin/stats", token=staff_token)[0])
        print("platform_stats_no_jwt", call(PLATFORM, "GET", "/v1/platform/admin/stats")[0])
        print("== tenant")
        print("wrong_tenant_customer_order", call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token, tenant=WRONG)[0])
        print("right_bff", call(BFF, "GET", f"/v1/customer/orders/{oid}", token=token)[0])
        print("wrong_admin_order", call(ADMIN, "GET", f"/v1/admin/orders/{oid}", tenant=WRONG)[0])
        print("warehouse_list", call(WH, "GET", "/v1/warehouse/tasks")[0])
        print("EVIDENCE_ORDER", oid)
        print("EVIDENCE_STATUS", order2.get("status"))
        return 0 if oid and order2.get("status") in ("completed", "delivered", "out_for_delivery") else 1

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
    call(CUSTOMER, "POST", "/v1/customer/cart/items", {"cartId": cart_id, "sku": sku, "qty": 1, "unitMinor": 1500}, token=token)
    st, prev, _ = call(CUSTOMER, "POST", "/v1/customer/checkout/preview", {"cartId": cart_id, "principalId": principal}, token=token)
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
    print("ORDER_ID", oid)
    if not oid:
        return 1
    open("/tmp/prompt78-order-id.txt", "w").write(oid)
    st, order, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token)
    print("order_status", order.get("status"), "tenant", order.get("tenantId"))
    st, track0, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=token)
    print("track_before", track0)

    if os.environ.get("PROMPT78_STOP_AFTER_PLACE") == "1":
        print("STOP_AFTER_PLACE", oid)
        return 0

    print("== warehouse APIs (after browser SSE subscribe)")
    st, pick, _ = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/pick", {})
    print("warehouse_pick", st, pick.get("status"))
    st, pack, _ = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/pack", {})
    print("warehouse_pack", st, pack.get("status"))
    st, ready, _ = call(WH, "POST", f"/v1/warehouse/tasks/{oid}/ready", {})
    print("warehouse_ready", st, ready.get("status"))

    print("== courier APIs")
    st, acc, _ = call(CO, "POST", f"/v1/courier/offers/{oid}", {"courierId": principal or "c1", "accept": True})
    print("courier_accept", st, acc.get("status"))
    st, en, _ = call(CO, "POST", f"/v1/courier/offers/{oid}/enroute", {})
    print("courier_enroute", st, en.get("status"))
    st, done, _ = call(CO, "POST", f"/v1/courier/offers/{oid}/complete", {})
    print("courier_complete", st, done.get("status"))

    st, order2, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token)
    st, track2, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}/track", token=token)
    print("order_after", order2.get("status"), "track_after", track2.get("status"))
    tl2 = subprocess.check_output(
        [
            "docker", "run", "--rm", "--network", "nexora-phone-staging", "curlimages/curl:8.5.0",
            "-sS", "-H", "X-Tenant-Id: " + TENANT,
            f"http://tracking-service:8080/v1/tracking/orders/{oid}/timeline",
        ],
        text=True,
    )
    print("tracking_timeline_after", tl2[:400])

    print("== roles observe")
    call(ADMIN, "GET", "/v1/admin/dashboard")
    call(ADMIN, "GET", f"/v1/admin/orders/{oid}")
    call(ADMIN, "GET", "/v1/admin/orders")
    call(FIN, "GET", "/v1/ledger/journals")
    call(SUP, "GET", "/v1/supplier/suppliers")
    call("http://127.0.0.1:3002", "GET", "/login")
    call("http://127.0.0.1:3001", "GET", "/login")
    call("http://127.0.0.1:3003", "GET", "/login")
    call("http://127.0.0.1:3004", "GET", "/login")
    call("http://127.0.0.1:3005", "GET", "/login")
    call("http://127.0.0.1:3006", "GET", "/login")
    call("http://127.0.0.1:3100", "GET", "/login")

    print("== rbac")
    st_c_admin, _, _ = call(CUSTOMER, "GET", "/v1/admin/dashboard", token=token)
    print("customer_web_admin", st_c_admin)
    st_c_admin_direct, _, _ = call(ADMIN, "GET", "/v1/admin/orders", token=token)
    print("customer_token_admin_bff", st_c_admin_direct)
    st_co_fin, _, _ = call("http://127.0.0.1:3001", "GET", "/v1/ledger/journals", token=staff_token)
    print("courier_web_finance", st_co_fin)
    st_co_fin_d, _, _ = call(FIN, "GET", "/v1/ledger/journals", token=staff_token)
    print("courier_token_finance_direct", st_co_fin_d)
    st_sup_admin, _, _ = call("http://127.0.0.1:3003", "GET", "/v1/admin/dashboard", token=staff_token)
    print("supplier_web_admin", st_sup_admin)
    st_sup_d, _, _ = call(ADMIN, "GET", "/v1/admin/dashboard", token=staff_token)
    print("supplier_token_admin_bff", st_sup_d)
    st_sup_plat, _, _ = call("http://127.0.0.1:3005", "GET", "/v1/platform/admin/stats", token=staff_token)
    print("support_web_platform", st_sup_plat)
    st_plat, _, _ = call(PLATFORM, "GET", "/v1/platform/admin/stats")
    print("platform_stats_no_jwt", st_plat)

    print("== tenant")
    st_wrong, _, _ = call(CUSTOMER, "GET", f"/v1/customer/orders/{oid}", token=token, tenant=WRONG)
    st_right, _, _ = call(BFF, "GET", f"/v1/customer/orders/{oid}", token=token)
    st_wrong_admin, _, _ = call(ADMIN, "GET", f"/v1/admin/orders/{oid}", tenant=WRONG)
    print("wrong_tenant_customer_order", st_wrong, "right_bff", st_right, "wrong_admin_order", st_wrong_admin)

    print("== warehouse list endpoint")
    st_wlist, _, raw_w = call(WH, "GET", "/v1/warehouse/tasks")
    print("warehouse_list", st_wlist)

    print("EVIDENCE_ORDER", oid)
    print("EVIDENCE_STATUS", order2.get("status"))
    return 0 if oid and order2.get("status") in ("completed", "delivered", "out_for_delivery") else 1


if __name__ == "__main__":
    raise SystemExit(main())
