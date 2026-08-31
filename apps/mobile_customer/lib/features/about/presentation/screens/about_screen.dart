import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:package_info_plus/package_info_plus.dart';

import '../../../../l10n/app_localizations.dart';

class AboutScreen extends ConsumerStatefulWidget {
  const AboutScreen({super.key, this.id});

  final String? id;

  @override
  ConsumerState<AboutScreen> createState() => _AboutScreenState();
}

class _AboutScreenState extends ConsumerState<AboutScreen> {
  PackageInfo? _info;

  @override
  void initState() {
    super.initState();
    PackageInfo.fromPlatform().then((info) {
      if (mounted) setState(() => _info = info);
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final info = _info;
    final versionLabel = info == null
        ? '…'
        : '${info.version} (${info.buildNumber})';
    final appName = info?.appName.isNotEmpty == true ? info!.appName : 'NEXORA';

    return Scaffold(
      appBar: NxTopBar(title: l10n.aboutTitle),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text('NEXORA', style: NxTypography.headlineLg),
          const SizedBox(height: NxSpacing.s2),
          Text(
            'Quick commerce for everyday essentials — '
            'delivered fast with a focus on reliability and care.',
            style: NxTypography.bodyMd,
          ),
          const SizedBox(height: NxSpacing.s6),
          ListTile(
            contentPadding: EdgeInsets.zero,
            title: Text(appName),
            subtitle: Text('Version $versionLabel'),
          ),
          const Divider(),
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: const Icon(Icons.description_outlined),
            title: Text(l10n.termsOfService),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/legal/terms'),
          ),
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: const Icon(Icons.privacy_tip_outlined),
            title: Text(l10n.privacyPolicy),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/legal/privacy'),
          ),
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: const Icon(Icons.code_outlined),
            title: Text(l10n.openSourceLicenses),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => showLicensePage(
              context: context,
              applicationName: 'NEXORA',
              applicationVersion: versionLabel,
            ),
          ),
        ],
      ),
    );
  }
}
