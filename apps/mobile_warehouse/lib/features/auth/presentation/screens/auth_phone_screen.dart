import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';

/// Stub phone login until full OTP auth is wired by core agent.
class AuthPhoneScreen extends ConsumerStatefulWidget {
  const AuthPhoneScreen({super.key});

  @override
  ConsumerState<AuthPhoneScreen> createState() => _AuthPhoneScreenState();
}

class _AuthPhoneScreenState extends ConsumerState<AuthPhoneScreen> {
  final _phone = TextEditingController();

  @override
  void dispose() {
    _phone.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Warehouse login'),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              controller: _phone,
              keyboardType: TextInputType.phone,
              decoration: const InputDecoration(labelText: 'Phone number'),
            ),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: 'Continue',
              expand: true,
              onPressed: () {
                final phone = _phone.text.trim();
                context.push(
                  '${RouteNames.authOtp}?phone=${Uri.encodeComponent(phone)}',
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
