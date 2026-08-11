import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/auth/presentation/providers/auth_session_provider.dart';
import 'route_names.dart';

String? authGuard(BuildContext context, GoRouterState state, Ref ref) {
  final session = ref.read(authSessionProvider);
  final path = state.uri.path;
  final needsAuth = authRequiredPrefixes.any(
    (p) => path == p || path.startsWith('$p/'),
  );

  if (needsAuth && !session.isAuthenticated) {
    return RouteNames.authPhone;
  }
  if (session.isAuthenticated &&
      (path == RouteNames.auth ||
          path.startsWith('${RouteNames.auth}/') ||
          path == RouteNames.splash)) {
    return RouteNames.home;
  }
  return null;
}
