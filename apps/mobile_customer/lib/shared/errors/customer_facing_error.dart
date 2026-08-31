import 'package:nexora_core/nexora_core.dart';

/// Maps raw API / transport errors into customer-visible copy.
String customerFacingError(
  Object err, {
  String languageCode = 'en',
}) {
  final tr = languageCode.toLowerCase().startsWith('tr');
  final message = err is NexoraException ? err.message : err.toString();
  final code = err is NexoraException ? err.code.code.toLowerCase() : '';
  final lower = message.toLowerCase();

  if (code == 'unauthorized' ||
      code == 'auth_invalid' ||
      lower.contains('unauthorized')) {
    return tr
        ? 'Oturumunuz sona erdi. Lütfen tekrar giriş yapın.'
        : 'Your session expired. Please sign in again.';
  }
  if (lower.contains('coupon') &&
      (lower.contains('expir') || code == 'conflict')) {
    return tr
        ? 'Bu kuponun süresi dolmuş.'
        : 'This coupon has expired.';
  }
  if (lower.contains('coupon') &&
      (lower.contains('minimum') || lower.contains('basket'))) {
    return tr
        ? 'Bu kupon için sepet tutarı yetersiz.'
        : 'Your basket is below this coupon’s minimum.';
  }
  if (lower.contains('coupon') || lower.contains('promo code')) {
    return tr
        ? 'Kupon geçersiz. Lütfen kodu kontrol edin.'
        : 'This coupon is not valid. Check the code and try again.';
  }
  if (code == 'not_found' || lower.contains('not found')) {
    return tr
        ? 'İstediğiniz kayıt bulunamadı. Lütfen yenileyip tekrar deneyin.'
        : 'We could not find that record. Refresh and try again.';
  }
  if (code == 'conflict' || lower.contains('conflict')) {
    return tr
        ? 'Bu işlem zaten tamamlanmış. Siparişlerinizi kontrol edin.'
        : 'This action was already completed. Check your orders.';
  }
  if (code == 'network_error' ||
      lower.contains('socket') ||
      lower.contains('timeout') ||
      lower.contains('network')) {
    return tr
        ? 'Bağlantı kurulamadı. Lütfen internetinizi kontrol edip tekrar deneyin.'
        : 'Could not reach the server. Check your connection and try again.';
  }
  if (code == 'internal' ||
      code == 'upstream_failure' ||
      lower.contains('server error') ||
      lower.contains('upstream')) {
    return tr
        ? 'Şu anda işleminizi tamamlayamadık. Lütfen biraz sonra tekrar deneyin.'
        : 'We could not complete that just now. Please try again shortly.';
  }

  final staleCart = lower.contains('invalid argument') ||
      lower.contains('invalid_argument') ||
      lower.contains('cart required') ||
      lower.contains('checkout not ready') ||
      lower.contains('stale') ||
      code == 'invalid_argument' ||
      code == 'validation_failed';
  if (staleCart &&
      (lower.contains('cart') ||
          lower.contains('checkout') ||
          lower.contains('session') ||
          lower.contains('invalid argument') ||
          code == 'invalid_argument')) {
    return tr
        ? 'Sepet bilgileri güncel değil. Lütfen sepetinizi yenileyip tekrar deneyin.'
        : 'Your cart is out of date. Refresh the cart and try again.';
  }

  if (lower.contains('coupon') &&
      (lower.contains('expir') || code == 'conflict')) {
    return tr
        ? 'Bu kuponun süresi dolmuş.'
        : 'This coupon has expired.';
  }
  if (lower.contains('coupon') &&
      (lower.contains('minimum') || lower.contains('basket'))) {
    return tr
        ? 'Bu kupon için sepet tutarı yetersiz.'
        : 'Your basket is below this coupon’s minimum.';
  }
  if (lower.contains('coupon') || lower.contains('promo code')) {
    return tr
        ? 'Kupon geçersiz. Lütfen kodu kontrol edin.'
        : 'This coupon is not valid. Check the code and try again.';
  }
  if (lower.contains('min') &&
      (lower.contains('order') || lower.contains('basket'))) {
    return tr
        ? 'Minimum sepet tutarına ulaşmadınız.'
        : 'Your basket is below the minimum order amount.';
  }

  if (lower.contains('address')) {
    return tr
        ? 'Teslimat adresi eksik veya bu konumda hizmet veremiyoruz.'
        : 'Delivery address is missing or this location is not serviceable.';
  }
  if (lower.contains('payment')) {
    return tr
        ? 'Ödeme alınamadı. Lütfen başka bir yöntem deneyin veya tekrar deneyin.'
        : 'Payment did not go through. Try another method or retry.';
  }

  if (message.trim().isEmpty ||
      lower.contains('exception') ||
      lower == 'invalid argument') {
    return tr
        ? 'İşlem tamamlanamadı. Lütfen tekrar deneyin.'
        : 'Something went wrong. Please try again.';
  }
  return message;
}
