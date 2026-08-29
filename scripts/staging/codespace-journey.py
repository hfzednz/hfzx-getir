import json
import re
import subprocess
import urllib.request
import uuid

TENANT = "11111111-1111-1111-1111-111111111111"
BFF = "http://127.0.0.1:3000"
SKU = "70fe49fa-7a60-4b7d-ae16-090116e8acbb"


def call(method, path, body=None, token=None, expect_json=True):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json", "X-Tenant-Id": TENANT}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(BFF + path, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            raw = resp.read().decode()
            obj = {}
            if expect_json and raw.lstrip().startswith("{"):
                obj = json.loads(raw)
            print(method, path, resp.status)
            return resp.status, obj, raw
    except urllib.error.HTTPError as e:
        raw = e.read().decode()[:400]
        print(method, path, "FAIL", e.code, raw.replace("\n", " "))
        return e.code, {}, raw


print("login", call("GET", "/login", expect_json=False)[0])
_, start, _ = call("POST", "/v1/customer/auth/otp/start", {"phone": "+905551112233"})
cid = start.get("challengeId")
print("challenge", bool(cid))
logs = subprocess.check_output(
    [
        "bash",
        "-lc",
        "docker logs nexora-staging-identity-service 2>&1 | grep otp.dev_mode | grep +905551112233 | tail -1",
    ],
    text=True,
)
m = re.search(r'"code":"(\d+)"', logs)
otp = m.group(1) if m else ""
print("otp_digits", len(otp))
_, sess, _ = call("POST", "/v1/customer/auth/otp/verify", {"challengeId": cid, "code": otp})
token = sess.get("accessToken") or sess.get("AccessToken") or ""
principal = sess.get("customerId") or sess.get("CustomerID") or ""
print("token", bool(token), "principal", bool(principal))
_, home, _ = call("GET", "/v1/customer/home?lat=41&lng=29", token=token)
print("products", len(home.get("products") or []))
cart_id = str(uuid.uuid4())
_, cart, _ = call(
    "POST",
    "/v1/customer/cart/items",
    {"cartId": cart_id, "sku": SKU, "qty": 1, "unitMinor": 1500},
    token=token,
)
cart_id = cart.get("cartId") or cart.get("ID") or cart.get("id") or cart_id
print("cart", bool(cart_id), list(cart)[:8])
_, prev, _ = call(
    "POST",
    "/v1/customer/checkout/preview",
    {"cartId": cart_id, "principalId": principal},
    token=token,
)
sid = prev.get("sessionId") or ""
print("preview sid", bool(sid), "total", prev.get("totalMinor"), "ready", prev.get("paymentReady"), prev)
_, placed, _ = call(
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
print("place", placed)
if oid:
    st, order, _ = call("GET", "/v1/customer/orders/" + oid, token=token)
    print("order", st, list(order)[:8] if order else None)
    st, track, _ = call("GET", "/v1/customer/orders/" + oid + "/track", token=token)
    print("track", st, track)
print("JOURNEY_DONE")
