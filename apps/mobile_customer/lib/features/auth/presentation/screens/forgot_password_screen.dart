import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/validators/email_validator.dart';
import '../providers/account_lifecycle_providers.dart';

class ForgotPasswordScreen extends ConsumerStatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  ConsumerState<ForgotPasswordScreen> createState() =>
      _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends ConsumerState<ForgotPasswordScreen> {
  final _emailController = TextEditingController();
  String? _error;

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final email = _emailController.text.trim();
    final validation = EmailValidator.validate(email);
    if (validation != null) {
      setState(() => _error = validation);
      return;
    }
    final ok = await ref
        .read(accountLifecycleControllerProvider.notifier)
        .forgotPassword(email);
    if (!mounted) return;
    if (ok) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Password reset email sent')),
      );
      context.pop();
    } else {
      final err = ref.read(accountLifecycleControllerProvider).error;
      setState(() => _error = err ?? 'Failed to send reset email');
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final state = ref.watch(accountLifecycleControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: 'Forgot password'),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Enter your email and we will send you a reset link.',
              style: NxTypography.bodyMd,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxField(
              controller: _emailController,
              label: l10n.emailLogin,
              keyboardType: TextInputType.emailAddress,
              error: _error,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: 'Send reset link',
              expand: true,
              loading: state.isLoading,
              onPressed: _submit,
            ),
            TextButton(
              onPressed: () => context.push(RouteNames.authResetPassword),
              child: const Text('Already have a reset token?'),
            ),
          ],
        ),
      ),
    );
  }
}
