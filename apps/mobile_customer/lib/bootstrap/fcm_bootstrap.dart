import 'dart:async';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../features/notifications/presentation/providers/notifications_providers.dart';

/// Registers FCM token with BFF when Firebase is available.
Future<void> registerFcmIfAvailable(WidgetRef ref) async {
  try {
    if (Firebase.apps.isEmpty) return;

    final messaging = FirebaseMessaging.instance;
    await messaging.requestPermission(alert: true, badge: true, sound: true);

    final token = await messaging.getToken();
    if (token != null && token.isNotEmpty) {
      await ref.read(registerFcmTokenUseCaseProvider).call(token);
    }

    FirebaseMessaging.instance.onTokenRefresh.listen((nextToken) async {
      if (nextToken.isEmpty) return;
      await ref.read(registerFcmTokenUseCaseProvider).call(nextToken);
    });
  } catch (error, stackTrace) {
    if (kDebugMode) {
      debugPrint('FCM registration skipped: $error\n$stackTrace');
    }
  }
}
