import 'dart:async';

import 'package:connectivity_plus/connectivity_plus.dart';

/// Observes device connectivity for offline-aware sync (CONSTITUTION §24).
class ConnectivityService {
  ConnectivityService({Connectivity? connectivity})
      : _connectivity = connectivity ?? Connectivity();

  final Connectivity _connectivity;

  Stream<List<ConnectivityResult>> get onConnectivityChanged =>
      _connectivity.onConnectivityChanged;

  Future<bool> get isOnline async {
    final results = await _connectivity.checkConnectivity();
    return _isOnline(results);
  }

  bool _isOnline(List<ConnectivityResult> results) {
    return results.any(
      (result) =>
          result == ConnectivityResult.mobile ||
          result == ConnectivityResult.wifi ||
          result == ConnectivityResult.ethernet ||
          result == ConnectivityResult.vpn,
    );
  }
}
