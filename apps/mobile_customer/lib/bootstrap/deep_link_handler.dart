import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../routing/app_router.dart';

/// Maps universal / app links → GoRouter locations (CONSTITUTION §47, DS §13).
class DeepLinkHandler {
  DeepLinkHandler(this._router);

  final GoRouter _router;
  StreamSubscription<Uri>? _sub;
  final _appLinks = AppLinks();

  Future<void> start() async {
    try {
      final initial = await _appLinks.getInitialLink();
      if (initial != null) {
        _navigate(initial);
      }
    } catch (_) {
      // Ignore cold-start link failures.
    }

    _sub = _appLinks.uriLinkStream.listen(
      _navigate,
      onError: (_) {},
    );
  }

  void _navigate(Uri uri) {
    final location = mapUriToLocation(uri);
    if (location == null) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _router.go(location);
    });
  }

  /// Public for unit tests.
  static String? mapUriToLocation(Uri uri) {
    final segments = uri.pathSegments;
    if (uri.scheme == 'nexora' && uri.host == 'product' && segments.isNotEmpty) {
      return '/p/${segments.first}';
    }
    if (segments.isEmpty) return null;

    switch (segments.first) {
      case 'p':
      case 'product':
        if (segments.length >= 2) return '/p/${segments[1]}';
        return null;
      case 'c':
        if (segments.length >= 2) return '/c/${segments[1]}';
        return null;
      case 'promo':
        if (segments.length >= 2) return '/promo/${segments[1]}';
        return null;
      case 'orders':
        if (segments.length >= 2) {
          if (segments.length >= 3 && segments[2] == 'track') {
            return '/orders/${segments[1]}/track';
          }
          return '/orders/${segments[1]}';
        }
        return '/orders';
      case 'cart':
        return '/cart';
      case 'search':
        return '/search';
      default:
        return null;
    }
  }

  Future<void> dispose() async {
    await _sub?.cancel();
  }
}

final deepLinkHandlerProvider = Provider<DeepLinkHandler>((ref) {
  final router = ref.watch(appRouterProvider);
  final handler = DeepLinkHandler(router);
  ref.onDispose(handler.dispose);
  return handler;
});
