import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/profile_entity.dart';
import '../providers/profile_providers.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(profileProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Profile'),
      body: AsyncValueWidget<CourierProfile>(
        value: async,
        data: (p) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s4),
          children: [
            _section('Personal', [
              _row('Name', p.displayName ?? '—', colors),
              _row('Phone', p.phone ?? '—', colors),
              _row('Email', p.email ?? '—', colors),
            ]),
            _section('Vehicle', [
              _row('Type', p.vehicleType ?? '—', colors),
              _row('Plate', p.vehiclePlate ?? '—', colors),
            ]),
            _section('Bank', [
              _row('Account', p.bankLast4 != null ? '•••• ${p.bankLast4}' : '—', colors),
            ]),
            _section('Tax', [
              _row('Tax ID', p.taxIdMasked ?? '—', colors),
            ]),
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: 'Documents / KYC',
              expand: true,
              onPressed: () => context.push(RouteNames.documents),
            ),
          ],
        ),
      ),
    );
  }

  Widget _section(String title, List<Widget> children) {
    return Padding(
      padding: const EdgeInsets.only(bottom: NxSpacing.s3),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: NxTypography.titleSm),
          const SizedBox(height: NxSpacing.s2),
          ...children,
        ],
      ),
    );
  }

  Widget _row(String label, String value, NxColorRoles colors) {
    return Padding(
      padding: const EdgeInsets.only(bottom: NxSpacing.s2),
      child: NxCard(
        child: Row(
          children: [
            Expanded(
              child: Text(
                label,
                style: NxTypography.bodySm.copyWith(color: colors.textSecondary),
              ),
            ),
            Text(
              value,
              style: NxTypography.bodyMd.copyWith(color: colors.textPrimary),
            ),
          ],
        ),
      ),
    );
  }
}
