import 'nexora_exception.dart';

/// Lightweight Result type (Either-style) for domain operations.
sealed class Result<T> {
  const Result();

  bool get isSuccess => this is Success<T>;
  bool get isFailure => this is Failure<T>;

  T? get valueOrNull => switch (this) {
        Success<T>(:final value) => value,
        Failure<T>() => null,
      };

  NexoraException? get errorOrNull => switch (this) {
        Success<T>() => null,
        Failure<T>(:final error) => error,
      };

  Result<R> map<R>(R Function(T value) transform) => switch (this) {
        Success<T>(:final value) => Success(transform(value)),
        Failure<T>(:final error) => Failure(error),
      };

  Result<R> flatMap<R>(Result<R> Function(T value) transform) => switch (this) {
        Success<T>(:final value) => transform(value),
        Failure<T>(:final error) => Failure(error),
      };

  R fold<R>({
    required R Function(T value) onSuccess,
    required R Function(NexoraException error) onFailure,
  }) =>
      switch (this) {
        Success<T>(:final value) => onSuccess(value),
        Failure<T>(:final error) => onFailure(error),
      };
}

final class Success<T> extends Result<T> {
  const Success(this.value);

  final T value;
}

final class Failure<T> extends Result<T> {
  const Failure(this.error);

  final NexoraException error;
}
