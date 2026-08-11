import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../domain/entities/checkout_entity.dart';
import '../providers/checkout_providers.dart';

class CheckoutScheduleScreen extends ConsumerStatefulWidget {
  const CheckoutScheduleScreen({super.key});

  @override
  ConsumerState<CheckoutScheduleScreen> createState() =>
      _CheckoutScheduleScreenState();
}

class _CheckoutScheduleScreenState
    extends ConsumerState<CheckoutScheduleScreen> {
  late final TextEditingController _noteController;
  late final TextEditingController _giftMessageController;
  late final TextEditingController _companyNameController;
  late final TextEditingController _taxIdController;
  late final TextEditingController _taxOfficeController;

  @override
  void initState() {
    super.initState();
    final checkout = ref.read(checkoutControllerProvider);
    _noteController = TextEditingController(text: checkout.courierNote ?? '');
    _giftMessageController =
        TextEditingController(text: checkout.giftMessage ?? '');
    _companyNameController =
        TextEditingController(text: checkout.invoiceFields?.companyName ?? '');
    _taxIdController =
        TextEditingController(text: checkout.invoiceFields?.taxId ?? '');
    _taxOfficeController =
        TextEditingController(text: checkout.invoiceFields?.taxOffice ?? '');
  }

  @override
  void dispose() {
    _noteController.dispose();
    _giftMessageController.dispose();
    _companyNameController.dispose();
    _taxIdController.dispose();
    _taxOfficeController.dispose();
    super.dispose();
  }

  void _syncInvoiceFields() {
    ref.read(checkoutControllerProvider.notifier).setInvoice(
          fields: CompanyInvoiceFields(
            companyName: _companyNameController.text.trim().isEmpty
                ? null
                : _companyNameController.text.trim(),
            taxId: _taxIdController.text.trim().isEmpty
                ? null
                : _taxIdController.text.trim(),
            taxOffice: _taxOfficeController.text.trim().isEmpty
                ? null
                : _taxOfficeController.text.trim(),
          ),
        );
  }

  Future<void> _pickSchedule() async {
    final now = DateTime.now();
    final date = await showDatePicker(
      context: context,
      initialDate: now.add(const Duration(hours: 1)),
      firstDate: now,
      lastDate: now.add(const Duration(days: 7)),
    );
    if (date == null || !mounted) return;

    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(now.add(const Duration(hours: 1))),
    );
    if (time == null || !mounted) return;

    final scheduled = DateTime(
      date.year,
      date.month,
      date.day,
      time.hour,
      time.minute,
    );
    ref.read(checkoutControllerProvider.notifier).setSchedule(scheduledAt: scheduled);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;
    final checkout = ref.watch(checkoutControllerProvider);
    final isAsap = checkout.scheduledAt == null;
    final scheduleLabel = checkout.scheduledAt == null
        ? 'As soon as possible'
        : DateFormat('EEE, d MMM · HH:mm').format(checkout.scheduledAt!);

    return Scaffold(
      appBar: NxTopBar(title: l10n.scheduleDelivery),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              children: [
                Text(
                  'When should we deliver?',
                  style: NxTypography.headlineSm.copyWith(color: colors.textPrimary),
                ),
                const SizedBox(height: NxSpacing.s4),
                NxCard(
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSchedule(asap: true),
                  child: Row(
                    children: [
                      Icon(
                        isAsap
                            ? Icons.radio_button_checked
                            : Icons.radio_button_off,
                        color: isAsap ? colors.bgBrand : colors.textTertiary,
                      ),
                      const SizedBox(width: NxSpacing.s3),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'ASAP',
                              style: NxTypography.titleSm
                                  .copyWith(color: colors.textPrimary),
                            ),
                            Text(
                              'We will deliver as soon as your order is ready.',
                              style: NxTypography.bodySm
                                  .copyWith(color: colors.textSecondary),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: NxSpacing.s3),
                NxCard(
                  onTap: _pickSchedule,
                  child: Row(
                    children: [
                      Icon(
                        !isAsap
                            ? Icons.radio_button_checked
                            : Icons.radio_button_off,
                        color: !isAsap ? colors.bgBrand : colors.textTertiary,
                      ),
                      const SizedBox(width: NxSpacing.s3),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              l10n.scheduleDelivery,
                              style: NxTypography.titleSm
                                  .copyWith(color: colors.textPrimary),
                            ),
                            Text(
                              scheduleLabel,
                              style: NxTypography.bodySm
                                  .copyWith(color: colors.textSecondary),
                            ),
                          ],
                        ),
                      ),
                      Icon(Icons.schedule, color: colors.textTertiary),
                    ],
                  ),
                ),
                const SizedBox(height: NxSpacing.s5),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    l10n.contactlessDelivery,
                    style: NxTypography.titleSm
                        .copyWith(color: colors.textPrimary),
                  ),
                  subtitle: Text(
                    'Leave at the door — no handoff needed.',
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
                  value: checkout.contactless,
                  activeThumbColor: colors.bgBrand,
                  onChanged: (value) => ref
                      .read(checkoutControllerProvider.notifier)
                      .setNotes(contactless: value),
                ),
                const SizedBox(height: NxSpacing.s3),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    l10n.giftOrder,
                    style: NxTypography.titleSm
                        .copyWith(color: colors.textPrimary),
                  ),
                  subtitle: Text(
                    'Hide prices on the receipt for the recipient.',
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
                  value: checkout.gift,
                  activeThumbColor: colors.bgBrand,
                  onChanged: (value) => ref
                      .read(checkoutControllerProvider.notifier)
                      .setNotes(gift: value),
                ),
                if (checkout.gift) ...[
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: 'Gift message',
                    controller: _giftMessageController,
                    maxLines: 3,
                    onChanged: (value) => ref
                        .read(checkoutControllerProvider.notifier)
                        .setGiftMessage(value),
                  ),
                ],
                const SizedBox(height: NxSpacing.s5),
                Text(
                  'If an item is unavailable',
                  style: NxTypography.titleMd.copyWith(color: colors.textPrimary),
                ),
                const SizedBox(height: NxSpacing.s2),
                Text(
                  'Substitution preference',
                  style: NxTypography.bodySm.copyWith(color: colors.textSecondary),
                ),
                const SizedBox(height: NxSpacing.s2),
                _OptionTile(
                  selected: checkout.substitutionPreference ==
                      SubstitutionPreference.allow,
                  title: 'Allow substitutions',
                  subtitle: 'Replace with a similar item automatically',
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSubstitution(preference: SubstitutionPreference.allow),
                ),
                const SizedBox(height: NxSpacing.s2),
                _OptionTile(
                  selected: checkout.substitutionPreference ==
                      SubstitutionPreference.contact,
                  title: 'Contact me',
                  subtitle: 'Call before replacing anything',
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSubstitution(
                        preference: SubstitutionPreference.contact,
                      ),
                ),
                const SizedBox(height: NxSpacing.s2),
                _OptionTile(
                  selected: checkout.substitutionPreference ==
                      SubstitutionPreference.reject,
                  title: 'Do not substitute',
                  subtitle: 'Skip or refund unavailable items',
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSubstitution(
                        preference: SubstitutionPreference.reject,
                      ),
                ),
                const SizedBox(height: NxSpacing.s4),
                Text(
                  'Out of stock rule',
                  style: NxTypography.bodySm.copyWith(color: colors.textSecondary),
                ),
                const SizedBox(height: NxSpacing.s2),
                _OptionTile(
                  selected:
                      checkout.outOfStockRule == OutOfStockReplacementRule.similar,
                  title: 'Replace with similar',
                  subtitle: 'Find a close match when possible',
                  enabled: checkout.substitutionPreference !=
                      SubstitutionPreference.reject,
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSubstitution(
                        outOfStockRule: OutOfStockReplacementRule.similar,
                      ),
                ),
                const SizedBox(height: NxSpacing.s2),
                _OptionTile(
                  selected:
                      checkout.outOfStockRule == OutOfStockReplacementRule.refund,
                  title: 'Refund item',
                  subtitle: 'Remove the item and refund its price',
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSubstitution(
                        outOfStockRule: OutOfStockReplacementRule.refund,
                      ),
                ),
                const SizedBox(height: NxSpacing.s2),
                _OptionTile(
                  selected:
                      checkout.outOfStockRule == OutOfStockReplacementRule.cancel,
                  title: 'Cancel order',
                  subtitle: 'Cancel the whole order if anything is missing',
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setSubstitution(
                        outOfStockRule: OutOfStockReplacementRule.cancel,
                      ),
                ),
                const SizedBox(height: NxSpacing.s5),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    'Company invoice',
                    style: NxTypography.titleSm
                        .copyWith(color: colors.textPrimary),
                  ),
                  subtitle: Text(
                    'Request a corporate invoice for this order.',
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
                  value: checkout.wantInvoice,
                  activeThumbColor: colors.bgBrand,
                  onChanged: (value) {
                    ref
                        .read(checkoutControllerProvider.notifier)
                        .setInvoice(wantInvoice: value);
                    if (value) _syncInvoiceFields();
                  },
                ),
                if (checkout.wantInvoice) ...[
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: 'Company name',
                    controller: _companyNameController,
                    onChanged: (_) => _syncInvoiceFields(),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: 'Tax ID',
                    controller: _taxIdController,
                    onChanged: (_) => _syncInvoiceFields(),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: 'Tax office',
                    controller: _taxOfficeController,
                    onChanged: (_) => _syncInvoiceFields(),
                  ),
                ],
                const SizedBox(height: NxSpacing.s4),
                NxField(
                  label: l10n.orderNotes,
                  controller: _noteController,
                  maxLines: 3,
                  onChanged: (value) => ref
                      .read(checkoutControllerProvider.notifier)
                      .setNotes(
                        courierNote: value,
                        clearCourierNote: value.trim().isEmpty,
                      ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: NxButton(
              label: l10n.continueLabel,
              expand: true,
              onPressed: () => context.push(RouteNames.checkoutPayment),
            ),
          ),
        ],
      ),
    );
  }
}

class _OptionTile extends StatelessWidget {
  const _OptionTile({
    required this.selected,
    required this.title,
    required this.onTap,
    this.subtitle,
    this.enabled = true,
  });

  final bool selected;
  final String title;
  final String? subtitle;
  final VoidCallback onTap;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    return Opacity(
      opacity: enabled ? 1 : 0.45,
      child: NxCard(
        onTap: enabled ? onTap : null,
        child: Row(
          children: [
            Icon(
              selected ? Icons.radio_button_checked : Icons.radio_button_off,
              color: selected ? colors.bgBrand : colors.textTertiary,
            ),
            const SizedBox(width: NxSpacing.s3),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: NxTypography.titleSm
                        .copyWith(color: colors.textPrimary),
                  ),
                  if (subtitle != null)
                    Text(
                      subtitle!,
                      style: NxTypography.bodySm
                          .copyWith(color: colors.textSecondary),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
