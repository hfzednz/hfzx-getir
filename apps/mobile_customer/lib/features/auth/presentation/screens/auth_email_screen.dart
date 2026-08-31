import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/validators/email_validator.dart';
import '../../../../shared/validators/password_validator.dart';
import '../providers/auth_controller.dart';

class AuthEmailScreen extends ConsumerStatefulWidget {
  const AuthEmailScreen({super.key});

  @override
  ConsumerState<AuthEmailScreen> createState() => _AuthEmailScreenState();
}

class _AuthEmailScreenState extends ConsumerState<AuthEmailScreen> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _nameController = TextEditingController();
  bool _registerMode = false;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final email = _emailController.text.trim();
    final password = _passwordController.text;
    final emailError = EmailValidator.validate(email);
    if (emailError != null) return;
    final passwordError = PasswordValidator.validate(password);
    if (passwordError != null) return;

    final notifier = ref.read(authControllerProvider.notifier);
    if (_registerMode) {
      final ok = await notifier.registerEmail(
        email: email,
        password: password,
        name: _nameController.text.trim().isEmpty ? null : _nameController.text.trim(),
        context: context,
      );
      if (ok && mounted) {
        context.go('${RouteNames.authEmailVerify}?email=${Uri.encodeComponent(email)}');
      }
      return;
    }

    final ok = await notifier.signInEmail(
      email: email,
      password: password,
      context: context,
    );
    if (ok && mounted) context.go(RouteNames.home);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: _registerMode ? l10n.createAccount : l10n.emailLogin),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_registerMode) ...[
              NxField(
                controller: _nameController,
                label: l10n.nameLabel,
              ),
              const SizedBox(height: NxSpacing.s3),
            ],
            NxField(
              controller: _emailController,
              label: l10n.emailLogin,
              keyboardType: TextInputType.emailAddress,
            ),
            const SizedBox(height: NxSpacing.s3),
            NxField(
              controller: _passwordController,
              label: l10n.passwordLabel,
              obscureText: true,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: _registerMode ? l10n.register : l10n.signIn,
              expand: true,
              loading: auth.isLoading,
              onPressed: _submit,
            ),
            const SizedBox(height: NxSpacing.s3),
            if (!_registerMode)
              TextButton(
                onPressed: () => context.push(RouteNames.authForgotPassword),
                child: Text(l10n.forgotPassword),
              ),
            TextButton(
              onPressed: () => setState(() => _registerMode = !_registerMode),
              child: Text(
                _registerMode
                    ? l10n.alreadyHaveAccount
                    : l10n.needAnAccount,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
