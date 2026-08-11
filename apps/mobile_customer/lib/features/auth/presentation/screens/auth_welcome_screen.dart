import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../providers/auth_controller.dart';

class AuthWelcomeScreen extends ConsumerWidget {
  const AuthWelcomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.authTitle),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Spacer(),
            NxButton(
              label: l10n.phoneLogin,
              expand: true,
              onPressed: () => context.push(RouteNames.authPhone),
            ),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: l10n.emailLogin,
              variant: NxButtonVariant.secondary,
              expand: true,
              onPressed: () => context.push(RouteNames.authEmail),
            ),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: l10n.googleLogin,
              variant: NxButtonVariant.secondary,
              expand: true,
              loading: auth.isLoading,
              onPressed: () => ref.read(authControllerProvider.notifier).signInGoogle(context),
            ),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: l10n.appleLogin,
              variant: NxButtonVariant.secondary,
              expand: true,
              loading: auth.isLoading,
              onPressed: () => ref.read(authControllerProvider.notifier).signInApple(context),
            ),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: l10n.guestContinue,
              variant: NxButtonVariant.tertiary,
              expand: true,
              onPressed: () async {
                await ref.read(authControllerProvider.notifier).continueAsGuest(context);
              },
            ),
            const Spacer(),
          ],
        ),
      ),
    );
  }
}
