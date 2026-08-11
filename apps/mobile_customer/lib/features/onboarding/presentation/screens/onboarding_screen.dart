import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:permission_handler/permission_handler.dart';

import '../../../auth/presentation/providers/auth_session_provider.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';

class OnboardingScreen extends ConsumerStatefulWidget {
  const OnboardingScreen({super.key});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  final _pageController = PageController();
  int _page = 0;

  static const _pages = [
    ('Deliver in minutes', 'Curated essentials from local dark stores.'),
    ('Track live', 'See your courier approach on the map.'),
    ('Save & reorder', 'Favorites, lists, and smart recommendations.'),
  ];

  Future<void> _finish() async {
    await Permission.locationWhenInUse.request();
    await ref.read(authSessionProvider.notifier).completeOnboarding();
    if (!mounted) return;
    context.go(RouteNames.auth);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;

    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            Align(
              alignment: Alignment.centerRight,
              child: TextButton(
                onPressed: _finish,
                child: Text(l10n.skip),
              ),
            ),
            Expanded(
              child: PageView.builder(
                controller: _pageController,
                itemCount: _pages.length,
                onPageChanged: (i) => setState(() => _page = i),
                itemBuilder: (context, index) {
                  final (title, subtitle) = _pages[index];
                  return Padding(
                    padding: const EdgeInsets.all(NxSpacing.s6),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          height: 180,
                          width: double.infinity,
                          decoration: BoxDecoration(
                            color: colors.bgSurfaceRaised,
                            borderRadius: BorderRadius.circular(NxRadius.lg),
                          ),
                          child: Icon(
                            Icons.local_shipping_outlined,
                            size: 72,
                            color: colors.bgBrand,
                          ),
                        ),
                        const SizedBox(height: NxSpacing.s6),
                        Text(title, style: NxTypography.headlineMd),
                        const SizedBox(height: NxSpacing.s3),
                        Text(
                          subtitle,
                          style: NxTypography.bodyLg.copyWith(
                            color: colors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: NxButton(
                label: _page == _pages.length - 1
                    ? l10n.continueLabel
                    : l10n.continueLabel,
                expand: true,
                onPressed: () {
                  if (_page < _pages.length - 1) {
                    _pageController.nextPage(
                      duration: NxDuration.medium,
                      curve: NxCurves.standard,
                    );
                  } else {
                    _finish();
                  }
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}
