import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';
import 'package:go_router/go_router.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/utils/idempotency.dart';
import '../../../../shared/validators/address_validator.dart';
import '../../../../shared/widgets/nx_map_view.dart';
import '../../domain/entities/addresses_entity.dart';
import '../providers/addresses_providers.dart';
import '../widgets/address_map_picker.dart';

class AddressEditScreen extends ConsumerStatefulWidget {
  const AddressEditScreen({super.key, this.addressId});

  final String? addressId;

  bool get isEditing => addressId != null && addressId!.isNotEmpty;

  @override
  ConsumerState<AddressEditScreen> createState() => _AddressEditScreenState();
}

class _AddressEditScreenState extends ConsumerState<AddressEditScreen> {
  final _customLabelController = TextEditingController();
  final _formattedController = TextEditingController();
  final _buildingController = TextEditingController();
  final _floorController = TextEditingController();
  final _doorController = TextEditingController();
  final _instructionsController = TextEditingController();
  final _recipientNameController = TextEditingController();
  final _recipientPhoneController = TextEditingController();

  AddressLabel _label = AddressLabel.home;
  LatLng? _pin;
  bool _isFavorite = false;
  bool _isDefault = false;
  bool _serviceable = true;
  bool _saving = false;
  bool _loaded = false;
  String? _formattedError;
  String? _customLabelError;

  @override
  void dispose() {
    _customLabelController.dispose();
    _formattedController.dispose();
    _buildingController.dispose();
    _floorController.dispose();
    _doorController.dispose();
    _instructionsController.dispose();
    _recipientNameController.dispose();
    _recipientPhoneController.dispose();
    super.dispose();
  }

  void _populateFromAddress(Address address) {
    if (_loaded) return;
    _loaded = true;
    _label = address.label;
    _customLabelController.text = address.customLabel;
    _formattedController.text = address.formatted;
    _buildingController.text = address.building;
    _floorController.text = address.floor;
    _doorController.text = address.door;
    _instructionsController.text = address.deliveryInstructions;
    _recipientNameController.text = address.recipientName;
    _recipientPhoneController.text = address.recipientPhone;
    _isFavorite = address.isFavorite;
    _isDefault = address.isDefault;
    _serviceable = address.serviceable;
    if (address.lat != null && address.lng != null) {
      _pin = LatLng(address.lat!, address.lng!);
    }
  }

  Map<String, dynamic> _buildBody() => {
        'label': addressLabelToJson(_label),
        if (_label == AddressLabel.custom)
          'custom_label': _customLabelController.text.trim(),
        'title': _label == AddressLabel.custom
            ? _customLabelController.text.trim()
            : addressLabelToJson(_label),
        'formatted': _formattedController.text.trim(),
        if (_pin != null) ...{'lat': _pin!.latitude, 'lng': _pin!.longitude},
        if (_buildingController.text.trim().isNotEmpty)
          'building': _buildingController.text.trim(),
        if (_floorController.text.trim().isNotEmpty)
          'floor': _floorController.text.trim(),
        if (_doorController.text.trim().isNotEmpty)
          'door': _doorController.text.trim(),
        if (_instructionsController.text.trim().isNotEmpty)
          'delivery_instructions': _instructionsController.text.trim(),
        if (_recipientNameController.text.trim().isNotEmpty)
          'recipient_name': _recipientNameController.text.trim(),
        if (_recipientPhoneController.text.trim().isNotEmpty)
          'phone': _recipientPhoneController.text.trim(),
        'is_favorite': _isFavorite,
        'is_default': _isDefault,
      };

  bool _validateForm() {
    final formattedError =
        AddressValidator.validateFormatted(_formattedController.text);
    final customLabelError = _label == AddressLabel.custom
        ? AddressValidator.validateTitle(_customLabelController.text)
        : null;

    setState(() {
      _formattedError = formattedError;
      _customLabelError = customLabelError;
    });

    return formattedError == null && customLabelError == null;
  }

  Future<void> _validateZone() async {
    if (_pin == null) return;
    final result = await ref.read(validateAddressZoneUseCaseProvider).call(
          lat: _pin!.latitude,
          lng: _pin!.longitude,
        );
    if (!mounted) return;
    result.fold(
      onSuccess: (zone) => setState(() => _serviceable = zone.serviceable),
      onFailure: (_) {},
    );
  }

  Future<void> _openMapPicker() async {
    LatLng initial = _pin ?? const LatLng(41.0082, 28.9784);
    if (_pin == null) {
      try {
        final permission = await Geolocator.checkPermission();
        if (permission == LocationPermission.denied) {
          await Geolocator.requestPermission();
        }
        final pos = await Geolocator.getCurrentPosition();
        initial = LatLng(pos.latitude, pos.longitude);
      } catch (_) {}
    }

    if (!mounted) return;
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (ctx) {
        return SizedBox(
          height: MediaQuery.sizeOf(ctx).height * 0.85,
          child: AddressMapPicker(
            initial: initial,
            onConfirm: (latLng) async {
              Navigator.of(ctx).pop();
              setState(() => _pin = latLng);
              await _validateZone();
            },
          ),
        );
      },
    );
  }

  Future<void> _save() async {
    if (!_validateForm()) return;

    final coords = AddressValidator.validateCoordinatesResult(
      lat: _pin?.latitude,
      lng: _pin?.longitude,
    );
    if (coords.isFailure) {
      NxToast.show(context, message: coords.errorOrNull!.message);
      return;
    }

    setState(() => _saving = true);
    final repo = ref.read(addressesRepositoryProvider);
    final body = _buildBody();
    final result = widget.isEditing
        ? await repo.updateAddress(
            id: widget.addressId!,
            body: body,
            idempotencyKey: Idempotency.generate(),
          )
        : await repo.createAddress(
            body: body,
            idempotencyKey: Idempotency.generate(),
          );

    if (!mounted) return;
    final l10n = AppLocalizations.of(context);
    setState(() => _saving = false);

    result.fold(
      onSuccess: (_) {
        ref.invalidate(addressesListProvider);
        if (widget.addressId != null) {
          ref.invalidate(addressDetailProvider(widget.addressId!));
        }
        context.pop();
        NxToast.show(
          context,
          message: widget.isEditing ? l10n.addressUpdated : l10n.addressSaved,
          variant: NxToastVariant.success,
        );
      },
      onFailure: (e) => NxToast.show(
        context,
        message: e.message,
        variant: NxToastVariant.danger,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final editing = widget.isEditing;
    final detailAsync = editing
        ? ref.watch(addressDetailProvider(widget.addressId!))
        : null;

    detailAsync?.whenData(_populateFromAddress);

    return Scaffold(
      appBar: NxTopBar(title: editing ? l10n.editAddress : l10n.addAddress),
      body: detailAsync?.isLoading == true && !_loaded
          ? const Center(child: NxSpinner())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  NxCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(l10n.addressLabel, style: NxTypography.titleSm),
                        const SizedBox(height: NxSpacing.s2),
                        SegmentedButton<AddressLabel>(
                          segments: [
                            ButtonSegment(
                              value: AddressLabel.home,
                              label: Text(l10n.addressHome),
                              icon: const Icon(Icons.home_outlined),
                            ),
                            ButtonSegment(
                              value: AddressLabel.work,
                              label: Text(l10n.addressWork),
                              icon: const Icon(Icons.work_outline),
                            ),
                            ButtonSegment(
                              value: AddressLabel.custom,
                              label: Text(l10n.addressCustom),
                              icon: const Icon(Icons.edit_outlined),
                            ),
                          ],
                          selected: {_label},
                          onSelectionChanged: (s) =>
                              setState(() => _label = s.first),
                        ),
                        if (_label == AddressLabel.custom) ...[
                          const SizedBox(height: NxSpacing.s3),
                          NxField(
                            label: l10n.customLabel,
                            controller: _customLabelController,
                            error: _customLabelError,
                          ),
                        ],
                      ],
                    ),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: l10n.streetAddress,
                    controller: _formattedController,
                    error: _formattedError,
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  Row(
                    children: [
                      Expanded(
                        child: NxField(
                          label: l10n.building,
                          controller: _buildingController,
                        ),
                      ),
                      const SizedBox(width: NxSpacing.s3),
                      Expanded(
                        child: NxField(
                          label: l10n.floor,
                          controller: _floorController,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: l10n.doorApt,
                    controller: _doorController,
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: l10n.deliveryInstructions,
                    controller: _instructionsController,
                    maxLines: 3,
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: l10n.recipientName,
                    controller: _recipientNameController,
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: l10n.recipientPhone,
                    controller: _recipientPhoneController,
                    keyboardType: TextInputType.phone,
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  SwitchListTile(
                    title: Text(l10n.favorite),
                    value: _isFavorite,
                    onChanged: (v) => setState(() => _isFavorite = v),
                  ),
                  SwitchListTile(
                    title: Text(l10n.defaultAddress),
                    value: _isDefault,
                    onChanged: (v) => setState(() => _isDefault = v),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxButton(
                    label: _pin == null ? l10n.pickOnMap : l10n.changeLocation,
                    variant: NxButtonVariant.secondary,
                    expand: true,
                    onPressed: _openMapPicker,
                  ),
                  if (_pin != null) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxMapView(
                      initialTarget: _pin!,
                      height: 160,
                      myLocationEnabled: false,
                      markers: {
                        Marker(
                          markerId: const MarkerId('pin'),
                          position: _pin!,
                        ),
                      },
                    ),
                    const SizedBox(height: NxSpacing.s2),
                    if (!_serviceable)
                      Text(
                        l10n.deliveryNotAvailableHere,
                        style: NxTypography.captionMd.copyWith(
                          color: context.nxColors.danger,
                        ),
                      ),
                  ],
                  const SizedBox(height: NxSpacing.s4),
                  NxButton(
                    label: editing ? l10n.saveChanges : l10n.saveAddress,
                    expand: true,
                    loading: _saving,
                    disabled: _saving,
                    onPressed: _save,
                  ),
                ],
              ),
            ),
    );
  }
}
