#!/usr/bin/env python3
"""Repeatable bilingual catalog + store seed for the Flutter customer app.

Creates multiple inventory warehouses (stores), catalog categories, and products
with TR/EN locales. Does not seed secrets. Safe to re-run: slugs are unique per
run suffix unless SEED_SUFFIX is pinned.
"""
from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request

TENANT = os.environ.get("TENANT", "11111111-1111-1111-1111-111111111111")
CATALOG = os.environ.get("CATALOG_URL", "http://127.0.0.1:8083").rstrip("/")
INVENTORY = os.environ.get("INVENTORY_URL", "http://127.0.0.1:8084").rstrip("/")
SUFFIX = os.environ.get("SEED_SUFFIX", "mobile")


def call(method: str, url: str, payload: dict | None = None) -> dict:
    data = None if payload is None else json.dumps(payload).encode()
    req = urllib.request.Request(
        url,
        data=data,
        method=method,
        headers={
            "Content-Type": "application/json",
            "X-Tenant-Id": TENANT,
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as res:
            raw = res.read().decode() or "{}"
            return json.loads(raw)
    except urllib.error.HTTPError as exc:
        body = exc.read().decode(errors="replace")
        print(f"WARN {method} {url} -> {exc.code} {body[:300]}", file=sys.stderr)
        return {}
    except OSError as exc:
        print(f"WARN {method} {url} -> {exc}", file=sys.stderr)
        return {}


STORES = [
    {"code": f"KDK-{SUFFIX}", "name": "Nexora Market Kadıköy", "timezone": "Europe/Istanbul"},
    {"code": f"BJK-{SUFFIX}", "name": "Nexora Market Beşiktaş", "timezone": "Europe/Istanbul"},
    {"code": f"BKR-{SUFFIX}", "name": "Nexora Market Bakırköy", "timezone": "Europe/Istanbul"},
]

CATEGORIES = [
    ("Süt & Kahvaltı", "sut-kahvalti"),
    ("Ekmek & Fırın", "ekmek-firin"),
    ("İçecek", "icecek"),
    ("Su", "su"),
    ("Kahve & Çay", "kahve-cay"),
    ("Atıştırmalık", "atistirmalik"),
    ("Çikolata", "cikolata"),
    ("Meyve", "meyve"),
    ("Sebze", "sebze"),
    ("Et", "et"),
    ("Tavuk", "tavuk"),
    ("Peynir", "peynir"),
    ("Yoğurt", "yogurt"),
    ("Yumurta", "yumurta"),
    ("Temel Gıda", "temel-gida"),
    ("Pirinç", "pirinc"),
    ("Makarna", "makarna"),
    ("Un", "un"),
    ("Şeker", "seker"),
    ("Konserve", "konserve"),
    ("Sos", "sos"),
    ("Dondurulmuş", "dondurulmus"),
    ("Temizlik", "temizlik"),
    ("Kağıt Ürünleri", "kagit-urunleri"),
    ("Kişisel Bakım", "kisisel-bakim"),
    ("Bebek", "bebek"),
    ("Ev & Yaşam", "ev-yasam"),
    ("Evcil Hayvan", "evcil-hayvan"),
]

PRODUCTS = [
    ("fresh-milk", "MILK-1L", "sut-kahvalti", "Taze Süt", "Fresh Milk", "1 litre tam yağlı süt", "1 litre whole milk", "Nexora"),
    ("village-bread", "BREAD-1", "ekmek-firin", "Köy Ekmeği", "Village Bread", "Taze fırın ekmeği", "Fresh baked bread", "Fırın"),
    ("strained-yogurt", "YOG-1", "yogurt", "Süzme Yoğurt", "Strained Yogurt", "Kremamsı süzme yoğurt", "Creamy strained yogurt", "Nexora"),
    ("sparkling-water", "WATER-1", "su", "Maden Suyu", "Sparkling Water", "6x200ml cam şişe", "6x200ml glass bottles", "Nexora"),
    ("filter-coffee", "COFFEE-1", "kahve-cay", "Filtre Kahve", "Filter Coffee", "250g çekirdek", "250g beans", "Nexora"),
    ("dark-chocolate", "CHOC-1", "cikolata", "Bitter Çikolata", "Dark Chocolate", "80g tablet", "80g bar", "Nexora"),
    ("bananas", "FRUIT-1", "meyve", "Muz", "Bananas", "1 kg", "1 kg", "Nexora"),
    ("tomato", "VEG-1", "sebze", "Domates", "Tomato", "1 kg", "1 kg", "Nexora"),
    ("chicken-breast", "CHICK-1", "tavuk", "Tavuk Göğsü", "Chicken Breast", "500g", "500g", "Nexora"),
    ("white-cheese", "CHEESE-1", "peynir", "Beyaz Peynir", "White Cheese", "500g", "500g", "Nexora"),
    ("eggs", "EGG-1", "yumurta", "Yumurta", "Eggs", "15'li kutu", "Box of 15", "Nexora"),
    ("rice", "RICE-1", "pirinc", "Pirinç", "Rice", "1 kg", "1 kg", "Nexora"),
    ("pasta", "PASTA-1", "makarna", "Makarna", "Pasta", "500g", "500g", "Nexora"),
    ("flour", "FLOUR-1", "un", "Un", "Flour", "1 kg", "1 kg", "Nexora"),
    ("sugar", "SUGAR-1", "seker", "Şeker", "Sugar", "1 kg", "1 kg", "Nexora"),
    ("dish-soap", "CLEAN-1", "temizlik", "Bulaşık Deterjanı", "Dish Soap", "750ml", "750ml", "Nexora"),
    ("baby-wipes", "BABY-1", "bebek", "Bebek Mendili", "Baby Wipes", "72'li", "72 pack", "Nexora"),
    ("cat-food", "PET-1", "evcil-hayvan", "Kedi Maması", "Cat Food", "1 kg", "1 kg", "Nexora"),
]


def pick_id(payload: dict, *keys: str) -> str:
    node = payload.get("product") or payload.get("category") or payload.get("warehouse") or payload
    for key in keys:
        val = node.get(key) if isinstance(node, dict) else None
        if val:
            return str(val)
        val = payload.get(key)
        if val:
            return str(val)
    return ""


def main() -> int:
    print(f"seed-mobile-catalog tenant={TENANT} catalog={CATALOG} inventory={INVENTORY}")
    for store in STORES:
        out = call("POST", f"{INVENTORY}/v1/inventory/warehouses", store)
        print(f"store {store['code']} id={pick_id(out, 'id', 'ID') or 'n/a'}")

    cat_ids: dict[str, str] = {}
    for name, slug in CATEGORIES:
        out = call(
            "POST",
            f"{CATALOG}/v1/catalog/categories",
            {"name": name, "slug": f"{slug}-{SUFFIX}", "kind": "standard"},
        )
        cid = pick_id(out, "id", "ID")
        cat_ids[slug] = cid
        print(f"category {slug} id={cid or 'n/a'}")

    for slug, sku, cat, tr, en, trd, end, brand in PRODUCTS:
        out = call(
            "POST",
            f"{CATALOG}/v1/catalog/products",
            {"kind": "standard", "slug": f"{slug}-{SUFFIX}", "skuCode": f"{sku}-{SUFFIX}"},
        )
        pid = pick_id(out, "id", "ID")
        if not pid:
            continue
        call("PUT", f"{CATALOG}/v1/catalog/products/{pid}/locales/tr", {"title": tr, "description": trd})
        call("PUT", f"{CATALOG}/v1/catalog/products/{pid}/locales/en", {"title": en, "description": end})
        call("POST", f"{CATALOG}/v1/catalog/products/{pid}/variants", {"skuCode": f"{sku}-UNIT", "name": "Default"})
        cid = cat_ids.get(cat)
        if cid:
            call(
                "POST",
                f"{CATALOG}/v1/catalog/products/{pid}/categories",
                {"categoryId": cid, "isPrimary": True, "sortOrder": 0},
            )
        print(f"product {sku} id={pid} tr={tr} en={en}")

    call("POST", f"{CATALOG}/v1/catalog/search/reindex", {})
    print("OK mobile catalog seed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
