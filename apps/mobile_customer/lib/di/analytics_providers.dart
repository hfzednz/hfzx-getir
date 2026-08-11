import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../shared/analytics/analytics_events.dart';
import '../shared/analytics/funnel_tracker.dart';
import 'providers.dart';

final analyticsTrackerProvider = Provider<AnalyticsTracker>((ref) {
  return AnalyticsTracker(
    gateway: ref.watch(analyticsProvider),
    cityIdProvider: () => ref.read(cityIdProvider),
  );
});

final funnelTrackerProvider = Provider<FunnelTracker>((ref) {
  return FunnelTracker(ref.watch(analyticsTrackerProvider));
});
