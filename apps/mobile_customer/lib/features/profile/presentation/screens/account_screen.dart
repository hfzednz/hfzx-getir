import 'package:flutter/material.dart';
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
    final l10n = AppLocalizations.of(context);
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
              title: Text(l10n.signOut, style: TextStyle(color: context.nxColors.danger)),
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
