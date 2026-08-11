#!/usr/bin/env python3
"""Generate remaining custom screens for NEXORA customer app."""
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent / "lib" / "features"

SCREENS = {
    "cart/presentation/screens/cart_screen.dart": '''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../providers/cart_providers.dart';

class CartScreen extends ConsumerWidget {
  const CartScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final cartAsync = ref.watch(cartItemsStreamProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.cartTitle),
      body: AsyncValueWidget(
        value: cartAsync,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              message: l10n.emptyMessage,
              actionLabel: l10n.homeTitle,
              onAction: () => context.go(RouteNames.home),
            );
          }
          var subtotal = 0;
          for (final item in items) {
            subtotal += item.unitPriceMinor * item.quantity;
          }
          final total = Money(minorUnits: subtotal, currency: 'TRY');
          return Column(
            children: [
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  itemCount: items.length,
                  itemBuilder: (context, index) {
                    final item = items[index];
                    final price = Money(minorUnits: item.unitPriceMinor, currency: item.currency);
                    return NxCard(
                      child: Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(item.title, style: NxTypography.bodyMd),
                                NxPriceBlock(price: Formatters.money(price)),
                              ],
                            ),
                          ),
                          NxQtySelector(
                            quantity: item.quantity,
                            onIncrement: () => ref.read(cartRepositoryProvider).updateQuantity(
                                  item.productId,
                                  item.quantity + 1,
                                  variantId: item.variantId,
                                ),
                            onDecrement: () => ref.read(cartRepositoryProvider).updateQuantity(
                                  item.productId,
                                  item.quantity - 1,
                                  variantId: item.variantId,
                                ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ),
              NxCartBar(
                itemCount: items.fold(0, (s, i) => s + i.quantity),
                totalLabel: Formatters.money(total),
                onCheckout: () => context.push(RouteNames.checkoutAddress),
              ),
            ],
          );
        },
      ),
    );
  }
}
''',
    "search/presentation/screens/search_screen.dart": '''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../providers/search_providers.dart';

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({super.key});

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen> {
  final _controller = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final results = ref.watch(searchResultsProvider(_query));

    return Scaffold(
      appBar: NxTopBar(title: l10n.searchTitle),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: NxSearchField(
              controller: _controller,
              hint: l10n.searchHint,
              onChanged: (v) => setState(() => _query = v),
              onVoiceTap: () => _showVoiceSheet(context),
              onScanTap: () => context.push(RouteNames.barcodeScanner),
            ),
          ),
          Expanded(
            child: results.when(
              data: (items) => ListView.builder(
                itemCount: items.length,
                itemBuilder: (context, i) => ListTile(
                  title: Text(items[i].title),
                  onTap: () => context.push('/p/${items[i].id}'),
                ),
              ),
              loading: () => const Center(child: NxSpinner()),
              error: (e, _) => Center(child: Text(e.toString())),
            ),
          ),
        ],
      ),
    );
  }

  void _showVoiceSheet(BuildContext context) {
    NxSheet.show(
      context: context,
      title: AppLocalizations.of(context)!.voiceSearch,
      child: const Padding(
        padding: EdgeInsets.all(NxSpacing.s4),
        child: Text('Speak your search query'),
      ),
    );
  }
}
''',
    "search/presentation/providers/search_providers.dart": '''import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../../../di/providers.dart';
import '../../../home/presentation/providers/home_providers.dart';

final searchResultsProvider = FutureProvider.family.autoDispose<List<HomeProduct>, String>((ref, query) async {
  if (query.trim().length < 2) return [];
  final client = ref.watch(apiClientProvider);
  final result = await client.get<List<HomeProduct>>(
    '/search',
    queryParameters: {'q': query},
    parser: (json) => (json['items'] as List<dynamic>? ?? json as List<dynamic>)
        .map((e) => HomeProduct.fromJson(e as Map<String, dynamic>))
        .toList(),
  );
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
''',
    "search/presentation/screens/barcode_scanner_screen.dart": '''import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';

class BarcodeScannerScreen extends StatelessWidget {
  const BarcodeScannerScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: NxTopBar(title: l10n.barcodeScan),
      body: MobileScanner(
        onDetect: (capture) {
          final code = capture.barcodes.firstOrNull?.rawValue;
          if (code != null) context.pop(code);
        },
      ),
    );
  }
}
''',
    "categories/presentation/screens/category_detail_screen.dart": '''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../home/presentation/providers/home_providers.dart';
import '../providers/categories_providers.dart';

class CategoryDetailScreen extends ConsumerWidget {
  const CategoryDetailScreen({super.key, required this.categoryId});

  final String categoryId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(categoryProductsProvider(categoryId));
    return Scaffold(
      appBar: NxTopBar(title: categoryId),
      body: products.when(
        data: (items) => GridView.builder(
          padding: const EdgeInsets.all(NxSpacing.s4),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 2,
            mainAxisSpacing: NxSpacing.s3,
            crossAxisSpacing: NxSpacing.s3,
            childAspectRatio: 0.72,
          ),
          itemCount: items.length,
          itemBuilder: (context, i) {
            final p = items[i];
            return NxProductCard(
              title: p.title,
              price: '${p.priceMinor / 100}',
              onTap: () => context.push('/p/${p.id}'),
            );
          },
        ),
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}
''',
    "categories/presentation/providers/categories_providers.dart": '''import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../home/presentation/providers/home_providers.dart';

final categoryProductsProvider = FutureProvider.family.autoDispose<List<HomeProduct>, String>((ref, id) async {
  final client = ref.watch(apiClientProvider);
  final result = await client.get<List<HomeProduct>>(
    '/categories/$id/products',
    parser: (json) => (json['items'] as List<dynamic>? ?? [])
        .map((e) => HomeProduct.fromJson(e as Map<String, dynamic>))
        .toList(),
  );
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
''',
    "product/presentation/screens/product_screen.dart": '''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../cart/presentation/providers/cart_providers.dart';
import '../providers/product_providers.dart';

class ProductScreen extends ConsumerWidget {
  const ProductScreen({super.key, required this.productId});

  final String productId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final productAsync = ref.watch(productDetailProvider(productId));

    return Scaffold(
      body: productAsync.when(
        data: (p) {
          final price = Money(minorUnits: p.priceMinor, currency: p.currency);
          return CustomScrollView(
            slivers: [
              SliverAppBar(
                expandedHeight: 280,
                pinned: true,
                flexibleSpace: FlexibleSpaceBar(
                  background: p.imageUrl != null
                      ? Image.network(p.imageUrl!, fit: BoxFit.cover)
                      : ColoredBox(color: context.nxColors.bgSurfaceRaised),
                ),
              ),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(p.title, style: NxTypography.headlineMd),
                      const SizedBox(height: NxSpacing.s2),
                      NxPriceBlock(price: Formatters.money(price)),
                      if (p.description != null) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text(p.description!, style: NxTypography.bodyMd),
                      ],
                      const SizedBox(height: NxSpacing.s6),
                      NxButton(
                        label: l10n.addToCart,
                        expand: true,
                        onPressed: () => ref.read(cartRepositoryProvider).addItem(
                              productId: p.id,
                              title: p.title,
                              imageUrl: p.imageUrl,
                              unitPriceMinor: p.priceMinor,
                              currency: p.currency,
                            ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}
''',
    "product/presentation/providers/product_providers.dart": '''import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';

class ProductDetail {
  const ProductDetail({
    required this.id,
    required this.title,
    required this.priceMinor,
    this.currency = 'TRY',
    this.imageUrl,
    this.description,
  });

  final String id;
  final String title;
  final int priceMinor;
  final String currency;
  final String? imageUrl;
  final String? description;

  factory ProductDetail.fromJson(Map<String, dynamic> json) => ProductDetail(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        priceMinor: (json['price_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
        imageUrl: json['image_url']?.toString(),
        description: json['description']?.toString(),
      );
}

final productDetailProvider = FutureProvider.family.autoDispose<ProductDetail, String>((ref, id) async {
  final client = ref.watch(apiClientProvider);
  final result = await client.get<ProductDetail>(
    '/products/$id',
    parser: (json) => ProductDetail.fromJson(json as Map<String, dynamic>),
  );
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
''',
}

CHECKOUT = '''import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';

class {class_name} extends StatelessWidget {{
  const {class_name}({{super.key}});

  @override
  Widget build(BuildContext context) {{
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: NxTopBar(title: l10n.checkoutTitle),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('{step_title}', style: NxTypography.headlineSm),
            const Spacer(),
            NxButton(
              label: l10n.continueLabel,
              expand: true,
              onPressed: () => context.push('{next_route}'),
            ),
          ],
        ),
      ),
    );
  }}
}}
'''

for path_suffix, content in SCREENS.items():
    write_path = ROOT.parent / "features" / path_suffix.replace("/", "\\").replace("\\", "/")
    # fix path
    write_path = Path(str(ROOT.parent / "features" / path_suffix))
    write_path.parent.mkdir(parents=True, exist_ok=True)
    write_path.write_text(content, encoding="utf-8")

checkout_steps = [
    ("checkout/presentation/screens/checkout_address_screen.dart", "CheckoutAddressScreen", "Delivery address", "/checkout/schedule"),
]
# manual checkout screens
steps = [
    ("checkout_address_screen.dart", "CheckoutAddressScreen", "Select delivery address", "/checkout/schedule"),
    ("checkout_schedule_screen.dart", "CheckoutScheduleScreen", "Schedule delivery", "/checkout/payment"),
    ("checkout_payment_screen.dart", "CheckoutPaymentScreen", "Payment method", "/checkout/review"),
    ("checkout_review_screen.dart", "CheckoutReviewScreen", "Review order", "/orders"),
]
for fname, cls, title, next_route in steps:
    p = ROOT / "checkout" / "presentation" / "screens" / fname
    p.write_text(CHECKOUT.format(class_name=cls, step_title=title, next_route=next_route), encoding="utf-8")

# profile account screen
(ROOT / "profile" / "presentation" / "screens" / "account_screen.dart").write_text('''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../auth/presentation/providers/auth_controller.dart';
import '../../../auth/presentation/providers/auth_session_provider.dart';

class AccountScreen extends ConsumerWidget {
  const AccountScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final session = ref.watch(authSessionProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.accountTitle),
      body: ListView(
        children: [
          ListTile(
            title: Text(session.displayName ?? l10n.guestContinue),
            subtitle: Text(session.isAuthenticated ? (session.phone ?? session.email ?? '') : l10n.guestContinue),
          ),
          _tile(context, l10n.ordersTitle, RouteNames.orders),
          _tile(context, l10n.favoritesTitle, RouteNames.favorites),
          _tile(context, l10n.walletTitle, RouteNames.wallet),
          _tile(context, l10n.addressesTitle, RouteNames.addresses),
          _tile(context, l10n.settingsTitle, RouteNames.settings),
          _tile(context, l10n.supportTitle, RouteNames.support),
          if (session.isAuthenticated)
            ListTile(
              title: Text(l10n.signIn, style: TextStyle(color: context.nxColors.danger)),
              onTap: () => ref.read(authControllerProvider.notifier).signOut(),
            )
          else
            ListTile(
              title: Text(l10n.signIn),
              onTap: () => context.push(RouteNames.auth),
            ),
        ],
      ),
    );
  }

  Widget _tile(BuildContext context, String title, String route) {
    return ListTile(title: Text(title), trailing: const Icon(Icons.chevron_right), onTap: () => context.push(route));
  }
}
''', encoding="utf-8")

# orders detail + tracking
(ROOT / "orders" / "presentation" / "screens" / "order_detail_screen.dart").write_text('''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../providers/orders_providers.dart';

class OrderDetailScreen extends ConsumerWidget {
  const OrderDetailScreen({super.key, required this.orderId});

  final String orderId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final order = ref.watch(orderDetailProvider(orderId));
    return Scaffold(
      appBar: NxTopBar(title: 'Order $orderId'),
      body: order.when(
        data: (o) => Padding(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              NxOrderCard(
                orderId: o.id,
                status: o.status,
                total: o.totalLabel,
                itemCount: o.itemCount,
              ),
              const SizedBox(height: NxSpacing.s4),
              NxButton(
                label: 'Track',
                expand: true,
                onPressed: () => context.push('/orders/$orderId/track'),
              ),
            ],
          ),
        ),
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}
''', encoding="utf-8")

(ROOT / "orders" / "presentation" / "providers" / "orders_providers.dart").write_text('''import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';

class OrderSummary {
  const OrderSummary({required this.id, required this.status, required this.totalLabel, required this.itemCount});
  final String id;
  final String status;
  final String totalLabel;
  final int itemCount;

  factory OrderSummary.fromJson(Map<String, dynamic> json) => OrderSummary(
        id: json['id']?.toString() ?? '',
        status: json['status']?.toString() ?? '',
        totalLabel: json['total_label']?.toString() ?? '',
        itemCount: (json['item_count'] as num?)?.toInt() ?? 0,
      );
}

final orderDetailProvider = FutureProvider.family.autoDispose<OrderSummary, String>((ref, id) async {
  final client = ref.watch(apiClientProvider);
  final result = await client.get<OrderSummary>(
    '/orders/$id',
    parser: (json) => OrderSummary.fromJson(json as Map<String, dynamic>),
  );
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
''', encoding="utf-8")

(ROOT / "tracking" / "presentation" / "screens" / "tracking_screen.dart").write_text('''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../providers/tracking_providers.dart';

class TrackingScreen extends ConsumerWidget {
  const TrackingScreen({super.key, required this.orderId});

  final String orderId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tracking = ref.watch(trackingProvider(orderId));
    return Scaffold(
      appBar: NxTopBar(title: 'Track order'),
      body: tracking.when(
        data: (t) => Column(
          children: [
            NxEtaCard(minutes: t.etaMinutes, label: 'Estimated arrival'),
            Expanded(
              child: NxTrackingTimeline(steps: t.steps),
            ),
            Padding(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Row(
                children: [
                  Expanded(child: NxButton(label: 'Call courier', variant: NxButtonVariant.secondary, onPressed: () {})),
                  const SizedBox(width: NxSpacing.s3),
                  Expanded(child: NxButton(label: 'Chat', onPressed: () {})),
                ],
              ),
            ),
          ],
        ),
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}
''', encoding="utf-8")

(ROOT / "tracking" / "presentation" / "providers" / "tracking_providers.dart").write_text('''import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';

class TrackingViewModel {
  const TrackingViewModel({required this.etaMinutes, required this.steps});
  final int etaMinutes;
  final List<NxTrackingStep> steps;
}

final trackingProvider = FutureProvider.family.autoDispose<TrackingViewModel, String>((ref, orderId) async {
  final client = ref.watch(apiClientProvider);
  final result = await client.get<Map<String, dynamic>>('/orders/$orderId/tracking', parser: (j) => j as Map<String, dynamic>);
  return result.fold(
    onSuccess: (json) {
      final steps = (json['steps'] as List<dynamic>? ?? [])
          .map((e) => NxTrackingStep(
                title: e['title']?.toString() ?? '',
                subtitle: e['subtitle']?.toString(),
                state: NxTrackingStepState.values.firstWhere(
                  (s) => s.name == (e['state']?.toString() ?? 'pending'),
                  orElse: () => NxTrackingStepState.pending,
                ),
              ))
          .toList();
      return TrackingViewModel(
        etaMinutes: (json['eta_minutes'] as num?)?.toInt() ?? 0,
        steps: steps,
      );
    },
    onFailure: (e) => throw e,
  );
});
''', encoding="utf-8")

(ROOT / "settings" / "presentation" / "screens" / "settings_screen.dart").write_text('''import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final themeMode = ref.watch(themeModeProvider);
    final locale = ref.watch(localeCodeProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.settingsTitle),
      body: ListView(
        children: [
          ListTile(
            title: Text(l10n.language),
            subtitle: Text(locale),
            onTap: () => ref.read(localeCodeProvider.notifier).state = locale == 'en' ? 'tr' : 'en',
          ),
          ListTile(
            title: Text(l10n.theme),
            subtitle: Text(themeMode.name),
            onTap: () {
              final next = switch (themeMode) {
                ThemeMode.system => ThemeMode.light,
                ThemeMode.light => ThemeMode.dark,
                ThemeMode.dark => ThemeMode.system,
              };
              ref.read(themeModeProvider.notifier).state = next;
            },
          ),
          ListTile(title: Text(l10n.biometricUnlock)),
          ListTile(title: Text(l10n.privacy)),
          ListTile(title: Text(l10n.security)),
          ListTile(title: Text(l10n.deleteAccount), textColor: context.nxColors.danger),
        ],
      ),
    );
  }
}
''', encoding="utf-8")

# Fix coupons screen to accept promoCode
coupons_path = ROOT / "coupons" / "presentation" / "screens" / "coupons_screen.dart"
if coupons_path.exists():
    t = coupons_path.read_text(encoding="utf-8")
    t = t.replace("class CouponsScreen extends ConsumerWidget {", "class CouponsScreen extends ConsumerWidget {\n  const CouponsScreen({super.key, this.promoCode});\n\n  final String? promoCode;")
    t = t.replace("const CouponsScreen({super.key});", "")
    coupons_path.write_text(t, encoding="utf-8")

print("Extra screens generated")
