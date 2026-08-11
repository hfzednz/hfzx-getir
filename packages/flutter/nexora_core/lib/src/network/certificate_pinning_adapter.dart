import 'dart:io';

import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import 'package:logging/logging.dart';

/// Optional TLS certificate pinning adapter. No-op when [pins] is empty.
class CertificatePinningAdapter {
  CertificatePinningAdapter({
    required this.pins,
    Logger? logger,
  }) : _logger = logger ?? Logger('CertificatePinningAdapter');

  final List<String> pins;
  final Logger _logger;

  bool get isEnabled => pins.isNotEmpty;

  void configure(Dio dio) {
    if (!isEnabled) {
      return;
    }

    final adapter = dio.httpClientAdapter;
    if (adapter is! IOHttpClientAdapter) {
      _logger.warning('Certificate pinning requires IOHttpClientAdapter');
      return;
    }

    adapter.createHttpClient = () {
      final client = HttpClient();
      client.badCertificateCallback = (cert, host, port) {
        final fingerprint = _sha256Fingerprint(cert.der);
        final allowed = pins.contains(fingerprint);
        if (!allowed) {
          _logger.severe(
            'Certificate pin mismatch for $host:$port',
          );
        }
        return allowed;
      };
      return client;
    };
  }

  static String _sha256Fingerprint(List<int> der) {
    final digest = sha256.convert(der);
    return digest.bytes
        .map((b) => b.toRadixString(16).padLeft(2, '0'))
        .join(':')
        .toUpperCase();
  }
}
