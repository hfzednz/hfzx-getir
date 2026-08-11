/// Route path constants for GoRouter (ARCHITECTURE.md).
abstract final class RouteNames {
  static const splash = '/';
  static const onboarding = '/onboarding';
  static const auth = '/auth';
  static const authPhone = '/auth/phone';
  static const authOtp = '/auth/otp';
  static const authEmail = '/auth/email';
  static const authWelcome = '/auth/welcome';
  static const authForgotPassword = '/auth/forgot-password';
  static const authResetPassword = '/auth/reset-password';
  static const authEmailVerify = '/auth/email-verify';
  static const shell = '/shell';
  static const home = '/home';
  static const categories = '/categories';
  static const categoryDetail = '/categories/:categoryId';
  static const search = '/search';
  static const cart = '/cart';
  static const account = '/account';
  static const product = '/p/:productId';
  static const checkout = '/checkout';
  static const checkoutAddress = '/checkout/address';
  static const checkoutSchedule = '/checkout/schedule';
  static const checkoutPayment = '/checkout/payment';
  static const checkoutReview = '/checkout/review';
  static const orders = '/orders';
  static const orderDetail = '/orders/:orderId';
  static const orderTrack = '/orders/:orderId/track';
  static const orderReview = '/orders/:orderId/review';
  static const favorites = '/favorites';
  static const wallet = '/wallet';
  static const coupons = '/coupons';
  static const loyalty = '/loyalty';
  static const notifications = '/notifications';
  static const profile = '/profile';
  static const addresses = '/addresses';
  static const addressAdd = '/addresses/add';
  static const addressEdit = '/addresses/:addressId/edit';
  static const support = '/support';
  static const supportTicket = '/support/tickets';
  static const supportAssistant = '/support/assistant';
  static const settings = '/settings';
  static const settingsLanguage = '/settings/language';
  static const settingsTheme = '/settings/theme';
  static const settingsA11y = '/settings/accessibility';
  static const settingsNotifications = '/settings/notifications';
  static const settingsPrivacy = '/settings/privacy';
  static const settingsSecurity = '/settings/security';
  static const settingsDevices = '/settings/devices';
  static const settingsDeleteAccount = '/settings/delete-account';
  static const settingsPrivacyControls = '/settings/privacy-controls';
  static const referral = '/referral';
  static const help = '/help';
  static const about = '/about';
  static const legal = '/legal/:doc';
  static const campaign = '/c/:slug';
  static const promo = '/promo/:code';
  static const barcodeScanner = '/search/barcode';
  static const city = '/city';
  static const ai = '/ai';
  static const aiRecipes = '/ai/recipes';
}

/// Paths requiring authenticated session.
const authRequiredPaths = {
  RouteNames.checkout,
  RouteNames.checkoutAddress,
  RouteNames.checkoutSchedule,
  RouteNames.checkoutPayment,
  RouteNames.checkoutReview,
  RouteNames.orders,
  RouteNames.orderDetail,
  RouteNames.orderTrack,
  RouteNames.orderReview,
  RouteNames.wallet,
  RouteNames.profile,
  RouteNames.addresses,
  RouteNames.addressAdd,
  RouteNames.addressEdit,
  RouteNames.settings,
  RouteNames.settingsLanguage,
  RouteNames.settingsTheme,
  RouteNames.settingsA11y,
  RouteNames.settingsNotifications,
  RouteNames.settingsPrivacy,
  RouteNames.settingsSecurity,
  RouteNames.settingsDevices,
  RouteNames.settingsDeleteAccount,
  RouteNames.settingsPrivacyControls,
};
