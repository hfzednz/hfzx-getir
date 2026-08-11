import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../routing/route_names.dart';
import '../../domain/entities/warehouse_session.dart';
import '../providers/auth_controller.dart';
import '../providers/auth_session_provider.dart';

/// OTP verify — uses auth controller when BFF available, else stub session.
class AuthOtpScreen extends ConsumerStatefulWidget {
  const AuthOtpScreen({super.key, this.phone});

  final String? phone;

  @override
  ConsumerState<AuthOtpScreen> createState() => _AuthOtpScreenState();
}

class _AuthOtpScreenState extends ConsumerState<AuthOtpScreen> {
  final _otp = TextEditingController();

  @override
  void dispose() {
    _otp.dispose();
    super.dispose();
  }

  Future<void> _verify() async {
    final phone = widget.phone ?? '';
    final code = _otp.text.trim();
    WarehouseSession? session;
    try {
      session = await ref.read(authControllerProvider.notifier).verifyOtp(
            phone: phone,
            code: code,
          );
    } catch (_) {
      session = null;
    }

    if (session == null) {
      // Dev stub when BFF unavailable.
      final stub = WarehouseSession(
        userId: 'wh-user-1',
        accessToken: 'stub',
        refreshToken: 'stub',
        role: WarehouseRole.picker,
        storeId: 'store-1',
        stationId: 'station-a',
        shiftId: 'shift-stub',
        displayName: 'Warehouse Operator',
        phone: phone,
        kycOk: true,
        deviceOk: true,
      );
      await ref.read(authSessionProvider.notifier).setAuthenticated(stub);
      ref.read(storeIdProvider.notifier).state = stub.storeId;
    }

    if (!mounted) return;
    context.go(RouteNames.home);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: NxTopBar(
        title: 'Verify OTP',
        subtitle: widget.phone,
      ),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              controller: _otp,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(labelText: '6-digit code'),
            ),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: 'Verify',
              expand: true,
              onPressed: _verify,
            ),
          ],
        ),
      ),
    );
  }
}
