import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../domain/entities/courier_session.dart';
import '../providers/auth_controller.dart';
import '../providers/auth_session_provider.dart';

class KycScreen extends ConsumerStatefulWidget {
  const KycScreen({super.key});

  @override
  ConsumerState<KycScreen> createState() => _KycScreenState();
}

class _KycScreenState extends ConsumerState<KycScreen> {
  final _picker = ImagePicker();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(authControllerProvider.notifier).refreshKyc();
    });
  }

  Future<void> _upload(KycDocumentType type) async {
    final file = await _picker.pickImage(source: ImageSource.gallery);
    if (file == null) return;
    final kyc = await ref.read(authControllerProvider.notifier).uploadKycDocument(
          type: type,
          filePath: file.path,
        );
    if (!mounted) return;
    if (kyc != null && kyc.isApproved) {
      context.go(RouteNames.home);
    }
  }

  String _statusLabel(KycDocumentStatus status) => switch (status) {
        KycDocumentStatus.missing => 'Missing',
        KycDocumentStatus.pending => 'Pending review',
        KycDocumentStatus.approved => 'Approved',
        KycDocumentStatus.rejected => 'Rejected',
      };

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final session = ref.watch(authSessionProvider);
    final auth = ref.watch(authControllerProvider);
    final kyc = session.kycStatus;
    final colors = context.nxColors;

    final docs = [
      KycDocumentType.license,
      KycDocumentType.vehicle,
      KycDocumentType.insurance,
    ];

    return Scaffold(
      appBar: NxTopBar(title: l10n.kycTitle),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text(
            'Upload license, vehicle, and insurance documents to go online.',
            style: NxTypography.bodyMd.copyWith(color: colors.textSecondary),
          ),
          const SizedBox(height: NxSpacing.s4),
          for (final type in docs) ...[
            _KycDocTile(
              type: type,
              status: kyc.document(type)?.status ?? KycDocumentStatus.missing,
              statusLabel: _statusLabel(
                kyc.document(type)?.status ?? KycDocumentStatus.missing,
              ),
              rejectionReason: kyc.document(type)?.rejectionReason,
              loading: auth.isLoading,
              onUpload: () => _upload(type),
            ),
            const SizedBox(height: NxSpacing.s3),
          ],
          if (kyc.isApproved) ...[
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: l10n.continueLabel,
              expand: true,
              onPressed: () => context.go(RouteNames.home),
            ),
          ] else if (kyc.isPending) ...[
            const SizedBox(height: NxSpacing.s4),
            Text(
              'Documents are under review. You can continue once approved.',
              style: NxTypography.bodySm.copyWith(color: colors.textSecondary),
            ),
          ],
        ],
      ),
    );
  }
}

class _KycDocTile extends StatelessWidget {
  const _KycDocTile({
    required this.type,
    required this.status,
    required this.statusLabel,
    required this.onUpload,
    this.rejectionReason,
    this.loading = false,
  });

  final KycDocumentType type;
  final KycDocumentStatus status;
  final String statusLabel;
  final String? rejectionReason;
  final VoidCallback onUpload;
  final bool loading;

  String get _title => switch (type) {
        KycDocumentType.license => 'Driving license',
        KycDocumentType.vehicle => 'Vehicle registration',
        KycDocumentType.insurance => 'Insurance',
      };

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final needsUpload = status == KycDocumentStatus.missing ||
        status == KycDocumentStatus.rejected;

    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.bgSurface,
        border: Border.all(color: colors.borderSubtle),
        borderRadius: BorderRadius.circular(NxRadius.md),
      ),
      child: Padding(
        padding: const EdgeInsets.all(NxSpacing.s3),
        child: Row(
          children: [
            Icon(
              switch (type) {
                KycDocumentType.license => Icons.badge_outlined,
                KycDocumentType.vehicle => Icons.two_wheeler_outlined,
                KycDocumentType.insurance => Icons.health_and_safety_outlined,
              },
              color: colors.textSecondary,
            ),
            const SizedBox(width: NxSpacing.s3),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(_title, style: NxTypography.titleSm),
                  Text(
                    statusLabel,
                    style: NxTypography.captionMd.copyWith(
                      color: status == KycDocumentStatus.approved
                          ? colors.success
                          : colors.textSecondary,
                    ),
                  ),
                  if (rejectionReason != null && rejectionReason!.isNotEmpty)
                    Text(
                      rejectionReason!,
                      style: NxTypography.captionMd.copyWith(color: colors.danger),
                    ),
                ],
              ),
            ),
            if (needsUpload)
              NxButton(
                label: 'Upload',
                loading: loading,
                onPressed: onUpload,
              ),
          ],
        ),
      ),
    );
  }
}
