import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../di/providers.dart';
import '../l10n/app_localizations.dart';
import '../routing/app_router.dart';
import '../shared/widgets/offline_banner.dart';

class NexoraCourierApp extends ConsumerWidget {
  const NexoraCourierApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    final themeMode = ref.watch(themeModeProvider);
    final localeCode = ref.watch(localeCodeProvider);

    return MaterialApp.router(
      title: 'NEXORA Courier',
      debugShowCheckedModeBanner: false,
      theme: NxTheme.light(density: NxDensity.compact),
      darkTheme: NxTheme.dark(density: NxDensity.compact),
      themeMode: themeMode,
      locale: Locale(localeCode),
      supportedLocales: const [Locale('en'), Locale('tr')],
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      routerConfig: router,
      builder: (context, child) {
        return MediaQuery.withClampedTextScaling(
          minScaleFactor: 0.85,
          maxScaleFactor: 1.3,
          child: Column(
            children: [
              const OfflineBanner(),
              Expanded(child: child ?? const SizedBox.shrink()),
            ],
          ),
        );
      },
    );
  }
}
