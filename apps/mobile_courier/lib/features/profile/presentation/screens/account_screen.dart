import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../auth/presentation/providers/auth_controller.dart';
import '../../../auth/presentation/providers/auth_session_provider.dart';

class AccountScreen extends ConsumerWidget {
  const AccountScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(authSessionProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Account'),
      body: ListView(
        children: [
          ListTile(
            title: Text(session.displayName ?? session.phone ?? 'Courier'),
            subtitle: Text(session.courierId ?? ''),
          ),
          ListTile(
            title: const Text('Profile'),
            onTap: () => context.push(RouteNames.profile),
          ),
          ListTile(
            title: const Text('Documents'),
            onTap: () => context.push(RouteNames.documents),
          ),
          ListTile(
            title: const Text('Start shift'),
            onTap: () => context.push(RouteNames.shifts),
          ),
          ListTile(
            title: const Text('Performance'),
            onTap: () => context.push(RouteNames.performance),
          ),
          ListTile(
            title: const Text('Support'),
            onTap: () => context.push(RouteNames.support),
          ),
          ListTile(
            title: const Text('Settings'),
            onTap: () => context.push(RouteNames.settings),
          ),
          ListTile(
            title: const Text('Notifications'),
            onTap: () => context.push(RouteNames.notifications),
          ),
          const Divider(),
          ListTile(
            title: const Text('Sign out'),
            onTap: () async {
              await ref.read(authControllerProvider.notifier).signOut();
              if (context.mounted) context.go(RouteNames.authPhone);
            },
          ),
        ],
      ),
    );
  }
}
