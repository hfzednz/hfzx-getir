import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../providers/auth_controller.dart';

class AuthOtpScreen extends ConsumerStatefulWidget {
  const AuthOtpScreen({super.key, this.phone});

  final String? phone;

  @override
  ConsumerState<AuthOtpScreen> createState() => _AuthOtpScreenState();
}

class _AuthOtpScreenState extends ConsumerState<AuthOtpScreen> {
  String _code = '';

  Future<void> _resend() async {
    if (widget.phone == null) return;
    final ok =
        await ref.read(authControllerProvider.notifier).resendOtp(widget.phone!);
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(ok ? 'Code resent' : 'Failed to resend code')),
    );
  }

  Future<void> _verify() async {
    if (_code.length < 6 || widget.phone == null) return;
    final session = await ref.read(authControllerProvider.notifier).verifyOtp(
          phone: widget.phone!,
          code: _code,
        );
    if (!mounted || session == null) return;
    if (session.kycStatus.isApproved) {
      context.go(RouteNames.home);
    } else {
      context.go(RouteNames.authKyc);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.otpHint),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (widget.phone != null)
              Text(widget.phone!, style: NxTypography.bodyMd),
            const SizedBox(height: NxSpacing.s4),
            NxOtpInput(
              length: 6,
              onChanged: (v) => setState(() => _code = v),
              onCompleted: (_) => _verify(),
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: l10n.verifyOtp,
              expand: true,
              loading: auth.isLoading,
              onPressed: _verify,
            ),
            TextButton(
              onPressed: _resend,
              child: const Text('Resend code'),
            ),
          ],
        ),
      ),
    );
  }
}
