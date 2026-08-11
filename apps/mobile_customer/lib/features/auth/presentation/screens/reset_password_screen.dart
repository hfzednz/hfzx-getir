import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

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
      setState(() => _error = 'Reset token is required');
      return;
    }
    final passwordError = PasswordValidator.validate(password);
    if (passwordError != null) {
      setState(() => _error = passwordError);
      return;
    }
    if (password != confirm) {
      setState(() => _error = 'Passwords do not match');
      return;
    }

    final ok = await ref
        .read(accountLifecycleControllerProvider.notifier)
        .resetPassword(token: token, newPassword: password);
    if (!mounted) return;
    if (ok) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Password updated successfully')),
      );
      context.go(RouteNames.authEmail);
    } else {
      final err = ref.read(accountLifecycleControllerProvider).error;
      setState(() => _error = err ?? 'Failed to reset password');
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(accountLifecycleControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: 'Reset password'),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            NxField(
              controller: _tokenController,
              label: 'Reset token',
              error: _error,
            ),
            const SizedBox(height: NxSpacing.s3),
            NxField(
              controller: _passwordController,
              label: 'New password',
              obscureText: true,
            ),
            const SizedBox(height: NxSpacing.s3),
            NxField(
              controller: _confirmController,
              label: 'Confirm password',
              obscureText: true,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: 'Update password',
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
