import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../deliveries/domain/entities/delivery_job.dart';
import '../../../deliveries/presentation/providers/deliveries_providers.dart';
import '../../domain/entities/route_entity.dart';
import '../providers/navigation_providers.dart';
import '../../../../shared/widgets/async_value_widget.dart';

class DeliveryNavigateScreen extends ConsumerWidget {
  const DeliveryNavigateScreen({super.key, required this.deliveryId});

  final String deliveryId;

  Future<void> _openExternalNav(DeliveryJob job) async {
    final dest = job.status.index <= DeliveryLifecycleStatus.atStore.index
        ? job.storeLocation
        : job.customerLocation;
    final google = Uri.parse(
      'https://www.google.com/maps/dir/?api=1&destination=${dest.lat},${dest.lng}&travelmode=driving',
    );
    final apple = Uri.parse(
      'https://maps.apple.com/?daddr=${dest.lat},${dest.lng}&dirflg=d',
    );
    final uri = Platform.isIOS ? apple : google;
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final jobAsync = ref.watch(deliveryDetailProvider(deliveryId));
    final routeAsync = ref.watch(deliveryRouteProvider(deliveryId));

    return Scaffold(
      appBar: const NxTopBar(title: 'Navigate'),
      body: AsyncValueWidget<DeliveryJob>(
        value: jobAsync,
        data: (job) {
          final route = routeAsync.valueOrNull;
          final markers = <Marker>{
            Marker(
              markerId: const MarkerId('store'),
              position: LatLng(job.storeLocation.lat, job.storeLocation.lng),
              infoWindow: InfoWindow(title: job.storeName),
            ),
            Marker(
              markerId: const MarkerId('customer'),
              position:
                  LatLng(job.customerLocation.lat, job.customerLocation.lng),
              infoWindow: InfoWindow(title: job.customerArea),
            ),
          };
          final polyPoints = (route?.points ?? const <RoutePoint>[])
              .map((p) => LatLng(p.lat, p.lng))
              .toList();
          final polylines = polyPoints.length >= 2
              ? {
                  Polyline(
                    polylineId: const PolylineId('route'),
                    points: polyPoints,
                    color: context.nxColors.bgBrand,
                    width: 4,
                  ),
                }
              : <Polyline>{};

          return Column(
            children: [
              Expanded(
                child: GoogleMap(
                  initialCameraPosition: CameraPosition(
                    target: LatLng(job.storeLocation.lat, job.storeLocation.lng),
                    zoom: 13,
                  ),
                  markers: markers,
                  polylines: polylines,
                  myLocationEnabled: true,
                  myLocationButtonEnabled: true,
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(NxSpacing.s4),
                child: Column(
                  children: [
                    NxEtaCard(
                      etaRange: route?.etaMinutes != null
                          ? '${route!.etaMinutes} min'
                          : '—',
                      storeName: job.storeName,
                      confidenceCopy: route?.distanceMeters != null
                          ? '${(route!.distanceMeters! / 1000).toStringAsFixed(1)} km'
                          : null,
                      live: true,
                    ),
                    const SizedBox(height: NxSpacing.s3),
                    NxButton(
                      label: 'Open maps',
                      expand: true,
                      onPressed: () => _openExternalNav(job),
                    ),
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}
