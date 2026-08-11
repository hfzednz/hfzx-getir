import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/auth/presentation/providers/auth_session_provider.dart';
import 'route_names.dart';

String? authGuard(BuildContext context, GoRouterState state, Ref ref) {
  final path = state.matchedLocation;
  final requiresAuth = authRequiredPaths.any(
    (p) => path == p || path.startsWith('$p/'),
  );
  if (!requiresAuth) return null;

  final session = ref.read(authSessionProvider);
  if (session.isAuthenticated || session.isGuestCheckoutAllowed) {
    return null;
  }

  final redirect = Uri.encodeComponent(state.uri.toString());
  return '${RouteNames.auth}?redirect=$redirect';
}
