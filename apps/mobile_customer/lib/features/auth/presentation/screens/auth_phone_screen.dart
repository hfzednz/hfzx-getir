import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../providers/auth_controller.dart';

class AuthPhoneScreen extends ConsumerStatefulWidget {
  const AuthPhoneScreen({super.key});

  @override
  ConsumerState<AuthPhoneScreen> createState() => _AuthPhoneScreenState();
}

class _AuthPhoneScreenState extends ConsumerState<AuthPhoneScreen> {
  final _controller = TextEditingController();
  String? _error;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final phone = _controller.text.trim();
    if (phone.isEmpty) {
      setState(() => _error = AppLocalizations.of(context).enterPhone);
      return;
    }
    final ok = await ref.read(authControllerProvider.notifier).requestOtp(phone);
    if (!mounted) return;
    if (ok) {
      unawaited(context.push('${RouteNames.authOtp}?phone=${Uri.encodeComponent(phone)}'));
    } else {
      setState(() => _error = ref.read(authControllerProvider).error ??
          AppLocalizations.of(context).otpSendFailed,);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.phoneLogin),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            NxField(
              controller: _controller,
              label: l10n.phoneLogin,
              keyboardType: TextInputType.phone,
              error: _error ?? auth.error,
            ),
            const SizedBox(height: NxSpacing.s4),
            SizedBox(
              height: 48,
              child: NxButton(
                label: l10n.sendOtp,
                expand: true,
                loading: auth.isLoading,
                onPressed: _submit,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
