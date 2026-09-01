import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../providers/account_lifecycle_providers.dart';

class DevicesScreen extends ConsumerWidget {
  const DevicesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final devicesAsync = ref.watch(devicesListProvider);
    final lifecycle = ref.watch(accountLifecycleControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.security),
      body: devicesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(l10n.emptyMessage),
              const SizedBox(height: NxSpacing.s3),
              NxButton(
                label: l10n.retry,
                onPressed: () => ref.invalidate(devicesListProvider),
              ),
            ],
          ),
        ),
        data: (devices) {
          if (devices.isEmpty) {
            return Center(child: Text(l10n.emptyMessage));
          }
          return ListView.separated(
            itemCount: devices.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final device = devices[index];
              return ListTile(
                leading: Icon(
                  device.platform?.toLowerCase().contains('ios') == true
                      ? Icons.phone_iphone
                      : Icons.devices,
                ),
                title: Text(device.label),
                subtitle: Text(
                  [
                    if (device.platform != null) device.platform!,
                    if (device.lastActiveAt != null)
                      'Last active: ${device.lastActiveAt!.toLocal()}',
                  ].join(' · '),
                ),
                trailing: device.isCurrent
                    ? const Chip(label: Text('This device'))
                    : lifecycle.isLoading
                        ? const SizedBox(
                            width: 24,
                            height: 24,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : IconButton(
                            icon: Icon(
                              Icons.logout,
                              color: context.nxColors.danger,
                            ),
                            onPressed: () async {
                              final ok = await ref
                                  .read(accountLifecycleControllerProvider
                                      .notifier,)
                                  .revokeDevice(device.id);
                              if (ok) ref.invalidate(devicesListProvider);
                            },
                          ),
              );
            },
          );
        },
      ),
    );
  }
}
