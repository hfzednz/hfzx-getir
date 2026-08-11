#!/usr/bin/env python3
"""Generate NEXORA customer app feature module boilerplate."""
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent / "lib" / "features"

FEATURES = {
    "splash": {"entity": "SplashState", "api_get": None, "api_post": None},
    "onboarding": {"entity": "OnboardingPage", "api_get": None},
    "auth": {"entity": "AuthSession", "api_get": "/auth/session", "api_post": "/auth/otp/request"},
    "city": {"entity": "CityContext", "api_get": "/cities/serviceability"},
    "home": {"entity": "HomeFeed", "api_get": "/home"},
    "search": {"entity": "SearchResult", "api_get": "/search"},
    "categories": {"entity": "Category", "api_get": "/categories"},
    "product": {"entity": "Product", "api_get": "/products/{id}"},
    "cart": {"entity": "Cart", "api_get": "/cart", "api_post": "/cart/items"},
    "checkout": {"entity": "CheckoutSession", "api_get": "/checkout/sessions", "api_post": "/checkout/confirm"},
    "orders": {"entity": "Order", "api_get": "/orders"},
    "tracking": {"entity": "TrackingState", "api_get": "/orders/{id}/tracking"},
    "favorites": {"entity": "FavoriteItem", "api_get": "/favorites"},
    "wallet": {"entity": "Wallet", "api_get": "/wallet"},
    "coupons": {"entity": "Coupon", "api_get": "/coupons"},
    "loyalty": {"entity": "LoyaltyAccount", "api_get": "/loyalty"},
    "notifications": {"entity": "AppNotification", "api_get": "/notifications"},
    "profile": {"entity": "UserProfile", "api_get": "/profile", "api_post": "/profile"},
    "addresses": {"entity": "Address", "api_get": "/addresses", "api_post": "/addresses"},
    "reviews": {"entity": "Review", "api_get": None, "api_post": "/orders/{id}/review"},
    "support": {"entity": "SupportTicket", "api_get": "/support/tickets", "api_post": "/support/tickets"},
    "settings": {"entity": "AppSettings", "api_get": None},
    "referral": {"entity": "ReferralInfo", "api_get": "/referral"},
    "help": {"entity": "HelpArticle", "api_get": "/help"},
    "about": {"entity": "AboutInfo", "api_get": None},
    "legal": {"entity": "LegalDocument", "api_get": "/legal/{doc}"},
}

def snake_to_pascal(s: str) -> str:
    return "".join(w.capitalize() for w in s.split("_"))

def write(path: Path, content: str):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")

def gen_entity(name: str, entity: str) -> str:
    return f"""import 'package:equatable/equatable.dart';

class {entity} extends Equatable {{
  const {entity}({{
    required this.id,
    this.title = '',
    this.payload = const {{}},
  }});

  final String id;
  final String title;
  final Map<String, dynamic> payload;

  factory {entity}.fromJson(Map<String, dynamic> json) => {entity}(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        payload: Map<String, dynamic>.from(json),
      );

  Map<String, dynamic> toJson() => {{'id': id, 'title': title, ...payload}};

  @override
  List<Object?> get props => [id, title, payload];
}}
"""

def gen_repo_interface(name: str, entity: str) -> str:
    return f"""import 'package:nexora_core/nexora_core.dart';

import '../entities/{name}_entity.dart';

abstract class {snake_to_pascal(name)}Repository {{
  Future<Result<{entity}>> fetch({{String? id}});
  Future<Result<List<{entity}>>> fetchList({{Map<String, dynamic>? params}});
  Future<Result<{entity}>> mutate({{required Map<String, dynamic> body, String? idempotencyKey}});
}}
"""

def gen_usecase(name: str, entity: str) -> str:
    pascal = snake_to_pascal(name)
    return f"""import 'package:nexora_core/nexora_core.dart';

import '../entities/{name}_entity.dart';
import '../repositories/{name}_repository.dart';

class Get{pascal}UseCase {{
  const Get{pascal}UseCase(this._repository);
  final {pascal}Repository _repository;

  Future<Result<{entity}>> call({{String? id}}) => _repository.fetch(id: id);
}}

class List{pascal}UseCase {{
  const List{pascal}UseCase(this._repository);
  final {pascal}Repository _repository;

  Future<Result<List<{entity}>>> call({{Map<String, dynamic>? params}}) =>
      _repository.fetchList(params: params);
}}
"""

def gen_remote_ds(name: str, cfg: dict) -> str:
    api_get = cfg.get("api_get") or f"/{name}"
    api_post = cfg.get("api_post") or f"/{name}"
    entity = cfg["entity"]
    return f"""import 'package:nexora_core/nexora_core.dart';

import '../models/{name}_model.dart';

class {snake_to_pascal(name)}RemoteDataSource {{
  const {snake_to_pascal(name)}RemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '{api_get}';
  static const _mutatePath = '{api_post}';

  Future<Result<{entity}>> fetch({{String? id}}) async {{
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<{entity}>(
      path,
      parser: (json) => {entity}Model.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }}

  Future<Result<List<{entity}>>> fetchList({{Map<String, dynamic>? params}}) async {{
    return _client.get<List<{entity}>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => {entity}Model.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }}

  Future<Result<{entity}>> mutate({{
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }}) async {{
    return _client.post<{entity}>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => {entity}Model.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }}
}}
"""

def gen_model(name: str, entity: str) -> str:
    return f"""import '../../domain/entities/{name}_entity.dart';

class {entity}Model {{
  const {entity}Model({{required this.id, required this.title, required this.raw}});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory {entity}Model.fromJson(Map<String, dynamic> json) => {entity}Model(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  {entity} toEntity() => {entity}(id: id, title: title, payload: raw);
}}
"""

def gen_repo_impl(name: str, entity: str) -> str:
    pascal = snake_to_pascal(name)
    return f"""import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/{name}_entity.dart';
import '../../domain/repositories/{name}_repository.dart';
import '../datasources/{name}_remote_datasource.dart';

class {pascal}RepositoryImpl implements {pascal}Repository {{
  const {pascal}RepositoryImpl(this._remote);
  final {pascal}RemoteDataSource _remote;

  @override
  Future<Result<{entity}>> fetch({{String? id}}) => _remote.fetch(id: id);

  @override
  Future<Result<List<{entity}>>> fetchList({{Map<String, dynamic>? params}}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<{entity}>> mutate({{
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }}) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}}
"""

def gen_providers(name: str, entity: str) -> str:
    pascal = snake_to_pascal(name)
    return f"""import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/{name}_remote_datasource.dart';
import '../../data/repositories/{name}_repository_impl.dart';
import '../../domain/entities/{name}_entity.dart';
import '../../domain/repositories/{name}_repository.dart';
import '../../domain/usecases/{name}_usecases.dart';

final {name}RemoteDataSourceProvider = Provider<{pascal}RemoteDataSource>((ref) {{
  return {pascal}RemoteDataSource(ref.watch(apiClientProvider));
}});

final {name}RepositoryProvider = Provider<{pascal}Repository>((ref) {{
  return {pascal}RepositoryImpl(ref.watch({name}RemoteDataSourceProvider));
}});

final get{pascal}UseCaseProvider = Provider((ref) =>
    Get{pascal}UseCase(ref.watch({name}RepositoryProvider)));

final list{pascal}UseCaseProvider = Provider((ref) =>
    List{pascal}UseCase(ref.watch({name}RepositoryProvider)));

final {name}ListProvider = FutureProvider.autoDispose<List<{entity}>>((ref) async {{
  final result = await ref.watch(list{pascal}UseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
}});
"""

def gen_screen(name: str, entity: str) -> str:
    pascal = snake_to_pascal(name)
    return f"""import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/{name}_providers.dart';

class {pascal}Screen extends ConsumerWidget {{
  const {pascal}Screen({{super.key, this.id}});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {{
    final l10n = AppLocalizations.of(context)!;
    final asyncItems = ref.watch({name}ListProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.{name}Title),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {{
          if (items.isEmpty) {{
            return NxEmptyState(
              title: l10n.emptyTitle,
              message: l10n.emptyMessage,
              actionLabel: l10n.retry,
              onAction: () => ref.invalidate({name}ListProvider),
            );
          }}
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate({name}ListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {{
                final item = items[index];
                return NxCard(
                  child: ListTile(
                    title: Text(item.title.isNotEmpty ? item.title : item.id),
                    subtitle: Text(item.id),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () {{}},
                  ),
                );
              }},
            ),
          );
        }},
        error: (e, _) => ErrorView(
          message: e.toString(),
          onRetry: () => ref.invalidate({name}ListProvider),
        ),
      ),
    );
  }}
}}
"""

for feat, cfg in FEATURES.items():
    entity = cfg["entity"]
    base = ROOT / feat
    write(base / "domain" / "entities" / f"{feat}_entity.dart", gen_entity(feat, entity))
    write(base / "domain" / "repositories" / f"{feat}_repository.dart", gen_repo_interface(feat, entity))
    write(base / "domain" / "usecases" / f"{feat}_usecases.dart", gen_usecase(feat, entity))
    write(base / "data" / "datasources" / f"{feat}_remote_datasource.dart", gen_remote_ds(feat, cfg))
    write(base / "data" / "models" / f"{feat}_model.dart", gen_model(feat, entity))
    write(base / "data" / "repositories" / f"{feat}_repository_impl.dart", gen_repo_impl(feat, entity))
    write(base / "presentation" / "providers" / f"{feat}_providers.dart", gen_providers(feat, entity))
    if feat not in ("shell",):
        write(base / "presentation" / "screens" / f"{feat}_screen.dart", gen_screen(feat, entity))

print(f"Generated {len(FEATURES)} feature modules")
