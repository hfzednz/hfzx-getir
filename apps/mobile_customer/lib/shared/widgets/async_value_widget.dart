import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../errors/error_copy.dart';
import 'error_view.dart';

typedef AsyncDataBuilder<T> = Widget Function(T data);

class AsyncValueWidget<T> extends StatelessWidget {
  const AsyncValueWidget({
    super.key,
    required this.value,
    required this.data,
    this.loading,
    this.error,
  });

  final AsyncValue<T> value;
  final AsyncDataBuilder<T> data;
  final Widget Function()? loading;
  final Widget Function(Object error, StackTrace stack)? error;

  @override
  Widget build(BuildContext context) {
    return value.when(
      data: (d) => data(d),
      loading: () => loading?.call() ?? const Center(child: NxSpinner()),
      error: (e, st) =>
          error?.call(e, st) ??
          ErrorView(message: localizedCustomerError(context, e)),
    );
  }
}
