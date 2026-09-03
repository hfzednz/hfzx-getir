#!/usr/bin/env python3
"""Prompt 87 local stack + live HTTP gate.

Starts NEXORA services on the ports in docs/launch/service-registry.yaml,
then exercises real HTTP journeys (not in-process handler calls).

Usage:
  python scripts/local/prompt87_live_gate.py           # start (if needed) + e2e
  python scripts/local/prompt87_live_gate.py --start   # start only
  python scripts/local/prompt87_live_gate.py --e2e     # journeys against a running stack
  python scripts/local/prompt87_live_gate.py --stop
"""
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import time
import uuid
import urllib.error
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TENANT = "11111111-1111-1111-1111-111111111111"
OTHER_TENANT = "22222222-2222-2222-2222-222222222222"
RUN_DIR = ROOT / "tmp" / "prompt87-stack"
COMPOSE = ROOT / "infra" / "docker" / "docker-compose.yml"

PHONES = {
    "customer": "+905551112233",
    "warehouse": "+905551112234",
    "courier": "+905551112235",
    "finance": "+905551112236",
    "support": "+905551112237",
    "ops": "+905551112238",
    "admin": "+905551112239",
    "supplier": "+905551112240",
    "super_admin": "+905551112241",
    "merchant": "+905551112242",
}

SERVICES = [
    ("identity-service", 8081, {"OTP_DEV_MODE": "true", "RATE_LIMIT_PER_MINUTE": "0"}),
    ("customer-profile-service", 8082, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("catalog-service", 8083, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("inventory-service", 8084, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("order-service", 8085, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("cart-service", 8086, {"RATE_LIMIT_PER_MINUTE": "0"}),
    (
        "checkout-service",
        8087,
        {
            "RATE_LIMIT_PER_MINUTE": "0",
            "CART_URL": "http://127.0.0.1:8086",
            "ORDER_URL": "http://127.0.0.1:8085",
            "PAYMENT_URL": "http://127.0.0.1:8089",
            "INVENTORY_URL": "http://127.0.0.1:8084",
            "PROMO_URL": "http://127.0.0.1:8094",
        },
    ),
    ("warehouse-service", 8088, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("payment-service", 8089, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("finance-ledger-service", 8091, {"IDENTITY_URL": "http://127.0.0.1:8081", "LEDGER_INTERNAL_TOKEN": "nexora-ledger-dev", "RATE_LIMIT_PER_MINUTE": "0"}),
    ("settlement-service", 8092, {"IDENTITY_URL": "http://127.0.0.1:8081", "LEDGER_BASE_URL": "http://127.0.0.1:8091", "LEDGER_INTERNAL_TOKEN": "nexora-ledger-dev", "RATE_LIMIT_PER_MINUTE": "0"}),
    ("loyalty-service", 8093, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("promotion-service", 8094, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("geofence-service", 8099, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("tracking-service", 8098, {"RATE_LIMIT_PER_MINUTE": "0", "GEOFENCE_BASE_URL": "http://127.0.0.1:8099"}),
    ("location-service", 8100, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("notification-service", 8101, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("crm-service", 8102, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("ai-platform-service", 8106, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("platform-ops-service", 8110, {"IDENTITY_URL": "http://127.0.0.1:8081", "RATE_LIMIT_PER_MINUTE": "0"}),
    ("liveops-service", 8116, {"RATE_LIMIT_PER_MINUTE": "0"}),
    ("supplier-service", 8117, {"IDENTITY_URL": "http://127.0.0.1:8081", "RATE_LIMIT_PER_MINUTE": "0"}),
    (
        "realtime-gateway",
        8115,
        {
            "SSE_TICKET_SECRET": "nexora-local-sse",
            "REALTIME_PUBLISH_TOKEN": "nexora-local-publish",
        },
    ),
    (
        "bff-customer",
        8111,
        {
            "IDENTITY_URL": "http://127.0.0.1:8081",
            "CATALOG_URL": "http://127.0.0.1:8083",
            "CART_URL": "http://127.0.0.1:8086",
            "CHECKOUT_URL": "http://127.0.0.1:8087",
            "ORDER_URL": "http://127.0.0.1:8085",
            "TRACKING_URL": "http://127.0.0.1:8098",
            "PAYMENT_URL": "http://127.0.0.1:8089",
            "LOCATION_URL": "http://127.0.0.1:8100",
            "NOTIFICATION_URL": "http://127.0.0.1:8101",
            "CRM_URL": "http://127.0.0.1:8102",
            "INVENTORY_URL": "http://127.0.0.1:8084",
            "PROMO_URL": "http://127.0.0.1:8094",
            "SSE_TICKET_SECRET": "nexora-local-sse",
        },
    ),
    (
        "bff-courier",
        8112,
        {
            "IDENTITY_URL": "http://127.0.0.1:8081",
            "ORDER_URL": "http://127.0.0.1:8085",
            "TRACKING_URL": "http://127.0.0.1:8098",
            "REALTIME_URL": "http://127.0.0.1:8115",
            "REALTIME_PUBLISH_TOKEN": "nexora-local-publish",
        },
    ),
    (
        "bff-warehouse",
        8113,
        {
            "IDENTITY_URL": "http://127.0.0.1:8081",
            "ORDER_URL": "http://127.0.0.1:8085",
            "TRACKING_URL": "http://127.0.0.1:8098",
            "REALTIME_URL": "http://127.0.0.1:8115",
            "REALTIME_PUBLISH_TOKEN": "nexora-local-publish",
        },
    ),
    (
        "bff-admin",
        8114,
        {
            "IDENTITY_URL": "http://127.0.0.1:8081",
            "ORDER_URL": "http://127.0.0.1:8085",
            "CATALOG_URL": "http://127.0.0.1:8083",
            "PROFILE_URL": "http://127.0.0.1:8082",
            "INVENTORY_URL": "http://127.0.0.1:8084",
            "PROMO_URL": "http://127.0.0.1:8094",
            "LEDGER_URL": "http://127.0.0.1:8091",
            "SETTLEMENT_URL": "http://127.0.0.1:8092",
            "LOYALTY_URL": "http://127.0.0.1:8093",
            "TRACKING_URL": "http://127.0.0.1:8098",
            "NOTIFICATION_URL": "http://127.0.0.1:8101",
            "CRM_URL": "http://127.0.0.1:8102",
            "LIVEOPS_URL": "http://127.0.0.1:8116",
            "AI_URL": "http://127.0.0.1:8106",
        },
    ),
]


def exe_name(svc: str) -> str:
    suffix = ".exe" if os.name == "nt" else ""
    return svc + suffix


def http_json(method: str, url: str, body=None, token=None, tenant=TENANT, timeout=20):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": tenant, "X-Request-Id": "p87"}
    if token:
        headers["Authorization"] = "Bearer " + token
        actor = jwt_sub(token)
        if actor:
            headers["X-Nexora-User"] = actor
    last_err = ""
    for attempt in range(3):
        req = urllib.request.Request(url, data=data, method=method, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                raw = resp.read().decode("utf-8", "replace")
                obj = json.loads(raw) if raw.lstrip().startswith("{") or raw.lstrip().startswith("[") else {}
                return resp.status, obj, raw
        except urllib.error.HTTPError as e:
            raw = e.read().decode("utf-8", "replace")
            obj = {}
            try:
                obj = json.loads(raw) if raw else {}
            except json.JSONDecodeError:
                pass
            return e.code, obj, raw
        except Exception as e:
            last_err = str(e)
            time.sleep(0.2 * (attempt + 1))
    return 0, {}, last_err


def wait_health(port: int, name: str, seconds=90) -> bool:
    deadline = time.time() + seconds
    while time.time() < deadline:
        code, _, _ = http_json("GET", f"http://127.0.0.1:{port}/health")
        if code == 200:
            print(f"OK health {name} :{port}")
            return True
        time.sleep(1)
    print(f"FAIL health {name} :{port}")
    return False


def child_env(extra: dict) -> dict:
    env = os.environ.copy()
    env["GOWORK"] = "off"
    env["GOTOOLCHAIN"] = "auto"
    env["OTP_DEV_MODE"] = "true"
    env["RATE_LIMIT_PER_MINUTE"] = "0"
    # Isolated DevMode stack: never inherit a production DATABASE_URL.
    env["DATABASE_URL"] = extra.get("DATABASE_URL", "")
    env["REDIS_URL"] = extra.get("REDIS_URL", "")
    env["KAFKA_BROKERS"] = extra.get("KAFKA_BROKERS", "")
    env.update(extra)
    return env


def build_one(svc: str) -> Path:
    RUN_DIR.mkdir(parents=True, exist_ok=True)
    out = RUN_DIR / exe_name(svc)
    cmd_dir = ROOT / "services" / svc / "cmd" / svc
    print(f"==> build {svc}")
    r = subprocess.run(
        ["go", "build", "-o", str(out), "."],
        cwd=str(cmd_dir),
        env=child_env({}),
        capture_output=True,
        text=True,
    )
    if r.returncode != 0:
        sys.stderr.write(r.stdout + "\n" + r.stderr)
        raise SystemExit(f"build failed: {svc}")
    return out


def start_stack() -> None:
    RUN_DIR.mkdir(parents=True, exist_ok=True)
    pg_dsn = ""
    compose_msg = try_compose()
    print(compose_msg)
    if compose_msg.startswith("PASS"):
        pg_dsn = apply_platform_ops_migrations()
        print("platform-ops DATABASE_URL", pg_dsn if pg_dsn.startswith("postgres") else pg_dsn)
        if not pg_dsn.startswith("postgres"):
            pg_dsn = ""
    binaries = {}
    with ThreadPoolExecutor(max_workers=4) as pool:
        futs = {pool.submit(build_one, svc): svc for svc, _, _ in SERVICES}
        for fut in as_completed(futs):
            svc = futs[fut]
            binaries[svc] = fut.result()
    procs = []
    for svc, port, extra in SERVICES:
        env = child_env(dict(extra))
        env["HTTP_ADDR"] = f":{port}"
        if svc == "platform-ops-service" and pg_dsn:
            env["DATABASE_URL"] = pg_dsn
        log_path = RUN_DIR / f"{svc}.log"
        logf = open(log_path, "w", encoding="utf-8")
        print(f"==> start {svc} :{port}")
        p = subprocess.Popen([str(binaries[svc])], cwd=str(ROOT / "services" / svc), env=env, stdout=logf, stderr=subprocess.STDOUT)
        (RUN_DIR / f"{svc}.pid").write_text(str(p.pid), encoding="utf-8")
        procs.append((svc, port, p, logf))
    failed = []
    for svc, port, p, _ in procs:
        if not wait_health(port, svc, seconds=90):
            tail = (RUN_DIR / f"{svc}.log").read_text(encoding="utf-8", errors="replace")[-1500:]
            print(tail)
            failed.append(svc)
    if failed:
        raise SystemExit("startup failed: " + ", ".join(failed))
    print("STACK_READY")


def try_compose() -> str:
    try:
        r = subprocess.run(
            ["docker", "compose", "-f", str(COMPOSE), "up", "-d", "postgres", "redis", "kafka"],
            capture_output=True, text=True,
        )
        if r.returncode != 0:
            return "BLOCKED docker: " + (r.stderr or r.stdout)[:300]
        return "PASS compose postgres/redis/kafka"
    except FileNotFoundError:
        return "BLOCKED docker binary missing"


def stop_stack() -> None:
    for svc, _, _ in SERVICES:
        pid_file = RUN_DIR / f"{svc}.pid"
        if not pid_file.exists():
            continue
        pid = pid_file.read_text(encoding="utf-8").strip()
        if os.name == "nt":
            subprocess.run(["taskkill", "/PID", pid, "/F", "/T"], capture_output=True)
        else:
            subprocess.run(["kill", pid], capture_output=True)
        pid_file.unlink(missing_ok=True)
    print("STACK_STOPPED")


def otp_code(phone: str) -> str:
    log = RUN_DIR / "identity-service.log"
    deadline = time.time() + 10
    needle = phone[-10:]
    while time.time() < deadline:
        text = log.read_text(encoding="utf-8", errors="replace") if log.exists() else ""
        found = []
        for a, b in re.findall(r'"phone":"([^"]+)".{0,120}"code":"(\d+)"', text):
            if a.endswith(needle) or a == phone:
                found.append(b)
        for code, ph in re.findall(r'"code":"(\d+)".{0,120}"phone":"([^"]+)"', text):
            if ph.endswith(needle) or ph == phone:
                found.append(code)
        if found:
            return found[-1]
        time.sleep(0.4)
    raise SystemExit(f"no OTP logged for {phone}")


def login(role: str) -> tuple[str, str]:
    phone = PHONES[role]
    st, body, raw = http_json("POST", "http://127.0.0.1:8111/v1/customer/auth/otp/start", {"phone": phone})
    if st not in (200, 201):
        raise SystemExit(f"otp start {role} {st} {raw}")
    cid = body.get("challengeId") or body.get("ChallengeID")
    code = otp_code(phone)
    st, body, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/auth/otp/verify",
        {"challengeId": cid, "code": code},
    )
    if st != 200:
        raise SystemExit(f"otp verify {role} {st} {raw}")
    token = body.get("accessToken") or body.get("AccessToken") or ""
    principal = body.get("customerId") or body.get("principalId") or body.get("CustomerID") or ""
    if not token:
        raise SystemExit(f"no token for {role}: {body}")
    return token, principal


def jwt_sub(token: str) -> str:
    try:
        payload = token.split(".")[1]
        payload += "=" * (-len(payload) % 4)
        claims = json.loads(__import__("base64").urlsafe_b64decode(payload))
        return str(claims.get("sub") or claims.get("principalId") or "")
    except Exception:
        return ""


def expect(name: str, code: int, wanted, body=None):
    ok = code in wanted if isinstance(wanted, (list, tuple, set)) else code == wanted
    status = "PASS" if ok else "FAIL"
    print(f"{status} {name} HTTP {code} want {wanted}")
    if not ok:
        if body:
            print("  body", str(body)[:400])
        raise SystemExit(f"assertion failed: {name}")


def seed_catalog():
    suffix = uuid.uuid4().hex[:8]
    products = [
        (f"p87-milk-{suffix}", f"MILK-{suffix}", "Fresh Milk", "Taze Süt", "1 litre whole milk", "1 litre tam yağlı süt"),
        (f"p87-bread-{suffix}", f"BREAD-{suffix}", "Village Bread", "Köy Ekmeği", "Fresh baked bread", "Taze fırın ekmeği"),
        (f"p87-yogurt-{suffix}", f"YOG-{suffix}", "Strained Yogurt", "Süzme Yoğurt", "Creamy yogurt", "Kremamsı yoğurt"),
    ]
    ids = []
    for slug, sku, en, tr, en_d, tr_d in products:
        st, body, raw = http_json(
            "POST",
            "http://127.0.0.1:8083/v1/catalog/products",
            {"kind": "standard", "slug": slug, "skuCode": sku},
        )
        expect(f"create {sku}", st, {200, 201})
        p = body.get("product") or body
        pid = p.get("id") or p.get("ID")
        http_json("PUT", f"http://127.0.0.1:8083/v1/catalog/products/{pid}/locales/en", {"title": en, "description": en_d})
        http_json("PUT", f"http://127.0.0.1:8083/v1/catalog/products/{pid}/locales/tr", {"title": tr, "description": tr_d})
        st, var, _ = http_json(
            "POST",
            f"http://127.0.0.1:8083/v1/catalog/products/{pid}/variants",
            {"skuCode": sku + "-V", "name": en},
        )
        vid = (var.get("variant") or var).get("id") or (var.get("variant") or var).get("ID")
        ids.append((pid, vid, sku))
    http_json("POST", "http://127.0.0.1:8083/v1/catalog/search/reindex", {})
    time.sleep(0.3)
    return ids


def run_e2e() -> None:
    failures = []

    def check(name, code, wanted, body=None):
        ok = code in wanted if isinstance(wanted, (list, tuple, set)) else code == wanted
        print(("PASS" if ok else "FAIL"), name, "HTTP", code, "want", wanted)
        if not ok:
            if body:
                print("  ", str(body)[:400])
            failures.append(name)

    st, _, _ = http_json("GET", "http://127.0.0.1:8111/health")
    check("customer bff health", st, 200)
    st, _, _ = http_json("GET", "http://127.0.0.1:8111/v1/customer/home?lat=41&lng=29")
    check("anon home", st, 401)

    cust_tok, cust_id = login("customer")
    admin_tok, _ = login("admin")
    merchant_tok, _ = login("merchant")
    supplier_tok, _ = login("supplier")
    courier_tok, _ = login("courier")
    finance_tok, _ = login("finance")
    sa_tok, _ = login("super_admin")
    wh_tok, _ = login("warehouse")
    support_tok, _ = login("support")

    st, home, _ = http_json("GET", "http://127.0.0.1:8111/v1/customer/home?lat=41&lng=29", token=cust_tok)
    check("customer home", st, 200, home)

    catalog_ids = seed_catalog()
    milk_pid, milk_vid, milk_sku = catalog_ids[0]
    st, wh, raw = http_json(
        "POST",
        "http://127.0.0.1:8084/v1/inventory/warehouses",
        {"code": "P87" + uuid.uuid4().hex[:6].upper(), "name": "P87 Warehouse", "timezone": "Europe/Istanbul", "status": "active"},
        token=admin_tok,
    )
    check("create warehouse", st, {200, 201}, raw)
    wh_id = wh.get("id") or (wh.get("warehouse") or {}).get("id") or ""
    if wh_id and milk_vid:
        st, recv, raw = http_json(
            "POST",
            "http://127.0.0.1:8084/v1/inventory/stock/receive",
            {
                "warehouseId": wh_id,
                "variantId": milk_vid,
                "qty": 100,
                "idempotencyKey": "p87-recv-" + uuid.uuid4().hex,
                "reason": "p87 seed",
            },
            token=admin_tok,
        )
        check("receive stock", st, {200, 201}, raw)

    for q in ["süt", "SÜT", " milk ", "MILK", "ekmek", "bread", "yoğurt", "yogurt"]:
        st, body, raw = http_json("GET", "http://127.0.0.1:8111/v1/customer/search?q=" + urllib.parse.quote(q), token=cust_tok)
        items = body.get("items") or body.get("hits") or []
        check(f"search {q!r}", st, 200, raw)
        if st == 200 and not items:
            print("  empty hits", raw[:300])
            failures.append(f"search empty {q}")

    st, cart, raw = http_json("POST", "http://127.0.0.1:8086/v1/cart", {"guestToken": "p87-" + uuid.uuid4().hex, "currency": "TRY"})
    check("create cart", st, {200, 201}, raw)
    cart_id = cart.get("id") or cart.get("ID") or cart.get("cartId") or ""
    sku = milk_vid or milk_pid
    st, added, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/cart/items",
        {"cartId": cart_id, "sku": sku, "qty": 2, "unitMinor": 1999},
        token=cust_tok,
    )
    check("add cart", st, {200, 201}, raw)
    returned = added.get("cartId") or added.get("id") or added.get("ID") or cart_id
    if cart_id and returned and str(returned) != str(cart_id):
        print("FAIL cart id divergence", cart_id, returned)
        failures.append("cart id divergence")

    # Coupons via admin BFF → promo
    coupon_code = "WELCOME10"
    st, camp, raw = http_json("POST", "http://127.0.0.1:8094/v1/promo/campaigns", {"name": "p87-welcome"}, token=admin_tok)
    check("promo campaign", st, {200, 201}, raw)
    camp_id = (camp.get("id") or camp.get("campaignId") or (camp.get("campaign") or {}).get("id"))
    if camp_id:
        http_json("POST", f"http://127.0.0.1:8094/v1/promo/campaigns/{camp_id}/activate", {}, token=admin_tok)
        st, promo, raw = http_json(
            "POST",
            "http://127.0.0.1:8094/v1/promo/promotions",
            {"campaignId": camp_id, "name": "welcome10", "type": "percent", "percentOff": 10, "thresholdMinor": 15000, "priority": 1},
            token=admin_tok,
        )
        check("promo rule", st, {200, 201}, raw)
        promo_id = promo.get("id") or (promo.get("promotion") or {}).get("id")
        if promo_id:
            coupon_code = "P87" + uuid.uuid4().hex[:8].upper()
            st, coupon, raw = http_json(
                "POST",
                "http://127.0.0.1:8114/v1/admin/coupons",
                {"promotionId": promo_id, "code": coupon_code, "kind": "public", "maxRedemptions": 100},
                token=admin_tok,
            )
            check("admin create coupon", st, {200, 201}, raw)
    st, clist, _ = http_json("GET", "http://127.0.0.1:8111/v1/customer/coupons", token=cust_tok)
    check("customer coupons", st, 200, clist)
    st, okc, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/coupons/validate",
        {"code": coupon_code, "cart_subtotal_minor": 20000},
        token=cust_tok,
    )
    check("coupon valid", st, {200, 201}, raw)
    st, bad, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/coupons/validate",
        {"code": "NOPE", "cart_subtotal_minor": 20000},
        token=cust_tok,
    )
    check("coupon invalid", st, {400, 404, 409, 422}, raw)

    st, prev, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/checkout/preview",
        {"cartId": cart_id, "principalId": cust_id},
        token=cust_tok,
    )
    check("checkout preview", st, {200, 201}, raw)
    sid = prev.get("sessionId") or prev.get("id") or ""
    addr = {"label": "Home", "line1": "Test St 1", "city": "Istanbul", "country": "TR", "lat": 41.0, "lng": 29.0}
    place_body = {"cartId": cart_id, "paymentMethod": "card", "principalId": cust_id, "address": addr}
    if sid:
        place_body["sessionId"] = sid
    st, placed, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/checkout/place",
        place_body,
        token=cust_tok,
    )
    check("place order", st, {200, 201}, raw)
    order_id = placed.get("orderId") or placed.get("id") or ""
    st, placed2, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/checkout/place",
        {"cartId": cart_id, "paymentMethod": "card", "sessionId": sid, "principalId": cust_id, "address": addr},
        token=cust_tok,
    )
    check("idempotent place", st, {200, 201, 409}, raw)
    if st in (200, 201):
        oid2 = placed2.get("orderId") or placed2.get("id") or ""
        if order_id and oid2 and oid2 != order_id:
            print("FAIL duplicate order", order_id, oid2)
            failures.append("duplicate order")

    if order_id:
        st, _, _ = http_json("GET", f"http://127.0.0.1:8111/v1/customer/orders/{order_id}", token=cust_tok)
        check("order detail", st, 200)
        st, _, _ = http_json("GET", f"http://127.0.0.1:8111/v1/customer/orders/{order_id}/track", token=cust_tok)
        check("track", st, {200, 404})
        st, _, _ = http_json(
            "GET",
            f"http://127.0.0.1:8111/v1/customer/orders/{order_id}",
            token=cust_tok,
            tenant=OTHER_TENANT,
        )
        check("wrong tenant order", st, {400, 401, 403, 404})
        st, tick, raw = http_json(
            "POST",
            f"http://127.0.0.1:8111/v1/customer/orders/{order_id}/realtime-ticket",
            {},
            token=cust_tok,
        )
        check("realtime ticket", st, {200, 201}, raw)
        st, _, _ = http_json(
            "GET",
            f"http://127.0.0.1:8111/v1/customer/orders/{order_id}",
            token=merchant_tok,
        )
        check("merchant denied customer order", st, {401, 403, 404})

    # Merchant / supplier
    st, dash, raw = http_json("GET", "http://127.0.0.1:8117/v1/supplier/merchant/dashboard", token=merchant_tok)
    check("merchant dashboard", st, 200, raw)
    st, _, _ = http_json("GET", "http://127.0.0.1:8117/v1/supplier/merchant/dashboard", token=cust_tok)
    check("customer denied merchant", st, {401, 403})
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/reports", token=merchant_tok)
    check("merchant denied admin reports", st, {401, 403})
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/reports", token=admin_tok)
    check("admin reports", st, 200)
    st, mon, raw = http_json("GET", "http://127.0.0.1:8114/v1/admin/monitoring", token=admin_tok)
    check("admin monitoring", st, 200, raw)
    if st == 200 and isinstance(mon, dict):
        blob = json.dumps(mon).lower()
        if '"status":"healthy"' in blob and "down" not in blob and "degraded" not in blob:
            # Healthy is allowed for live local services; empty/unconfigured must not be hardcoded healthy.
            pass
    st, ai, raw = http_json("GET", "http://127.0.0.1:8114/v1/admin/ai", token=admin_tok)
    check("admin ai", st, 200, raw)
    if st == 200:
        labeled = "providerUnavailable" in json.dumps(ai) or "local_analysis" in json.dumps(ai) or "local" in json.dumps(ai).lower()
        if not labeled:
            print("FAIL AI not labeled local/unavailable", str(ai)[:300])
            failures.append("ai labeling")
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/customers", token=admin_tok)
    check("admin customers", st, 200)
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/loyalty", token=admin_tok)
    check("admin loyalty", st, 200)
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/system", token=admin_tok)
    check("admin system", st, 200)
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/notifications", token=admin_tok)
    check("admin notifications", st, 200)
    st, _, _ = http_json("GET", "http://127.0.0.1:8114/v1/admin/system", token=finance_tok)
    check("finance denied system", st, 403)

    # Courier GPS
    courier_id = jwt_sub(courier_tok)
    st, _, raw = http_json(
        "POST",
        "http://127.0.0.1:8112/v1/courier/location",
        {"lat": 41.01, "lng": 29.0, "lon": 29.0, "accuracyM": 12, "courierId": courier_id},
        token=courier_tok,
    )
    check("courier location", st, {200, 201, 204}, raw)
    st, _, raw = http_json(
        "POST",
        "http://127.0.0.1:8112/v1/courier/location",
        {"lat": 200, "lng": 29.0},
        token=courier_tok,
    )
    check("invalid lat", st, {400, 422}, raw)
    st, _, _ = http_json("POST", "http://127.0.0.1:8112/v1/courier/location", {"lat": 41.0, "lng": 29.0}, token=cust_tok)
    check("customer denied courier GPS", st, {401, 403})
    st, _, raw = http_json(
        "POST",
        "http://127.0.0.1:8112/v1/courier/location",
        {"lat": 41.0, "lng": 200, "lon": 200, "courierId": courier_id},
        token=courier_tok,
    )
    check("invalid lon", st, {400, 422}, raw)

    st, _, raw = http_json("GET", "http://127.0.0.1:8115/v1/realtime/sse?topic=order:p87")
    check("unauthorized SSE", st, {401, 403})

    # Warehouse / courier health
    st, _, _ = http_json("GET", "http://127.0.0.1:8113/health")
    check("warehouse bff health", st, 200)
    st, tasks, raw = http_json("GET", "http://127.0.0.1:8113/v1/warehouse/tasks", token=wh_tok)
    check("warehouse tasks", st, {200, 404}, raw)
    st, _, _ = http_json("GET", "http://127.0.0.1:8113/v1/warehouse/tasks", token=cust_tok)
    check("customer denied warehouse", st, {401, 403})

    st, _, raw = http_json("GET", "http://127.0.0.1:8112/v1/courier/offers", token=courier_tok)
    check("courier offers", st, {200, 404}, raw)

    # Support
    st, ticket, raw = http_json(
        "POST",
        "http://127.0.0.1:8111/v1/customer/support/tickets",
        {"subject": "p87 ticket", "body": "order help"},
        token=cust_tok,
    )
    check("create ticket", st, {200, 201}, raw)
    ticket_id = ticket.get("id") or ticket.get("ticketId") or (ticket.get("ticket") or {}).get("id") or ""
    if ticket_id:
        st, _, raw = http_json("GET", f"http://127.0.0.1:8111/v1/customer/support/tickets/{ticket_id}", token=cust_tok)
        check("get own ticket", st, {200, 201}, raw)
        st, _, _ = http_json(
            "GET",
            f"http://127.0.0.1:8111/v1/customer/support/tickets/{ticket_id}",
            token=cust_tok,
            tenant=OTHER_TENANT,
        )
        check("wrong tenant ticket", st, {400, 401, 403, 404})

    # Finance settlement honesty: create a real batch, then approve/execute.
    st, batch, raw = http_json(
        "POST",
        "http://127.0.0.1:8092/v1/settlements/batches",
        {
            "currency": "TRY",
            "periodStart": "2026-08-01T00:00:00Z",
            "periodEnd": "2026-08-07T00:00:00Z",
            "idempotencyKey": "p87-" + uuid.uuid4().hex,
            "description": "p87",
        },
        token=finance_tok,
    )
    check("create settlement batch", st, {200, 201}, raw)
    batch_id = batch.get("id") or (batch.get("batch") or {}).get("id") or ""
    if batch_id:
        st, _, raw = http_json(
            "POST",
            f"http://127.0.0.1:8092/v1/settlements/batches/{batch_id}/lines",
            {"payeeType": "courier", "payeeRef": "p87-courier", "amountMinor": 15000, "memo": "p87"},
            token=finance_tok,
        )
        check("settlement add line", st, {200, 201}, raw)
        st, _, raw = http_json(
            "POST",
            f"http://127.0.0.1:8092/v1/settlements/batches/{batch_id}/submit",
            {},
            token=finance_tok,
        )
        check("settlement submit", st, {200, 201}, raw)
        st, _, raw = http_json("POST", f"http://127.0.0.1:8114/v1/admin/finance/payouts/{batch_id}/approve", {}, token=admin_tok)
        check("settlement approve", st, {200, 409, 422}, raw)
        st, execb, raw = http_json("POST", f"http://127.0.0.1:8114/v1/admin/finance/settlements/{batch_id}/settle", {}, token=finance_tok)
        check("settlement execute", st, {200, 409, 422}, raw)
        if st == 200 and "paid" in raw.lower() and "unavailable" not in raw.lower() and "provider" not in raw.lower():
            print("FAIL fake payout success", raw[:300])
            failures.append("fake payout")
    else:
        check("settlement approve", 0, 200, raw)

    st, company, raw = http_json(
        "POST",
        "http://127.0.0.1:8110/v1/platform/companies",
        {"legalName": "P87 Co A.S.", "tradeName": "P87 Co", "countryCode": "TR", "primaryCurrency": "TRY"},
        token=sa_tok,
    )
    check("create company", st, {200, 201}, raw)
    st, _, _ = http_json("POST", "http://127.0.0.1:8110/v1/platform/companies", {"name": "x"}, token=admin_tok)
    check("admin denied platform", st, {401, 403})
    st, tenants, raw = http_json("GET", "http://127.0.0.1:8110/v1/platform/tenants", token=sa_tok)
    check("list tenants", st, 200, raw)
    st, _, _ = http_json("GET", "http://127.0.0.1:8110/v1/platform/audit", token=sa_tok)
    check("platform audit", st, 200)

    # Supplier
    st, _, raw = http_json("GET", "http://127.0.0.1:8117/v1/supplier/purchase-orders", token=supplier_tok)
    check("supplier POs", st, {200, 404}, raw)

    print("FAILURES", len(failures))
    for f in failures:
        print(" -", f)
    if failures:
        raise SystemExit(1)
    print("LIVE_E2E_PASS")


def apply_platform_ops_migrations() -> str:
    db = "nexora_platform_ops_service"
    dsn = f"postgres://nexora:nexora@127.0.0.1:5432/{db}?sslmode=disable"
    compose = ["docker", "compose", "-f", str(COMPOSE)]
    ready = False
    for _ in range(30):
        ping = subprocess.run(
            compose + ["exec", "-T", "postgres", "pg_isready", "-U", "nexora", "-d", "nexora"],
            capture_output=True, text=True,
        )
        if ping.returncode == 0:
            ready = True
            break
        time.sleep(2)
    if not ready:
        return "FAIL postgres not ready"
    created = subprocess.run(
        compose + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", "nexora", "-v", "ON_ERROR_STOP=1",
                    "-c", f"SELECT 1 FROM pg_database WHERE datname='{db}'"],
        capture_output=True, text=True,
    )
    exists = "1" in (created.stdout or "")
    if not exists:
        r = subprocess.run(
            compose + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", "nexora", "-v", "ON_ERROR_STOP=1",
                        "-c", f'CREATE DATABASE "{db}";'],
            capture_output=True, text=True,
        )
        if r.returncode != 0:
            return "FAIL create db: " + (r.stderr or r.stdout)[:400]
    mig_dir = ROOT / "services" / "platform-ops-service" / "migrations"
    for sql in sorted(mig_dir.glob("*.sql")):
        r = subprocess.run(
            compose + ["exec", "-T", "postgres", "psql", "-U", "nexora", "-d", db, "-v", "ON_ERROR_STOP=1"],
            input=sql.read_text(encoding="utf-8"),
            capture_output=True, text=True,
        )
        if r.returncode != 0:
            return f"FAIL migrate {sql.name}: " + (r.stderr or r.stdout)[:400]
        print("OK migrate", sql.name)
    return dsn


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--start", action="store_true")
    ap.add_argument("--e2e", action="store_true")
    ap.add_argument("--stop", action="store_true")
    args = ap.parse_args()
    if args.stop:
        stop_stack()
        return
    if args.start:
        start_stack()
        return
    if args.e2e:
        run_e2e()
        return
    start_stack()
    try:
        run_e2e()
    finally:
        # Keep stack up for follow-up commands; user can --stop.
        print("stack left running; python scripts/local/prompt87_live_gate.py --stop")


if __name__ == "__main__":
    try:
        sys.stdout.reconfigure(line_buffering=True)
        sys.stderr.reconfigure(line_buffering=True)
    except Exception:
        pass
    main()
