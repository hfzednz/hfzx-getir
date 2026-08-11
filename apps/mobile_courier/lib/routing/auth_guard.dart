import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/auth/presentation/providers/auth_session_provider.dart';
import 'route_names.dart';

String? authGuard(BuildContext context, GoRouterState state, Ref ref) {
  final path = state.matchedLocation;
  final requiresAuth = authRequiredPrefixes.any(
    (p) => path == p || path.startsWith('$p/'),
  );

  final isAuthRoute = path == RouteNames.auth || path.startsWith('${RouteNames.auth}/');
  final isSplash = path == RouteNames.splash;

  if (isSplash) return null;

  final session = ref.read(authSessionProvider);

  if (requiresAuth) {
    if (!session.isAuthenticated) {
      final redirect = Uri.encodeComponent(state.uri.toString());
      return '${RouteNames.authPhone}?redirect=$redirect';
    }
    // Soft-gate: KYC incomplete → docs screen (allow KYC route itself).
    if (!session.kycStatus.isApproved && path != RouteNames.authKyc) {
      return RouteNames.authKyc;
    }
    return null;
  }

  if (isAuthRoute &&
      session.isAuthenticated &&
      session.kycStatus.isApproved &&
      path != RouteNames.authKyc) {
    return RouteNames.home;
  }

  return null;
}
