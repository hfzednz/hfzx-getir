import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../../../shared/widgets/offline_banner.dart';
import '../../domain/entities/tracking_entity.dart';
import '../providers/tracking_providers.dart';

class TrackingScreen extends ConsumerWidget {
  const TrackingScreen({super.key, required this.orderId});

  final String orderId;

  String _etaRange(TrackingSnapshot snapshot) {
    final min = snapshot.etaMin ?? snapshot.etaMinutes;
    final max = snapshot.etaMax ?? snapshot.etaMinutes;
    if (min != null && max != null && min != max) {
      return '$min–$max min';
    }
    if (min != null) return '$min min';
    if (max != null) return '$max min';
    return 'Calculating ETA';
  }

  NxTrackingStepState _stepState(String state) {
    final normalized = state.toLowerCase();
    return NxTrackingStepState.values.firstWhere(
      (s) => s.name == normalized,
      orElse: () {
        if (normalized == 'done' || normalized == 'complete') {
          return NxTrackingStepState.completed;
        }
        if (normalized == 'active' || normalized == 'in_progress') {
          return NxTrackingStepState.current;
        }
        if (normalized == 'error') return NxTrackingStepState.failed;
        return NxTrackingStepState.upcoming;
      },
    );
  }

  Set<Marker> _markers(TrackingSnapshot snapshot) {
    final markers = <Marker>{};
    if (snapshot.storeLat != null && snapshot.storeLng != null) {
      markers.add(
        Marker(
          markerId: const MarkerId('store'),
          position: LatLng(snapshot.storeLat!, snapshot.storeLng!),
          infoWindow: const InfoWindow(title: 'Store'),
          icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueAzure),
        ),
      );
    }
    if (snapshot.destLat != null && snapshot.destLng != null) {
      markers.add(
        Marker(
          markerId: const MarkerId('destination'),
          position: LatLng(snapshot.destLat!, snapshot.destLng!),
          infoWindow: const InfoWindow(title: 'Delivery'),
          icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueRed),
        ),
      );
    }
    if (snapshot.courierLat != null && snapshot.courierLng != null) {
      markers.add(
        Marker(
          markerId: const MarkerId('courier'),
          position: LatLng(snapshot.courierLat!, snapshot.courierLng!),
          infoWindow: InfoWindow(
            title: snapshot.courierName ?? 'Courier',
          ),
          icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueGreen),
        ),
      );
    }
    return markers;
  }

  Set<Polyline> _polylines(TrackingSnapshot snapshot, Color color) {
    if (snapshot.routePoints.length < 2) return const {};
    return {
      Polyline(
        polylineId: const PolylineId('route'),
        points: snapshot.routePoints
            .map((p) => LatLng(p.lat, p.lng))
            .toList(growable: false),
        color: color,
        width: 4,
      ),
    };
  }

  Future<void> _openCourierChat(
    BuildContext context,
    TrackingSnapshot snapshot,
  ) async {
    final chatUrl = snapshot.courierChatUrl?.trim();
    if (chatUrl != null && chatUrl.isNotEmpty) {
      final uri = Uri.tryParse(chatUrl);
      if (uri != null && await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
        return;
      }
    }
    if (!context.mounted) return;
    // ignore: unawaited_futures
    context.push('${RouteNames.supportAssistant}?orderId=$orderId');
  }

  LatLng _initialTarget(TrackingSnapshot snapshot, Set<Marker> markers) {
    if (snapshot.courierLat != null && snapshot.courierLng != null) {
      return LatLng(snapshot.courierLat!, snapshot.courierLng!);
    }
    if (snapshot.destLat != null && snapshot.destLng != null) {
      return LatLng(snapshot.destLat!, snapshot.destLng!);
    }
    if (snapshot.storeLat != null && snapshot.storeLng != null) {
      return LatLng(snapshot.storeLat!, snapshot.storeLng!);
    }
    if (snapshot.routePoints.isNotEmpty) {
      final p = snapshot.routePoints.first;
      return LatLng(p.lat, p.lng);
    }
    if (markers.isNotEmpty) return markers.first.position;
    return const LatLng(41.0082, 28.9784);
  }

  Future<void> _callCourier(String? phone) async {
    if (phone == null || phone.isEmpty) return;
    final uri = Uri(scheme: 'tel', path: phone);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri);
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;
    final trackingAsync = ref.watch(trackingRealtimeProvider(orderId));

    return Scaffold(
      appBar: NxTopBar(title: l10n.trackingTitle),
      body: Column(
        children: [
          const OfflineBanner(),
          Expanded(
            child: trackingAsync.when(
              data: (snapshot) {
                final markers = _markers(snapshot);
                final polylines = _polylines(snapshot, colors.textBrand);
                final hasMap = markers.isNotEmpty || polylines.isNotEmpty;
                final steps = snapshot.steps
                    .map(
                      (s) => NxTrackingStep(
                        label: s.title,
                        subtitle: s.subtitle,
                        state: _stepState(s.state),
                      ),
                    )
                    .toList();

                return Column(
                  children: [
                    if (hasMap)
                      SizedBox(
                        height: 220,
                        child: GoogleMap(
                          initialCameraPosition: CameraPosition(
                            target: _initialTarget(snapshot, markers),
                            zoom: 13.5,
                          ),
                          markers: markers,
                          polylines: polylines,
                          myLocationButtonEnabled: false,
                          zoomControlsEnabled: false,
                          compassEnabled: false,
                          mapToolbarEnabled: false,
                        ),
                      ),
                    Padding(
                      padding: const EdgeInsets.all(NxSpacing.s4),
                      child: NxEtaCard(
                        etaRange: _etaRange(snapshot),
                        confidenceCopy: snapshot.courierName != null
                            ? 'Courier: ${snapshot.courierName}'
                            : snapshot.status,
                        live: true,
                      ),
                    ),
                    Expanded(
                      child: SingleChildScrollView(
                        padding: const EdgeInsets.symmetric(
                          horizontal: NxSpacing.s4,
                        ),
                        child: steps.isEmpty
                            ? Text(
                                'Tracking updates will appear here as your order moves.',
                                style: NxTypography.bodyMd
                                    .copyWith(color: colors.textSecondary),
                              )
                            : NxTrackingTimeline(steps: steps),
                      ),
                    ),
                    SafeArea(
                      top: false,
                      child: Padding(
                        padding: const EdgeInsets.all(NxSpacing.s4),
                        child: Row(
                          children: [
                            Expanded(
                              child: NxButton(
                                label: l10n.callCourier,
                                variant: NxButtonVariant.secondary,
                                disabled: !snapshot.canCall ||
                                    (snapshot.courierPhone == null ||
                                        snapshot.courierPhone!.isEmpty),
                                onPressed: () =>
                                    _callCourier(snapshot.courierPhone),
                              ),
                            ),
                            const SizedBox(width: NxSpacing.s3),
                            Expanded(
                              child: NxButton(
                                label: l10n.chatSupport,
                                disabled: !snapshot.canChat,
                                onPressed: () =>
                                    _openCourierChat(context, snapshot),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                );
              },
              loading: () => const Center(child: NxSpinner()),
              error: (e, _) => ErrorView(
                message: e.toString(),
                onRetry: () {
                  ref.invalidate(trackingSnapshotProvider(orderId));
                  ref.invalidate(trackingRealtimeProvider(orderId));
                },
              ),
            ),
          ),
        ],
      ),
    );
  }
}
