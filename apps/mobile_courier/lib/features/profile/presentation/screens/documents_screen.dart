import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../auth/domain/entities/courier_session.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/profile_entity.dart';
import '../providers/profile_providers.dart';

class DocumentsScreen extends ConsumerWidget {
  const DocumentsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(profileProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Documents'),
      body: AsyncValueWidget<CourierProfile>(
        value: async,
        data: (p) {
          final docs = p.kycStatus.documents.isEmpty
              ? KycDocumentType.values
                  .map(
                    (t) => KycDocument(
                      type: t,
                      status: KycDocumentStatus.missing,
                    ),
                  )
                  .toList()
              : p.kycStatus.documents;

          return ListView.separated(
            padding: const EdgeInsets.all(NxSpacing.s4),
            itemCount: docs.length,
            separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
            itemBuilder: (context, i) {
              final d = docs[i];
              return NxCard(
                child: Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            d.type.name,
                            style: NxTypography.titleSm
                                .copyWith(color: colors.textPrimary),
                          ),
                          Text(
                            d.status.name,
                            style: NxTypography.captionMd.copyWith(
                              color: d.isApproved
                                  ? colors.success
                                  : colors.textSecondary,
                            ),
                          ),
                          if (d.rejectionReason != null)
                            Text(
                              d.rejectionReason!,
                              style: NxTypography.captionSm
                                  .copyWith(color: colors.danger),
                            ),
                        ],
                      ),
                    ),
                    Icon(
                      d.isApproved ? Icons.verified : Icons.upload_file,
                      color: d.isApproved ? colors.success : colors.iconSecondary,
                    ),
                  ],
                ),
              );
            },
          );
        },
      ),
    );
  }
}
