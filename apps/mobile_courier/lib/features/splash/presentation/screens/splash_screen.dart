import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../auth/presentation/providers/auth_session_provider.dart';
import '../../../../bootstrap/bootstrap.dart';
import '../../../../routing/route_names.dart';

class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _bootstrap());
  }

  Future<void> _bootstrap() async {
    await postBootstrap(ref);
    if (!mounted) return;

    final session = ref.read(authSessionProvider);

    if (session.isAuthenticated) {
      if (session.kycStatus.isApproved) {
        context.go(RouteNames.home);
      } else {
        context.go(RouteNames.authKyc);
      }
      return;
    }

    context.go(RouteNames.authPhone);
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    return Scaffold(
      backgroundColor: colors.bgBrand,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              'NEXORA',
              style: NxTypography.displayLg.copyWith(color: colors.textOnBrand),
            ),
            const SizedBox(height: NxSpacing.s2),
            Text(
              'Courier',
              style: NxTypography.bodyMd.copyWith(color: colors.textOnBrand),
            ),
            const SizedBox(height: NxSpacing.s6),
            NxSpinner(color: colors.textOnBrand),
          ],
        ),
      ),
    );
  }
}
