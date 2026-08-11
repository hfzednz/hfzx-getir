import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/repositories/auth_repository_impl.dart';

final authRepositoryProvider = Provider<AuthRepositoryImpl>((ref) {
  return AuthRepositoryImpl(
    ref.watch(apiClientProvider),
    ref.watch(tokenStoreProvider),
    ref.watch(preferencesStoreProvider),
  );
});
