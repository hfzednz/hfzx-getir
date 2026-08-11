import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

/// Normalized barcode / QR scan payload.
class BarcodeScanResult {
  const BarcodeScanResult({
    required this.rawValue,
    this.format,
    this.scannedAt,
  });

  final String rawValue;
  final String? format;
  final DateTime? scannedAt;

  factory BarcodeScanResult.fromBarcode(Barcode barcode) {
    return BarcodeScanResult(
      rawValue: barcode.rawValue ?? '',
      format: barcode.format.name,
      scannedAt: DateTime.now().toUtc(),
    );
  }
}

/// Camera barcode adapter over `mobile_scanner`.
class MobileScannerAdapter {
  MobileScannerAdapter({MobileScannerController? controller})
      : _controller = controller ?? MobileScannerController();

  final MobileScannerController _controller;

  MobileScannerController get controller => _controller;

  Stream<BarcodeScanResult> get scans => _controller.barcodes
      .expand((capture) => capture.barcodes)
      .where((b) => (b.rawValue ?? '').isNotEmpty)
      .map(BarcodeScanResult.fromBarcode);

  Future<void> start() => _controller.start();

  Future<void> stop() => _controller.stop();

  Future<void> dispose() => _controller.dispose();

  Widget buildPreview({
    required void Function(BarcodeScanResult result) onDetect,
    bool facingFront = false,
  }) {
    return MobileScanner(
      controller: _controller,
      onDetect: (capture) {
        for (final barcode in capture.barcodes) {
          final value = barcode.rawValue;
          if (value == null || value.isEmpty) continue;
          onDetect(BarcodeScanResult.fromBarcode(barcode));
          break;
        }
      },
    );
  }
}

/// Port for industrial RFID readers — wire vendor SDK later.
abstract class RfidScanPort {
  Stream<BarcodeScanResult> get tags;

  Future<void> connect();

  Future<void> disconnect();

  Future<void> startInventory();

  Future<void> stopInventory();
}

/// Stub RFID port until hardware SDK is integrated.
class UnimplementedRfidScanPort implements RfidScanPort {
  // Hardware SDK (e.g. Zebra / Impinj) to be plugged in here.

  @override
  Stream<BarcodeScanResult> get tags =>
      throw UnimplementedError('RFID SDK not integrated');

  @override
  Future<void> connect() =>
      throw UnimplementedError('RFID SDK not integrated');

  @override
  Future<void> disconnect() async {}

  @override
  Future<void> startInventory() =>
      throw UnimplementedError('RFID SDK not integrated');

  @override
  Future<void> stopInventory() async {}
}

/// Port for Bluetooth scan guns / wedge keyboards as HID alternatives.
abstract class BluetoothScanPort {
  Stream<BarcodeScanResult> get scans;

  Future<void> connect(String deviceId);

  Future<void> disconnect();

  Future<bool> isConnected();
}

/// Stub Bluetooth scan-gun port until vendor SDK / SPP profile is wired.
class UnimplementedBluetoothScanPort implements BluetoothScanPort {
  // Bluetooth SPP / vendor scan-gun SDK to be plugged in here.

  @override
  Stream<BarcodeScanResult> get scans =>
      throw UnimplementedError('Bluetooth scan gun SDK not integrated');

  @override
  Future<void> connect(String deviceId) =>
      throw UnimplementedError('Bluetooth scan gun SDK not integrated');

  @override
  Future<void> disconnect() async {}

  @override
  Future<bool> isConnected() async => false;
}
