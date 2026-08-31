import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';

/// Full-screen-ish sheet to pick a delivery pin on the map.
class AddressMapPicker extends StatefulWidget {
  const AddressMapPicker({
    super.key,
    required this.initial,
    required this.onConfirm,
  });

  final LatLng initial;
  final ValueChanged<LatLng> onConfirm;

  @override
  State<AddressMapPicker> createState() => _AddressMapPickerState();
}

class _AddressMapPickerState extends State<AddressMapPicker> {
  late LatLng _center;
  GoogleMapController? _controller;

  @override
  void initState() {
    super.initState();
    _center = widget.initial;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  l10n.pinYourAddress,
                  style: NxTypography.headlineSm.copyWith(
                    color: colors.textPrimary,
                  ),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ],
          ),
        ),
        Expanded(
          child: Stack(
            alignment: Alignment.center,
            children: [
              GoogleMap(
                initialCameraPosition: CameraPosition(
                  target: widget.initial,
                  zoom: 16,
                ),
                myLocationEnabled: true,
                myLocationButtonEnabled: true,
                zoomControlsEnabled: false,
                mapToolbarEnabled: false,
                onMapCreated: (c) => _controller = c,
                onCameraIdle: () async {
                  final c = _controller;
                  if (c == null) return;
                  final bounds = await c.getVisibleRegion();
                  final mid = LatLng(
                    (bounds.northeast.latitude + bounds.southwest.latitude) / 2,
                    (bounds.northeast.longitude + bounds.southwest.longitude) /
                        2,
                  );
                  setState(() => _center = mid);
                },
              ),
              IgnorePointer(
                child: Icon(
                  Icons.location_on,
                  size: 42,
                  color: colors.bgBrand,
                ),
              ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: NxButton(
            label: l10n.useThisLocation,
            expand: true,
            onPressed: () => widget.onConfirm(_center),
          ),
        ),
      ],
    );
  }
}
