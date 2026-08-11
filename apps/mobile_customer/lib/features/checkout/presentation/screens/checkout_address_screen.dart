import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../../addresses/domain/entities/addresses_entity.dart';
import '../../../addresses/presentation/providers/addresses_providers.dart';
import '../providers/checkout_providers.dart';

class CheckoutAddressScreen extends ConsumerWidget {
  const CheckoutAddressScreen({super.key});

  String _subtitle(Address address) {
    if (address.formatted.isNotEmpty) return address.formatted;
    return address.id;
  }

  bool _isDefault(Address address) => address.isDefault;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;
    final checkout = ref.watch(checkoutControllerProvider);
    final addressesAsync = ref.watch(addressesListProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.selectAddress),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: AsyncValueWidget(
              value: addressesAsync,
              data: (addresses) {
                if (addresses.isEmpty) {
                  return NxEmptyState(
                    title: l10n.emptyTitle,
                    body: 'Add a delivery address to continue checkout.',
                    primaryActionLabel: l10n.addAddress,
                    onPrimaryAction: () => context.push(RouteNames.addressAdd),
                  );
                }
                return ListView.separated(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  itemCount: addresses.length,
                  separatorBuilder: (context, index) =>
                      const SizedBox(height: NxSpacing.s3),
                  itemBuilder: (context, index) {
                    final address = addresses[index];
                    final selected = checkout.addressId == address.id;
                    return NxCard(
                      onTap: () => ref
                          .read(checkoutControllerProvider.notifier)
                          .setAddress(address.id),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(
                            selected
                                ? Icons.radio_button_checked
                                : Icons.radio_button_off,
                            color: selected
                                ? colors.bgBrand
                                : colors.textTertiary,
                          ),
                          const SizedBox(width: NxSpacing.s3),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Expanded(
                                      child: Text(
                                        address.title.isNotEmpty
                                            ? address.title
                                            : l10n.selectAddress,
                                        style: NxTypography.titleSm.copyWith(
                                          color: colors.textPrimary,
                                        ),
                                      ),
                                    ),
                                    if (_isDefault(address))
                                      Text(
                                        'Default',
                                        style: NxTypography.captionMd.copyWith(
                                          color: colors.textBrand,
                                        ),
                                      ),
                                  ],
                                ),
                                const SizedBox(height: NxSpacing.s1),
                                Text(
                                  _subtitle(address),
                                  style: NxTypography.bodySm.copyWith(
                                    color: colors.textSecondary,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                );
              },
              error: (e, _) => ErrorView(
                message: e.toString(),
                onRetry: () => ref.invalidate(addressesListProvider),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                NxButton(
                  label: l10n.addAddress,
                  variant: NxButtonVariant.secondary,
                  expand: true,
                  onPressed: () => context.push(RouteNames.addressAdd),
                ),
                const SizedBox(height: NxSpacing.s3),
                NxButton(
                  label: l10n.continueLabel,
                  expand: true,
                  disabled: checkout.addressId == null,
                  onPressed: checkout.addressId == null
                      ? null
                      : () => context.push(RouteNames.checkoutSchedule),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
