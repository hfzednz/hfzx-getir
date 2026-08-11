/// Device integrity signals for root/jailbreak detection (CONSTITUTION §27).
abstract class DeviceIntegrityChecker {
  Future<DeviceIntegrityReport> evaluate();
}

class DeviceIntegrityReport {
  const DeviceIntegrityReport({
    required this.isCompromised,
    this.signals = const {},
  });

  final bool isCompromised;
  final Map<String, bool> signals;
}

/// Default stub — always reports safe until wired to Play Integrity / App Attest.
class StubDeviceIntegrityChecker implements DeviceIntegrityChecker {
  const StubDeviceIntegrityChecker({
    this.isRooted = false,
    this.isJailbroken = false,
    this.isEmulator = false,
    this.isDebuggable = false,
  });

  final bool isRooted;
  final bool isJailbroken;
  final bool isEmulator;
  final bool isDebuggable;

  @override
  Future<DeviceIntegrityReport> evaluate() async {
    final signals = {
      'rooted': isRooted,
      'jailbroken': isJailbroken,
      'emulator': isEmulator,
      'debuggable': isDebuggable,
    };
    return DeviceIntegrityReport(
      isCompromised: signals.values.any((value) => value),
      signals: signals,
    );
  }
}

/// Composite checker that OR-merges multiple providers.
class CompositeDeviceIntegrityChecker implements DeviceIntegrityChecker {
  CompositeDeviceIntegrityChecker(this.checkers);

  final List<DeviceIntegrityChecker> checkers;

  @override
  Future<DeviceIntegrityReport> evaluate() async {
    final mergedSignals = <String, bool>{};
    for (final checker in checkers) {
      final report = await checker.evaluate();
      mergedSignals.addAll(report.signals);
    }
    return DeviceIntegrityReport(
      isCompromised: mergedSignals.values.any((value) => value),
      signals: mergedSignals,
    );
  }
}
