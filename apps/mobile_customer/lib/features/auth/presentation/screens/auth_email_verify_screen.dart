import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../providers/account_lifecycle_providers.dart';

class AuthEmailVerifyScreen extends ConsumerStatefulWidget {
  const AuthEmailVerifyScreen({super.key, this.email});

  final String? email;

  @override
  ConsumerState<AuthEmailVerifyScreen> createState() =>
      _AuthEmailVerifyScreenState();
}

class _AuthEmailVerifyScreenState extends ConsumerState<AuthEmailVerifyScreen> {
  final _codeController = TextEditingController();
  String? _error;
  bool _loading = false;

  @override
  void dispose() {
    _codeController.dispose();
    super.dispose();
  }

  Future<void> _verify() async {
    final code = _codeController.text.trim();
    if (code.isEmpty) {
      setState(() => _error = 'Enter the verification code');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    final result = await ref.read(verifyEmailUseCaseProvider).call(code);
    if (!mounted) return;
    setState(() => _loading = false);
    await result.fold(
      onSuccess: (_) async {
        NxToast.show(
          context,
          message: 'Email verified',
          variant: NxToastVariant.success,
        );
        if (mounted) context.go(RouteNames.home);
      },
      onFailure: (e) async {
        setState(() => _error = e.message);
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Verify email'),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              widget.email != null
                  ? 'Enter the code we sent to ${widget.email}.'
                  : 'Enter the verification code from your email.',
              style: NxTypography.bodyMd,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxField(
              controller: _codeController,
              label: 'Verification code',
              keyboardType: TextInputType.number,
              error: _error,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: 'Verify email',
              expand: true,
              loading: _loading,
              onPressed: _verify,
            ),
            TextButton(
              onPressed: () => context.go(RouteNames.home),
              child: const Text('Skip for now'),
            ),
          ],
        ),
      ),
    );
  }
}
