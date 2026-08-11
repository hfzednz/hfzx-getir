import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:nexora_design/nexora_design.dart';

/// Lightweight map chrome for address / tracking previews (DS §19).
class NxMapView extends StatelessWidget {
  const NxMapView({
    super.key,
    required this.initialTarget,
    this.height = 200,
    this.zoom = 15,
    this.markers = const {},
    this.myLocationEnabled = true,
    this.onTap,
    this.onCameraIdle,
    this.controllerCompleter,
  });

  final LatLng initialTarget;
  final double height;
  final double zoom;
  final Set<Marker> markers;
  final bool myLocationEnabled;
  final void Function(LatLng)? onTap;
  final VoidCallback? onCameraIdle;
  final void Function(GoogleMapController)? controllerCompleter;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(NxRadius.md),
      child: SizedBox(
        height: height,
        width: double.infinity,
        child: GoogleMap(
          initialCameraPosition: CameraPosition(
            target: initialTarget,
            zoom: zoom,
          ),
          markers: markers,
          myLocationEnabled: myLocationEnabled,
          myLocationButtonEnabled: false,
          zoomControlsEnabled: false,
          compassEnabled: false,
          mapToolbarEnabled: false,
          onTap: onTap,
          onCameraIdle: onCameraIdle,
          onMapCreated: controllerCompleter,
        ),
      ),
    );
  }
}
