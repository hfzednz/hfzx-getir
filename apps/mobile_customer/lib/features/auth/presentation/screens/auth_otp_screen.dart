import 'dart:async';

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
  int _cooldown = 0;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _startCooldown();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _startCooldown() {
    _timer?.cancel();
    setState(() => _cooldown = 30);
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (!mounted) {
        t.cancel();
        return;
      }
      if (_cooldown <= 1) {
        t.cancel();
        setState(() => _cooldown = 0);
      } else {
        setState(() => _cooldown -= 1);
      }
    });
  }

  Future<void> _resend() async {
    if (widget.phone == null || _cooldown > 0) return;
    final ok =
        await ref.read(authControllerProvider.notifier).resendOtp(widget.phone!);
    if (!mounted) return;
    if (ok) _startCooldown();
  }

  Future<void> _verify() async {
    if (_code.length < 6 || widget.phone == null) return;
    final ok = await ref.read(authControllerProvider.notifier).verifyOtp(
          phone: widget.phone!,
          code: _code,
          context: context,
        );
    if (ok && mounted) context.go(RouteNames.home);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.otpTitle),
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
            if (auth.error != null) ...[
              const SizedBox(height: NxSpacing.s3),
              Text(
                auth.error!,
                style: NxTypography.bodySm.copyWith(
                  color: context.nxColors.danger,
                ),
              ),
            ],
            const SizedBox(height: NxSpacing.s4),
            SizedBox(
              height: 48,
              child: NxButton(
                label: l10n.verifyOtp,
                expand: true,
                loading: auth.isLoading,
                onPressed: _verify,
              ),
            ),
            const SizedBox(height: NxSpacing.s2),
            SizedBox(
              height: 44,
              child: TextButton(
                onPressed: _cooldown > 0 || auth.isLoading ? null : _resend,
                child: Text(
                  _cooldown > 0
                      ? l10n.resendOtpIn(_cooldown)
                      : l10n.resendOtp,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
