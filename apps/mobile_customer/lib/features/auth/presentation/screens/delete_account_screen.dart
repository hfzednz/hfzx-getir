import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../providers/account_lifecycle_providers.dart';
import '../providers/auth_controller.dart';
import '../providers/auth_session_provider.dart';

class DeleteAccountScreen extends ConsumerStatefulWidget {
  const DeleteAccountScreen({super.key});

  @override
  ConsumerState<DeleteAccountScreen> createState() =>
      _DeleteAccountScreenState();
}

class _DeleteAccountScreenState extends ConsumerState<DeleteAccountScreen> {
  final _reasonController = TextEditingController();
  bool _confirmed = false;
  String? _error;

  @override
  void dispose() {
    _reasonController.dispose();
    super.dispose();
  }

  Future<void> _delete() async {
    if (!_confirmed) {
      setState(() => _error = AppLocalizations.of(context).confirmAccountDeletion);
      return;
    }
    final ok = await ref
        .read(accountLifecycleControllerProvider.notifier)
        .deleteAccount(reason: _reasonController.text.trim());
    if (!mounted) return;
    if (ok) {
      await ref.read(authControllerProvider.notifier).signOut();
      if (mounted) context.go(RouteNames.auth);
    } else {
      final err = ref.read(accountLifecycleControllerProvider).error;
      setState(() => _error = err ?? AppLocalizations.of(context).failedToDeleteAccount);
    }
  }

  Future<void> _exportData() async {
    final ok = await ref
        .read(accountLifecycleControllerProvider.notifier)
        .requestDataExport();
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          ok
              ? AppLocalizations.of(context).dataExportRequested
              : AppLocalizations.of(context).failedToRequestExport,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final lifecycle = ref.watch(accountLifecycleControllerProvider);
    final session = ref.watch(authSessionProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.deleteAccount),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.deleteAccountPermanent,
              style: NxTypography.bodyMd.copyWith(
                color: context.nxColors.danger,
              ),
            ),
            if (session.isAuthenticated) ...[
              const SizedBox(height: NxSpacing.s4),
              NxButton(
                label: l10n.requestDataExport,
                variant: NxButtonVariant.secondary,
                expand: true,
                loading: lifecycle.isLoading,
                onPressed: _exportData,
              ),
            ],
            const SizedBox(height: NxSpacing.s4),
            NxField(
              controller: _reasonController,
              label: l10n.reasonOptional,
              maxLines: 3,
              error: _error,
            ),
            const SizedBox(height: NxSpacing.s3),
            CheckboxListTile(
              value: _confirmed,
              onChanged: (v) => setState(() => _confirmed = v ?? false),
              title: Text(l10n.cannotBeUndone),
              controlAffinity: ListTileControlAffinity.leading,
            ),
            const Spacer(),
            NxButton(
              label: l10n.deleteAccount,
              expand: true,
              loading: lifecycle.isLoading,
              onPressed: _delete,
            ),
          ],
        ),
      ),
    );
  }
}
