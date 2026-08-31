import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/validators/password_validator.dart';
import '../providers/account_lifecycle_providers.dart';

class ResetPasswordScreen extends ConsumerStatefulWidget {
  const ResetPasswordScreen({super.key, this.token});

  final String? token;

  @override
  ConsumerState<ResetPasswordScreen> createState() =>
      _ResetPasswordScreenState();
}

class _ResetPasswordScreenState extends ConsumerState<ResetPasswordScreen> {
  final _tokenController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmController = TextEditingController();
  String? _error;

  @override
  void initState() {
    super.initState();
    if (widget.token != null) {
      _tokenController.text = widget.token!;
    }
  }

  @override
  void dispose() {
    _tokenController.dispose();
    _passwordController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final token = _tokenController.text.trim();
    final password = _passwordController.text;
    final confirm = _confirmController.text;

    if (token.isEmpty) {
      setState(() => _error = AppLocalizations.of(context).resetTokenRequired);
      return;
    }
    final passwordError = PasswordValidator.validate(password);
    if (passwordError != null) {
      setState(() => _error = passwordError);
      return;
    }
    if (password != confirm) {
      setState(() => _error = AppLocalizations.of(context).passwordsDoNotMatch);
      return;
    }

    final ok = await ref
        .read(accountLifecycleControllerProvider.notifier)
        .resetPassword(token: token, newPassword: password);
    if (!mounted) return;
    if (ok) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context).passwordUpdated)),
      );
      context.go(RouteNames.authEmail);
    } else {
      final err = ref.read(accountLifecycleControllerProvider).error;
      setState(() => _error = err ?? AppLocalizations.of(context).failedToResetPassword);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final state = ref.watch(accountLifecycleControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.resetPassword),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            NxField(
              controller: _tokenController,
              label: l10n.resetToken,
              error: _error,
            ),
            const SizedBox(height: NxSpacing.s3),
            NxField(
              controller: _passwordController,
              label: l10n.newPassword,
              obscureText: true,
            ),
            const SizedBox(height: NxSpacing.s3),
            NxField(
              controller: _confirmController,
              label: l10n.confirmPassword,
              obscureText: true,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: l10n.updatePassword,
              expand: true,
              loading: state.isLoading,
              onPressed: _submit,
            ),
          ],
        ),
      ),
    );
  }
}
