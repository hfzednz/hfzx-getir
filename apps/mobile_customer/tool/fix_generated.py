#!/usr/bin/env python3
"""Fix generated code issues across feature modules."""
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent / "lib" / "features"

# 1. Fix broken entity imports in remote datasources
for ds in ROOT.glob("*/data/datasources/*_remote_datasource.dart"):
    feat = ds.parent.parent.parent.name
    text = ds.read_text(encoding="utf-8")
    # Remove wrong imports
    text = "\n".join(
        line for line in text.splitlines()
        if "features_entity.dart" not in line
    )
    entity_import = f"import '../../domain/entities/{feat}_entity.dart';"
    if entity_import not in text:
        text = text.replace(
            "import '../models/",
            f"{entity_import}\nimport '../models/",
        )
    ds.write_text(text + "\n", encoding="utf-8")

# 2. Fix NxEmptyState API in screens
for screen in ROOT.glob("*/presentation/screens/*_screen.dart"):
    text = screen.read_text(encoding="utf-8")
    orig = text
    text = text.replace("message: l10n.emptyMessage", "body: l10n.emptyMessage")
    text = text.replace("actionLabel: l10n.retry", "primaryActionLabel: l10n.retry")
    text = text.replace("actionLabel: l10n.homeTitle", "primaryActionLabel: l10n.homeTitle")
    text = text.replace("onAction: () => ref.invalidate(", "onPrimaryAction: () => ref.invalidate(")
    text = text.replace("onAction: () => context.go(", "onPrimaryAction: () => context.go(")
    text = text.replace("AppLocalizations.of(context)!", "AppLocalizations.of(context)")
    if text != orig:
        screen.write_text(text, encoding="utf-8")

print("Fixes applied")
