import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_tr.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('tr')
  ];

  /// No description provided for @appTitle.
  ///
  /// In en, this message translates to:
  /// **'NEXORA'**
  String get appTitle;

  /// No description provided for @homeTitle.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get homeTitle;

  /// No description provided for @categoriesTitle.
  ///
  /// In en, this message translates to:
  /// **'Categories'**
  String get categoriesTitle;

  /// No description provided for @searchTitle.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get searchTitle;

  /// No description provided for @cartTitle.
  ///
  /// In en, this message translates to:
  /// **'Cart'**
  String get cartTitle;

  /// No description provided for @accountTitle.
  ///
  /// In en, this message translates to:
  /// **'Account'**
  String get accountTitle;

  /// No description provided for @splashTitle.
  ///
  /// In en, this message translates to:
  /// **'NEXORA'**
  String get splashTitle;

  /// No description provided for @onboardingTitle.
  ///
  /// In en, this message translates to:
  /// **'Welcome'**
  String get onboardingTitle;

  /// No description provided for @authTitle.
  ///
  /// In en, this message translates to:
  /// **'Sign in'**
  String get authTitle;

  /// No description provided for @cityTitle.
  ///
  /// In en, this message translates to:
  /// **'City'**
  String get cityTitle;

  /// No description provided for @productTitle.
  ///
  /// In en, this message translates to:
  /// **'Product'**
  String get productTitle;

  /// No description provided for @checkoutTitle.
  ///
  /// In en, this message translates to:
  /// **'Checkout'**
  String get checkoutTitle;

  /// No description provided for @ordersTitle.
  ///
  /// In en, this message translates to:
  /// **'Orders'**
  String get ordersTitle;

  /// No description provided for @trackingTitle.
  ///
  /// In en, this message translates to:
  /// **'Track order'**
  String get trackingTitle;

  /// No description provided for @favoritesTitle.
  ///
  /// In en, this message translates to:
  /// **'Favorites'**
  String get favoritesTitle;

  /// No description provided for @walletTitle.
  ///
  /// In en, this message translates to:
  /// **'Wallet'**
  String get walletTitle;

  /// No description provided for @couponsTitle.
  ///
  /// In en, this message translates to:
  /// **'Coupons'**
  String get couponsTitle;

  /// No description provided for @loyaltyTitle.
  ///
  /// In en, this message translates to:
  /// **'Loyalty'**
  String get loyaltyTitle;

  /// No description provided for @notificationsTitle.
  ///
  /// In en, this message translates to:
  /// **'Notifications'**
  String get notificationsTitle;

  /// No description provided for @profileTitle.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profileTitle;

  /// No description provided for @addressesTitle.
  ///
  /// In en, this message translates to:
  /// **'Addresses'**
  String get addressesTitle;

  /// No description provided for @reviewsTitle.
  ///
  /// In en, this message translates to:
  /// **'Reviews'**
  String get reviewsTitle;

  /// No description provided for @supportTitle.
  ///
  /// In en, this message translates to:
  /// **'Support'**
  String get supportTitle;

  /// No description provided for @settingsTitle.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settingsTitle;

  /// No description provided for @referralTitle.
  ///
  /// In en, this message translates to:
  /// **'Refer a friend'**
  String get referralTitle;

  /// No description provided for @helpTitle.
  ///
  /// In en, this message translates to:
  /// **'Help'**
  String get helpTitle;

  /// No description provided for @aboutTitle.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get aboutTitle;

  /// No description provided for @legalTitle.
  ///
  /// In en, this message translates to:
  /// **'Legal'**
  String get legalTitle;

  /// No description provided for @emptyTitle.
  ///
  /// In en, this message translates to:
  /// **'Nothing here yet'**
  String get emptyTitle;

  /// No description provided for @emptyMessage.
  ///
  /// In en, this message translates to:
  /// **'Check back later or try again.'**
  String get emptyMessage;

  /// No description provided for @retry.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retry;

  /// No description provided for @continueLabel.
  ///
  /// In en, this message translates to:
  /// **'Continue'**
  String get continueLabel;

  /// No description provided for @skip.
  ///
  /// In en, this message translates to:
  /// **'Skip'**
  String get skip;

  /// No description provided for @guestContinue.
  ///
  /// In en, this message translates to:
  /// **'Continue as guest'**
  String get guestContinue;

  /// No description provided for @phoneLogin.
  ///
  /// In en, this message translates to:
  /// **'Phone number'**
  String get phoneLogin;

  /// No description provided for @emailLogin.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get emailLogin;

  /// No description provided for @googleLogin.
  ///
  /// In en, this message translates to:
  /// **'Continue with Google'**
  String get googleLogin;

  /// No description provided for @appleLogin.
  ///
  /// In en, this message translates to:
  /// **'Continue with Apple'**
  String get appleLogin;

  /// No description provided for @otpTitle.
  ///
  /// In en, this message translates to:
  /// **'Enter verification code'**
  String get otpTitle;

  /// No description provided for @otpHint.
  ///
  /// In en, this message translates to:
  /// **'6-digit code'**
  String get otpHint;

  /// No description provided for @sendOtp.
  ///
  /// In en, this message translates to:
  /// **'Send code'**
  String get sendOtp;

  /// No description provided for @verifyOtp.
  ///
  /// In en, this message translates to:
  /// **'Verify'**
  String get verifyOtp;

  /// No description provided for @searchHint.
  ///
  /// In en, this message translates to:
  /// **'Search products, brands…'**
  String get searchHint;

  /// No description provided for @voiceSearch.
  ///
  /// In en, this message translates to:
  /// **'Voice search'**
  String get voiceSearch;

  /// No description provided for @barcodeScan.
  ///
  /// In en, this message translates to:
  /// **'Scan barcode'**
  String get barcodeScan;

  /// No description provided for @imageSearch.
  ///
  /// In en, this message translates to:
  /// **'Search by image'**
  String get imageSearch;

  /// No description provided for @addToCart.
  ///
  /// In en, this message translates to:
  /// **'Add to cart'**
  String get addToCart;

  /// No description provided for @checkout.
  ///
  /// In en, this message translates to:
  /// **'Checkout'**
  String get checkout;

  /// No description provided for @placeOrder.
  ///
  /// In en, this message translates to:
  /// **'Place order'**
  String get placeOrder;

  /// No description provided for @orderConfirmed.
  ///
  /// In en, this message translates to:
  /// **'Order confirmed'**
  String get orderConfirmed;

  /// No description provided for @deliveryEta.
  ///
  /// In en, this message translates to:
  /// **'Estimated delivery'**
  String get deliveryEta;

  /// No description provided for @offlineMessage.
  ///
  /// In en, this message translates to:
  /// **'You are offline'**
  String get offlineMessage;

  /// No description provided for @loginRequired.
  ///
  /// In en, this message translates to:
  /// **'Sign in required'**
  String get loginRequired;

  /// No description provided for @loginRequiredMessage.
  ///
  /// In en, this message translates to:
  /// **'Please sign in to continue.'**
  String get loginRequiredMessage;

  /// No description provided for @signIn.
  ///
  /// In en, this message translates to:
  /// **'Sign in'**
  String get signIn;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// No description provided for @theme.
  ///
  /// In en, this message translates to:
  /// **'Theme'**
  String get theme;

  /// No description provided for @themeLight.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get themeDark;

  /// No description provided for @themeSystem.
  ///
  /// In en, this message translates to:
  /// **'System'**
  String get themeSystem;

  /// No description provided for @biometricUnlock.
  ///
  /// In en, this message translates to:
  /// **'Biometric unlock'**
  String get biometricUnlock;

  /// No description provided for @deleteAccount.
  ///
  /// In en, this message translates to:
  /// **'Delete account'**
  String get deleteAccount;

  /// No description provided for @privacy.
  ///
  /// In en, this message translates to:
  /// **'Privacy'**
  String get privacy;

  /// No description provided for @security.
  ///
  /// In en, this message translates to:
  /// **'Security'**
  String get security;

  /// No description provided for @locationPermissionTitle.
  ///
  /// In en, this message translates to:
  /// **'Enable location'**
  String get locationPermissionTitle;

  /// No description provided for @locationPermissionMessage.
  ///
  /// In en, this message translates to:
  /// **'We use your location to show delivery times and available products.'**
  String get locationPermissionMessage;

  /// No description provided for @enableLocation.
  ///
  /// In en, this message translates to:
  /// **'Enable location'**
  String get enableLocation;

  /// No description provided for @subtotal.
  ///
  /// In en, this message translates to:
  /// **'Subtotal'**
  String get subtotal;

  /// No description provided for @deliveryFee.
  ///
  /// In en, this message translates to:
  /// **'Delivery fee'**
  String get deliveryFee;

  /// No description provided for @total.
  ///
  /// In en, this message translates to:
  /// **'Total'**
  String get total;

  /// No description provided for @applyCoupon.
  ///
  /// In en, this message translates to:
  /// **'Apply coupon'**
  String get applyCoupon;

  /// No description provided for @orderNotes.
  ///
  /// In en, this message translates to:
  /// **'Order notes'**
  String get orderNotes;

  /// No description provided for @contactlessDelivery.
  ///
  /// In en, this message translates to:
  /// **'Contactless delivery'**
  String get contactlessDelivery;

  /// No description provided for @giftOrder.
  ///
  /// In en, this message translates to:
  /// **'Gift order'**
  String get giftOrder;

  /// No description provided for @paymentMethod.
  ///
  /// In en, this message translates to:
  /// **'Payment method'**
  String get paymentMethod;

  /// No description provided for @scheduleDelivery.
  ///
  /// In en, this message translates to:
  /// **'Schedule delivery'**
  String get scheduleDelivery;

  /// No description provided for @selectAddress.
  ///
  /// In en, this message translates to:
  /// **'Select address'**
  String get selectAddress;

  /// No description provided for @addAddress.
  ///
  /// In en, this message translates to:
  /// **'Add address'**
  String get addAddress;

  /// No description provided for @callCourier.
  ///
  /// In en, this message translates to:
  /// **'Call courier'**
  String get callCourier;

  /// No description provided for @chatSupport.
  ///
  /// In en, this message translates to:
  /// **'Chat support'**
  String get chatSupport;

  /// No description provided for @reorder.
  ///
  /// In en, this message translates to:
  /// **'Reorder'**
  String get reorder;

  /// No description provided for @pointsBalance.
  ///
  /// In en, this message translates to:
  /// **'Points balance'**
  String get pointsBalance;

  /// No description provided for @referralCode.
  ///
  /// In en, this message translates to:
  /// **'Your referral code'**
  String get referralCode;

  /// No description provided for @shareReferral.
  ///
  /// In en, this message translates to:
  /// **'Share code'**
  String get shareReferral;

  /// No description provided for @termsOfService.
  ///
  /// In en, this message translates to:
  /// **'Terms of Service'**
  String get termsOfService;

  /// No description provided for @privacyPolicy.
  ///
  /// In en, this message translates to:
  /// **'Privacy Policy'**
  String get privacyPolicy;

  /// No description provided for @resendOtp.
  ///
  /// In en, this message translates to:
  /// **'Resend code'**
  String get resendOtp;

  /// No description provided for @resendOtpIn.
  ///
  /// In en, this message translates to:
  /// **'Resend in {seconds}s'**
  String resendOtpIn(int seconds);

  /// No description provided for @enterPhone.
  ///
  /// In en, this message translates to:
  /// **'Enter your phone number'**
  String get enterPhone;

  /// No description provided for @otpSendFailed.
  ///
  /// In en, this message translates to:
  /// **'Could not send a verification code. Please try again.'**
  String get otpSendFailed;

  /// No description provided for @signOut.
  ///
  /// In en, this message translates to:
  /// **'Sign out'**
  String get signOut;

  /// No description provided for @storesTitle.
  ///
  /// In en, this message translates to:
  /// **'Stores'**
  String get storesTitle;

  /// No description provided for @storeOpen.
  ///
  /// In en, this message translates to:
  /// **'Open'**
  String get storeOpen;

  /// No description provided for @storeClosed.
  ///
  /// In en, this message translates to:
  /// **'Closed'**
  String get storeClosed;

  /// No description provided for @payWithCard.
  ///
  /// In en, this message translates to:
  /// **'Card'**
  String get payWithCard;

  /// No description provided for @payWithCardHint.
  ///
  /// In en, this message translates to:
  /// **'Pay with a test card'**
  String get payWithCardHint;

  /// No description provided for @cashOnDelivery.
  ///
  /// In en, this message translates to:
  /// **'Cash on delivery'**
  String get cashOnDelivery;

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @submit.
  ///
  /// In en, this message translates to:
  /// **'Submit'**
  String get submit;

  /// No description provided for @firstName.
  ///
  /// In en, this message translates to:
  /// **'First name'**
  String get firstName;

  /// No description provided for @lastName.
  ///
  /// In en, this message translates to:
  /// **'Last name'**
  String get lastName;

  /// No description provided for @displayName.
  ///
  /// In en, this message translates to:
  /// **'Display name'**
  String get displayName;

  /// No description provided for @email.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get email;

  /// No description provided for @phone.
  ///
  /// In en, this message translates to:
  /// **'Phone'**
  String get phone;

  /// No description provided for @communicationPreferences.
  ///
  /// In en, this message translates to:
  /// **'Communication preferences'**
  String get communicationPreferences;

  /// No description provided for @emailMarketing.
  ///
  /// In en, this message translates to:
  /// **'Email marketing'**
  String get emailMarketing;

  /// No description provided for @orderUpdatesPush.
  ///
  /// In en, this message translates to:
  /// **'Order updates (push)'**
  String get orderUpdatesPush;

  /// No description provided for @promotionsPush.
  ///
  /// In en, this message translates to:
  /// **'Promotions (push)'**
  String get promotionsPush;

  /// No description provided for @smsAlerts.
  ///
  /// In en, this message translates to:
  /// **'SMS alerts'**
  String get smsAlerts;

  /// No description provided for @privacyControls.
  ///
  /// In en, this message translates to:
  /// **'Privacy controls'**
  String get privacyControls;

  /// No description provided for @recipientName.
  ///
  /// In en, this message translates to:
  /// **'Recipient name'**
  String get recipientName;

  /// No description provided for @recipientPhone.
  ///
  /// In en, this message translates to:
  /// **'Recipient phone'**
  String get recipientPhone;

  /// No description provided for @favorite.
  ///
  /// In en, this message translates to:
  /// **'Favorite'**
  String get favorite;

  /// No description provided for @unfavorite.
  ///
  /// In en, this message translates to:
  /// **'Remove favorite'**
  String get unfavorite;

  /// No description provided for @defaultAddress.
  ///
  /// In en, this message translates to:
  /// **'Default address'**
  String get defaultAddress;

  /// No description provided for @pickOnMap.
  ///
  /// In en, this message translates to:
  /// **'Pick on map'**
  String get pickOnMap;

  /// No description provided for @changeLocation.
  ///
  /// In en, this message translates to:
  /// **'Change location'**
  String get changeLocation;

  /// No description provided for @saveChanges.
  ///
  /// In en, this message translates to:
  /// **'Save changes'**
  String get saveChanges;

  /// No description provided for @saveAddress.
  ///
  /// In en, this message translates to:
  /// **'Save address'**
  String get saveAddress;

  /// No description provided for @deliveryNotAvailableHere.
  ///
  /// In en, this message translates to:
  /// **'Delivery may not be available at this location.'**
  String get deliveryNotAvailableHere;

  /// No description provided for @newTicket.
  ///
  /// In en, this message translates to:
  /// **'New ticket'**
  String get newTicket;

  /// No description provided for @faqTitle.
  ///
  /// In en, this message translates to:
  /// **'FAQ'**
  String get faqTitle;

  /// No description provided for @yourTickets.
  ///
  /// In en, this message translates to:
  /// **'Your tickets'**
  String get yourTickets;

  /// No description provided for @createSupportTicket.
  ///
  /// In en, this message translates to:
  /// **'Create a support ticket'**
  String get createSupportTicket;

  /// No description provided for @subject.
  ///
  /// In en, this message translates to:
  /// **'Subject'**
  String get subject;

  /// No description provided for @message.
  ///
  /// In en, this message translates to:
  /// **'Message'**
  String get message;

  /// No description provided for @partialCancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel selected items'**
  String get partialCancel;

  /// No description provided for @requestRefund.
  ///
  /// In en, this message translates to:
  /// **'Request a refund'**
  String get requestRefund;

  /// No description provided for @cancelOrder.
  ///
  /// In en, this message translates to:
  /// **'Cancel order'**
  String get cancelOrder;

  /// No description provided for @cannotCancelOrder.
  ///
  /// In en, this message translates to:
  /// **'This order can no longer be cancelled.'**
  String get cannotCancelOrder;

  /// No description provided for @trackOrder.
  ///
  /// In en, this message translates to:
  /// **'Track order'**
  String get trackOrder;

  /// No description provided for @payInFull.
  ///
  /// In en, this message translates to:
  /// **'Pay in full'**
  String get payInFull;

  /// No description provided for @installments.
  ///
  /// In en, this message translates to:
  /// **'Installments'**
  String get installments;

  /// No description provided for @defaultCard.
  ///
  /// In en, this message translates to:
  /// **'Default card'**
  String get defaultCard;

  /// No description provided for @noSavedCards.
  ///
  /// In en, this message translates to:
  /// **'No saved cards yet. Pay with cash or the staging test card.'**
  String get noSavedCards;

  /// No description provided for @recentlyOrdered.
  ///
  /// In en, this message translates to:
  /// **'Recently ordered'**
  String get recentlyOrdered;

  /// No description provided for @frequentlyPurchased.
  ///
  /// In en, this message translates to:
  /// **'Frequently purchased'**
  String get frequentlyPurchased;

  /// No description provided for @popularProducts.
  ///
  /// In en, this message translates to:
  /// **'Popular'**
  String get popularProducts;

  /// No description provided for @recommendedProducts.
  ///
  /// In en, this message translates to:
  /// **'Recommended for you'**
  String get recommendedProducts;

  /// No description provided for @nearbyStores.
  ///
  /// In en, this message translates to:
  /// **'Nearby stores'**
  String get nearbyStores;

  /// No description provided for @campaignsTitle.
  ///
  /// In en, this message translates to:
  /// **'Offers'**
  String get campaignsTitle;

  /// No description provided for @minOrder.
  ///
  /// In en, this message translates to:
  /// **'Minimum order'**
  String get minOrder;

  /// No description provided for @outOfStock.
  ///
  /// In en, this message translates to:
  /// **'Out of stock'**
  String get outOfStock;

  /// No description provided for @inStock.
  ///
  /// In en, this message translates to:
  /// **'In stock'**
  String get inStock;

  /// No description provided for @storeClosedBody.
  ///
  /// In en, this message translates to:
  /// **'This store is closed right now. You can browse, but ordering will be available when it reopens.'**
  String get storeClosedBody;

  /// No description provided for @trackingUpdatesSoon.
  ///
  /// In en, this message translates to:
  /// **'Tracking updates will appear here as your order moves.'**
  String get trackingUpdatesSoon;

  /// No description provided for @calculatingEta.
  ///
  /// In en, this message translates to:
  /// **'Calculating arrival time'**
  String get calculatingEta;

  /// No description provided for @courierLabel.
  ///
  /// In en, this message translates to:
  /// **'Courier'**
  String get courierLabel;

  /// No description provided for @markAllRead.
  ///
  /// In en, this message translates to:
  /// **'Mark all as read'**
  String get markAllRead;

  /// No description provided for @noNotifications.
  ///
  /// In en, this message translates to:
  /// **'No notifications yet.'**
  String get noNotifications;

  /// No description provided for @emptyCart.
  ///
  /// In en, this message translates to:
  /// **'Your cart is empty.'**
  String get emptyCart;

  /// No description provided for @retryPayment.
  ///
  /// In en, this message translates to:
  /// **'Retry payment'**
  String get retryPayment;

  /// No description provided for @paymentFailed.
  ///
  /// In en, this message translates to:
  /// **'Payment did not go through. Try again or choose another method.'**
  String get paymentFailed;

  /// No description provided for @sectionEmpty.
  ///
  /// In en, this message translates to:
  /// **'Nothing to show here yet.'**
  String get sectionEmpty;

  /// No description provided for @accessibility.
  ///
  /// In en, this message translates to:
  /// **'Accessibility'**
  String get accessibility;

  /// No description provided for @highContrast.
  ///
  /// In en, this message translates to:
  /// **'High contrast'**
  String get highContrast;

  /// No description provided for @reduceMotion.
  ///
  /// In en, this message translates to:
  /// **'Reduce motion'**
  String get reduceMotion;

  /// No description provided for @textSize.
  ///
  /// In en, this message translates to:
  /// **'Text size'**
  String get textSize;

  /// No description provided for @addressLabel.
  ///
  /// In en, this message translates to:
  /// **'Label'**
  String get addressLabel;

  /// No description provided for @addressHome.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get addressHome;

  /// No description provided for @addressWork.
  ///
  /// In en, this message translates to:
  /// **'Work'**
  String get addressWork;

  /// No description provided for @addressCustom.
  ///
  /// In en, this message translates to:
  /// **'Custom'**
  String get addressCustom;

  /// No description provided for @customLabel.
  ///
  /// In en, this message translates to:
  /// **'Custom label'**
  String get customLabel;

  /// No description provided for @streetAddress.
  ///
  /// In en, this message translates to:
  /// **'Street address'**
  String get streetAddress;

  /// No description provided for @editAddress.
  ///
  /// In en, this message translates to:
  /// **'Edit address'**
  String get editAddress;

  /// No description provided for @addressSaved.
  ///
  /// In en, this message translates to:
  /// **'Address saved'**
  String get addressSaved;

  /// No description provided for @addressUpdated.
  ///
  /// In en, this message translates to:
  /// **'Address updated'**
  String get addressUpdated;

  /// No description provided for @building.
  ///
  /// In en, this message translates to:
  /// **'Building'**
  String get building;

  /// No description provided for @floor.
  ///
  /// In en, this message translates to:
  /// **'Floor'**
  String get floor;

  /// No description provided for @doorApt.
  ///
  /// In en, this message translates to:
  /// **'Apartment / door'**
  String get doorApt;

  /// No description provided for @deliveryInstructions.
  ///
  /// In en, this message translates to:
  /// **'Delivery instructions'**
  String get deliveryInstructions;

  /// No description provided for @invoice.
  ///
  /// In en, this message translates to:
  /// **'Invoice'**
  String get invoice;

  /// No description provided for @receipt.
  ///
  /// In en, this message translates to:
  /// **'Receipt'**
  String get receipt;

  /// No description provided for @proofOfDelivery.
  ///
  /// In en, this message translates to:
  /// **'Proof of delivery'**
  String get proofOfDelivery;

  /// No description provided for @trackingCreated.
  ///
  /// In en, this message translates to:
  /// **'Order created'**
  String get trackingCreated;

  /// No description provided for @trackingWarehouse.
  ///
  /// In en, this message translates to:
  /// **'Warehouse assigned'**
  String get trackingWarehouse;

  /// No description provided for @trackingPicking.
  ///
  /// In en, this message translates to:
  /// **'Picking'**
  String get trackingPicking;

  /// No description provided for @trackingPacking.
  ///
  /// In en, this message translates to:
  /// **'Packing'**
  String get trackingPacking;

  /// No description provided for @trackingReady.
  ///
  /// In en, this message translates to:
  /// **'Ready for dispatch'**
  String get trackingReady;

  /// No description provided for @trackingCourier.
  ///
  /// In en, this message translates to:
  /// **'Courier assigned'**
  String get trackingCourier;

  /// No description provided for @trackingOut.
  ///
  /// In en, this message translates to:
  /// **'Out for delivery'**
  String get trackingOut;

  /// No description provided for @trackingCompleted.
  ///
  /// In en, this message translates to:
  /// **'Delivered'**
  String get trackingCompleted;

  /// No description provided for @apply.
  ///
  /// In en, this message translates to:
  /// **'Apply'**
  String get apply;

  /// No description provided for @remove.
  ///
  /// In en, this message translates to:
  /// **'Remove'**
  String get remove;

  /// No description provided for @giftCardCode.
  ///
  /// In en, this message translates to:
  /// **'Gift card code'**
  String get giftCardCode;

  /// No description provided for @walletAmount.
  ///
  /// In en, this message translates to:
  /// **'Wallet amount (₺)'**
  String get walletAmount;

  /// No description provided for @loyaltyRedeem.
  ///
  /// In en, this message translates to:
  /// **'Loyalty points to redeem'**
  String get loyaltyRedeem;

  /// No description provided for @redeem.
  ///
  /// In en, this message translates to:
  /// **'Redeem'**
  String get redeem;

  /// No description provided for @tax.
  ///
  /// In en, this message translates to:
  /// **'Tax'**
  String get tax;

  /// No description provided for @discount.
  ///
  /// In en, this message translates to:
  /// **'Discount'**
  String get discount;

  /// No description provided for @couponsAndPayments.
  ///
  /// In en, this message translates to:
  /// **'Coupons and payments'**
  String get couponsAndPayments;

  /// No description provided for @promotions.
  ///
  /// In en, this message translates to:
  /// **'Promotions'**
  String get promotions;

  /// No description provided for @couponApplied.
  ///
  /// In en, this message translates to:
  /// **'Coupon applied'**
  String get couponApplied;

  /// No description provided for @couponInvalid.
  ///
  /// In en, this message translates to:
  /// **'This coupon is not valid'**
  String get couponInvalid;

  /// No description provided for @validateInventory.
  ///
  /// In en, this message translates to:
  /// **'Check availability'**
  String get validateInventory;

  /// No description provided for @highContrastHint.
  ///
  /// In en, this message translates to:
  /// **'Increase contrast for text and controls'**
  String get highContrastHint;

  /// No description provided for @reduceMotionHint.
  ///
  /// In en, this message translates to:
  /// **'Prefer less animation where supported'**
  String get reduceMotionHint;

  /// No description provided for @textSizeFollowsDevice.
  ///
  /// In en, this message translates to:
  /// **'Follows your device text size. Change it in system accessibility settings.'**
  String get textSizeFollowsDevice;

  /// No description provided for @savePreferences.
  ///
  /// In en, this message translates to:
  /// **'Save preferences'**
  String get savePreferences;

  /// No description provided for @preferencesSaved.
  ///
  /// In en, this message translates to:
  /// **'Preferences saved'**
  String get preferencesSaved;

  /// No description provided for @notificationPreferences.
  ///
  /// In en, this message translates to:
  /// **'Notification preferences'**
  String get notificationPreferences;

  /// No description provided for @pushNotifications.
  ///
  /// In en, this message translates to:
  /// **'Push notifications'**
  String get pushNotifications;

  /// No description provided for @emailNotifications.
  ///
  /// In en, this message translates to:
  /// **'Email notifications'**
  String get emailNotifications;

  /// No description provided for @orderUpdates.
  ///
  /// In en, this message translates to:
  /// **'Order updates'**
  String get orderUpdates;

  /// No description provided for @transactional.
  ///
  /// In en, this message translates to:
  /// **'Order and payment messages'**
  String get transactional;

  /// No description provided for @deliveryAlerts.
  ///
  /// In en, this message translates to:
  /// **'Delivery alerts'**
  String get deliveryAlerts;

  /// No description provided for @promotionsPref.
  ///
  /// In en, this message translates to:
  /// **'Promotions'**
  String get promotionsPref;

  /// No description provided for @priceDrops.
  ///
  /// In en, this message translates to:
  /// **'Price drops'**
  String get priceDrops;

  /// No description provided for @backInStock.
  ///
  /// In en, this message translates to:
  /// **'Back in stock'**
  String get backInStock;

  /// No description provided for @trustedDevices.
  ///
  /// In en, this message translates to:
  /// **'Trusted devices'**
  String get trustedDevices;

  /// No description provided for @trustedDevicesHint.
  ///
  /// In en, this message translates to:
  /// **'Review and sign out devices'**
  String get trustedDevicesHint;

  /// No description provided for @biometricHint.
  ///
  /// In en, this message translates to:
  /// **'Unlock the app with your face or fingerprint'**
  String get biometricHint;

  /// No description provided for @privacySubtitle.
  ///
  /// In en, this message translates to:
  /// **'Marketing, analytics, and personalization'**
  String get privacySubtitle;

  /// No description provided for @cookiePolicy.
  ///
  /// In en, this message translates to:
  /// **'Cookie policy'**
  String get cookiePolicy;

  /// No description provided for @marketingEmail.
  ///
  /// In en, this message translates to:
  /// **'Marketing email'**
  String get marketingEmail;

  /// No description provided for @smsMarketing.
  ///
  /// In en, this message translates to:
  /// **'SMS marketing'**
  String get smsMarketing;

  /// No description provided for @personalization.
  ///
  /// In en, this message translates to:
  /// **'Personalization'**
  String get personalization;

  /// No description provided for @optOutAnalytics.
  ///
  /// In en, this message translates to:
  /// **'Opt out of analytics'**
  String get optOutAnalytics;

  /// No description provided for @shareWithPartners.
  ///
  /// In en, this message translates to:
  /// **'Share with partners'**
  String get shareWithPartners;

  /// No description provided for @comment.
  ///
  /// In en, this message translates to:
  /// **'Comment'**
  String get comment;

  /// No description provided for @addPhotos.
  ///
  /// In en, this message translates to:
  /// **'Add photos'**
  String get addPhotos;

  /// No description provided for @submitReview.
  ///
  /// In en, this message translates to:
  /// **'Submit review'**
  String get submitReview;

  /// No description provided for @inviteCode.
  ///
  /// In en, this message translates to:
  /// **'Invite code'**
  String get inviteCode;

  /// No description provided for @claimInvite.
  ///
  /// In en, this message translates to:
  /// **'Claim invite'**
  String get claimInvite;

  /// No description provided for @stackable.
  ///
  /// In en, this message translates to:
  /// **'Can be combined with other offers'**
  String get stackable;

  /// No description provided for @applyFilters.
  ///
  /// In en, this message translates to:
  /// **'Apply filters'**
  String get applyFilters;

  /// No description provided for @brandsComma.
  ///
  /// In en, this message translates to:
  /// **'Brands (comma-separated)'**
  String get brandsComma;

  /// No description provided for @sort.
  ///
  /// In en, this message translates to:
  /// **'Sort'**
  String get sort;

  /// No description provided for @availabilityFilter.
  ///
  /// In en, this message translates to:
  /// **'Availability'**
  String get availabilityFilter;

  /// No description provided for @chooseFromGallery.
  ///
  /// In en, this message translates to:
  /// **'Choose from gallery'**
  String get chooseFromGallery;

  /// No description provided for @takePhoto.
  ///
  /// In en, this message translates to:
  /// **'Take a photo'**
  String get takePhoto;

  /// No description provided for @filters.
  ///
  /// In en, this message translates to:
  /// **'Filters'**
  String get filters;

  /// No description provided for @openLiveChat.
  ///
  /// In en, this message translates to:
  /// **'Open live chat'**
  String get openLiveChat;

  /// No description provided for @ticketTitle.
  ///
  /// In en, this message translates to:
  /// **'Ticket'**
  String get ticketTitle;

  /// No description provided for @allFilter.
  ///
  /// In en, this message translates to:
  /// **'All'**
  String get allFilter;

  /// No description provided for @topUpWallet.
  ///
  /// In en, this message translates to:
  /// **'Top up wallet'**
  String get topUpWallet;

  /// No description provided for @amountTry.
  ///
  /// In en, this message translates to:
  /// **'Amount (₺)'**
  String get amountTry;

  /// No description provided for @topUp.
  ///
  /// In en, this message translates to:
  /// **'Top up'**
  String get topUp;

  /// No description provided for @transactionHistory.
  ///
  /// In en, this message translates to:
  /// **'Transaction history'**
  String get transactionHistory;

  /// No description provided for @availableBalance.
  ///
  /// In en, this message translates to:
  /// **'Available balance'**
  String get availableBalance;

  /// No description provided for @pending.
  ///
  /// In en, this message translates to:
  /// **'Pending'**
  String get pending;

  /// No description provided for @verificationCode.
  ///
  /// In en, this message translates to:
  /// **'Verification code'**
  String get verificationCode;

  /// No description provided for @verifyEmail.
  ///
  /// In en, this message translates to:
  /// **'Verify email'**
  String get verifyEmail;

  /// No description provided for @skipForNow.
  ///
  /// In en, this message translates to:
  /// **'Skip for now'**
  String get skipForNow;

  /// No description provided for @forgotPassword.
  ///
  /// In en, this message translates to:
  /// **'Forgot password?'**
  String get forgotPassword;

  /// No description provided for @sendResetLink.
  ///
  /// In en, this message translates to:
  /// **'Send reset link'**
  String get sendResetLink;

  /// No description provided for @alreadyHaveResetToken.
  ///
  /// In en, this message translates to:
  /// **'Already have a reset token?'**
  String get alreadyHaveResetToken;

  /// No description provided for @resetToken.
  ///
  /// In en, this message translates to:
  /// **'Reset token'**
  String get resetToken;

  /// No description provided for @newPassword.
  ///
  /// In en, this message translates to:
  /// **'New password'**
  String get newPassword;

  /// No description provided for @confirmPassword.
  ///
  /// In en, this message translates to:
  /// **'Confirm password'**
  String get confirmPassword;

  /// No description provided for @updatePassword.
  ///
  /// In en, this message translates to:
  /// **'Update password'**
  String get updatePassword;

  /// No description provided for @requestDataExport.
  ///
  /// In en, this message translates to:
  /// **'Request data export'**
  String get requestDataExport;

  /// No description provided for @reasonOptional.
  ///
  /// In en, this message translates to:
  /// **'Reason (optional)'**
  String get reasonOptional;

  /// No description provided for @cannotBeUndone.
  ///
  /// In en, this message translates to:
  /// **'I understand this cannot be undone'**
  String get cannotBeUndone;

  /// No description provided for @useThisLocation.
  ///
  /// In en, this message translates to:
  /// **'Use this location'**
  String get useThisLocation;

  /// No description provided for @somethingWentWrong.
  ///
  /// In en, this message translates to:
  /// **'Something went wrong'**
  String get somethingWentWrong;

  /// No description provided for @increaseQuantity.
  ///
  /// In en, this message translates to:
  /// **'Increase quantity'**
  String get increaseQuantity;

  /// No description provided for @decreaseQuantity.
  ///
  /// In en, this message translates to:
  /// **'Decrease quantity'**
  String get decreaseQuantity;

  /// No description provided for @quantityLabel.
  ///
  /// In en, this message translates to:
  /// **'Quantity {count}'**
  String quantityLabel(int count);

  /// No description provided for @walletAmountOptional.
  ///
  /// In en, this message translates to:
  /// **'Wallet amount (optional)'**
  String get walletAmountOptional;

  /// No description provided for @giftCard.
  ///
  /// In en, this message translates to:
  /// **'Gift card'**
  String get giftCard;

  /// No description provided for @nameLabel.
  ///
  /// In en, this message translates to:
  /// **'Name'**
  String get nameLabel;

  /// No description provided for @passwordLabel.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get passwordLabel;

  /// No description provided for @openSourceLicenses.
  ///
  /// In en, this message translates to:
  /// **'Open source licenses'**
  String get openSourceLicenses;

  /// No description provided for @companyName.
  ///
  /// In en, this message translates to:
  /// **'Company name'**
  String get companyName;

  /// No description provided for @taxId.
  ///
  /// In en, this message translates to:
  /// **'Tax ID'**
  String get taxId;

  /// No description provided for @taxOffice.
  ///
  /// In en, this message translates to:
  /// **'Tax office'**
  String get taxOffice;

  /// No description provided for @giftMessage.
  ///
  /// In en, this message translates to:
  /// **'Gift message'**
  String get giftMessage;

  /// No description provided for @supportAssistant.
  ///
  /// In en, this message translates to:
  /// **'Support assistant'**
  String get supportAssistant;

  /// No description provided for @favoriteProduct.
  ///
  /// In en, this message translates to:
  /// **'Products'**
  String get favoriteProduct;

  /// No description provided for @favoriteStore.
  ///
  /// In en, this message translates to:
  /// **'Stores'**
  String get favoriteStore;

  /// No description provided for @favoriteBrand.
  ///
  /// In en, this message translates to:
  /// **'Brands'**
  String get favoriteBrand;

  /// No description provided for @favoriteCategory.
  ///
  /// In en, this message translates to:
  /// **'Categories'**
  String get favoriteCategory;

  /// No description provided for @favoriteSearch.
  ///
  /// In en, this message translates to:
  /// **'Searches'**
  String get favoriteSearch;

  /// No description provided for @estimate.
  ///
  /// In en, this message translates to:
  /// **'Estimate'**
  String get estimate;

  /// No description provided for @couponLabel.
  ///
  /// In en, this message translates to:
  /// **'Coupon'**
  String get couponLabel;

  /// No description provided for @smartReorder.
  ///
  /// In en, this message translates to:
  /// **'Smart reorder'**
  String get smartReorder;

  /// No description provided for @budgetOptimize.
  ///
  /// In en, this message translates to:
  /// **'Optimize budget'**
  String get budgetOptimize;

  /// No description provided for @fixCartBeforeCheckout.
  ///
  /// In en, this message translates to:
  /// **'Fix cart issues before checkout'**
  String get fixCartBeforeCheckout;

  /// No description provided for @walletApplied.
  ///
  /// In en, this message translates to:
  /// **'Wallet applied'**
  String get walletApplied;

  /// No description provided for @loyaltyPointsApplied.
  ///
  /// In en, this message translates to:
  /// **'Loyalty points applied'**
  String get loyaltyPointsApplied;

  /// No description provided for @itemUnavailableTitle.
  ///
  /// In en, this message translates to:
  /// **'If an item is unavailable'**
  String get itemUnavailableTitle;

  /// No description provided for @substitutionPreference.
  ///
  /// In en, this message translates to:
  /// **'Substitution preference'**
  String get substitutionPreference;

  /// No description provided for @allowSubstitutions.
  ///
  /// In en, this message translates to:
  /// **'Allow substitutions'**
  String get allowSubstitutions;

  /// No description provided for @allowSubstitutionsHint.
  ///
  /// In en, this message translates to:
  /// **'Replace with a similar item automatically'**
  String get allowSubstitutionsHint;

  /// No description provided for @contactMe.
  ///
  /// In en, this message translates to:
  /// **'Contact me'**
  String get contactMe;

  /// No description provided for @contactMeHint.
  ///
  /// In en, this message translates to:
  /// **'Call before replacing anything'**
  String get contactMeHint;

  /// No description provided for @doNotSubstitute.
  ///
  /// In en, this message translates to:
  /// **'Do not substitute'**
  String get doNotSubstitute;

  /// No description provided for @doNotSubstituteHint.
  ///
  /// In en, this message translates to:
  /// **'Skip or refund unavailable items'**
  String get doNotSubstituteHint;

  /// No description provided for @outOfStockRule.
  ///
  /// In en, this message translates to:
  /// **'Out of stock rule'**
  String get outOfStockRule;

  /// No description provided for @replaceWithSimilar.
  ///
  /// In en, this message translates to:
  /// **'Replace with similar'**
  String get replaceWithSimilar;

  /// No description provided for @replaceWithSimilarHint.
  ///
  /// In en, this message translates to:
  /// **'Find a close match when possible'**
  String get replaceWithSimilarHint;

  /// No description provided for @refundItem.
  ///
  /// In en, this message translates to:
  /// **'Refund item'**
  String get refundItem;

  /// No description provided for @refundItemHint.
  ///
  /// In en, this message translates to:
  /// **'Remove the item and refund its price'**
  String get refundItemHint;

  /// No description provided for @cancelIfMissing.
  ///
  /// In en, this message translates to:
  /// **'Cancel the whole order if anything is missing'**
  String get cancelIfMissing;

  /// No description provided for @allergens.
  ///
  /// In en, this message translates to:
  /// **'Allergens'**
  String get allergens;

  /// No description provided for @privacySettingsSaved.
  ///
  /// In en, this message translates to:
  /// **'Privacy settings saved'**
  String get privacySettingsSaved;

  /// No description provided for @giftCardApplied.
  ///
  /// In en, this message translates to:
  /// **'Gift card applied'**
  String get giftCardApplied;

  /// No description provided for @createAccount.
  ///
  /// In en, this message translates to:
  /// **'Create account'**
  String get createAccount;

  /// No description provided for @enterVerificationCode.
  ///
  /// In en, this message translates to:
  /// **'Enter the verification code'**
  String get enterVerificationCode;

  /// No description provided for @emailVerified.
  ///
  /// In en, this message translates to:
  /// **'Email verified'**
  String get emailVerified;

  /// No description provided for @enterCodeFromEmail.
  ///
  /// In en, this message translates to:
  /// **'Enter the verification code from your email.'**
  String get enterCodeFromEmail;

  /// No description provided for @enterCodeSentTo.
  ///
  /// In en, this message translates to:
  /// **'Enter the code we sent to {email}.'**
  String enterCodeSentTo(String email);

  /// No description provided for @forgotPasswordHint.
  ///
  /// In en, this message translates to:
  /// **'Enter your email and we will send you a reset link.'**
  String get forgotPasswordHint;

  /// No description provided for @passwordResetEmailSent.
  ///
  /// In en, this message translates to:
  /// **'Password reset email sent'**
  String get passwordResetEmailSent;

  /// No description provided for @failedToSendResetEmail.
  ///
  /// In en, this message translates to:
  /// **'Could not send a reset email. Please try again.'**
  String get failedToSendResetEmail;

  /// No description provided for @deleteAccountPermanent.
  ///
  /// In en, this message translates to:
  /// **'This action is permanent. Your orders and personal data will be scheduled for deletion.'**
  String get deleteAccountPermanent;

  /// No description provided for @confirmAccountDeletion.
  ///
  /// In en, this message translates to:
  /// **'Please confirm account deletion'**
  String get confirmAccountDeletion;

  /// No description provided for @failedToDeleteAccount.
  ///
  /// In en, this message translates to:
  /// **'Could not delete the account. Please try again.'**
  String get failedToDeleteAccount;

  /// No description provided for @dataExportRequested.
  ///
  /// In en, this message translates to:
  /// **'Data export requested'**
  String get dataExportRequested;

  /// No description provided for @failedToRequestExport.
  ///
  /// In en, this message translates to:
  /// **'Could not request a data export. Please try again.'**
  String get failedToRequestExport;

  /// No description provided for @openReviewFromOrder.
  ///
  /// In en, this message translates to:
  /// **'Open this screen from an order to submit a review'**
  String get openReviewFromOrder;

  /// No description provided for @rateYourOrder.
  ///
  /// In en, this message translates to:
  /// **'Rate your order'**
  String get rateYourOrder;

  /// No description provided for @rateCourier.
  ///
  /// In en, this message translates to:
  /// **'Rate courier'**
  String get rateCourier;

  /// No description provided for @frequentlyAskedQuestions.
  ///
  /// In en, this message translates to:
  /// **'Frequently asked questions'**
  String get frequentlyAskedQuestions;

  /// No description provided for @helpFaqPlaceQ.
  ///
  /// In en, this message translates to:
  /// **'How do I place an order?'**
  String get helpFaqPlaceQ;

  /// No description provided for @helpFaqPlaceA.
  ///
  /// In en, this message translates to:
  /// **'Browse categories or search for products, add items to your cart, then checkout with a delivery address and payment method.'**
  String get helpFaqPlaceA;

  /// No description provided for @helpFaqEtaQ.
  ///
  /// In en, this message translates to:
  /// **'What are delivery times?'**
  String get helpFaqEtaQ;

  /// No description provided for @helpFaqEtaA.
  ///
  /// In en, this message translates to:
  /// **'ETA is shown on product and checkout screens based on your city and selected address. Most orders arrive within the estimated window.'**
  String get helpFaqEtaA;

  /// No description provided for @helpFaqTrackQ.
  ///
  /// In en, this message translates to:
  /// **'How do I track my order?'**
  String get helpFaqTrackQ;

  /// No description provided for @helpFaqTrackA.
  ///
  /// In en, this message translates to:
  /// **'Open Orders from your account, select the order, then tap Track. You will see live status updates until delivery.'**
  String get helpFaqTrackA;

  /// No description provided for @helpFaqRefundQ.
  ///
  /// In en, this message translates to:
  /// **'How do refunds work?'**
  String get helpFaqRefundQ;

  /// No description provided for @helpFaqRefundA.
  ///
  /// In en, this message translates to:
  /// **'If an item is missing or damaged, open the order and contact Support. Approved refunds return to your original payment method or wallet.'**
  String get helpFaqRefundA;

  /// No description provided for @helpFaqScheduleQ.
  ///
  /// In en, this message translates to:
  /// **'Can I schedule a delivery?'**
  String get helpFaqScheduleQ;

  /// No description provided for @helpFaqScheduleA.
  ///
  /// In en, this message translates to:
  /// **'Yes. During checkout choose Schedule delivery and pick an available slot for your address.'**
  String get helpFaqScheduleA;

  /// No description provided for @helpFaqCouponQ.
  ///
  /// In en, this message translates to:
  /// **'How do coupons and loyalty points work?'**
  String get helpFaqCouponQ;

  /// No description provided for @helpFaqCouponA.
  ///
  /// In en, this message translates to:
  /// **'Apply a coupon on checkout. Loyalty points accumulate on eligible orders and can be redeemed where shown in your wallet or loyalty screen.'**
  String get helpFaqCouponA;

  /// No description provided for @contactBeforeSubstituting.
  ///
  /// In en, this message translates to:
  /// **'Contact before substituting'**
  String get contactBeforeSubstituting;

  /// No description provided for @pinYourAddress.
  ///
  /// In en, this message translates to:
  /// **'Pin your address'**
  String get pinYourAddress;

  /// No description provided for @rateProducts.
  ///
  /// In en, this message translates to:
  /// **'Rate products'**
  String get rateProducts;

  /// No description provided for @verifiedPurchase.
  ///
  /// In en, this message translates to:
  /// **'Verified purchase'**
  String get verifiedPurchase;

  /// No description provided for @resetPassword.
  ///
  /// In en, this message translates to:
  /// **'Reset password'**
  String get resetPassword;

  /// No description provided for @resetTokenRequired.
  ///
  /// In en, this message translates to:
  /// **'Reset token is required'**
  String get resetTokenRequired;

  /// No description provided for @passwordsDoNotMatch.
  ///
  /// In en, this message translates to:
  /// **'Passwords do not match'**
  String get passwordsDoNotMatch;

  /// No description provided for @passwordUpdated.
  ///
  /// In en, this message translates to:
  /// **'Password updated'**
  String get passwordUpdated;

  /// No description provided for @failedToResetPassword.
  ///
  /// In en, this message translates to:
  /// **'Could not reset the password. Please try again.'**
  String get failedToResetPassword;

  /// No description provided for @anyAvailability.
  ///
  /// In en, this message translates to:
  /// **'Any'**
  String get anyAvailability;

  /// No description provided for @lowStock.
  ///
  /// In en, this message translates to:
  /// **'Low stock'**
  String get lowStock;

  /// No description provided for @register.
  ///
  /// In en, this message translates to:
  /// **'Register'**
  String get register;

  /// No description provided for @alreadyHaveAccount.
  ///
  /// In en, this message translates to:
  /// **'Already have an account? Sign in'**
  String get alreadyHaveAccount;

  /// No description provided for @needAnAccount.
  ///
  /// In en, this message translates to:
  /// **'Need an account? Register'**
  String get needAnAccount;

  /// No description provided for @cashback.
  ///
  /// In en, this message translates to:
  /// **'Cashback'**
  String get cashback;

  /// No description provided for @promoCredit.
  ///
  /// In en, this message translates to:
  /// **'Promo'**
  String get promoCredit;

  /// No description provided for @walletPayHint.
  ///
  /// In en, this message translates to:
  /// **'Pay from your wallet balance'**
  String get walletPayHint;

  /// No description provided for @cashPayHint.
  ///
  /// In en, this message translates to:
  /// **'Pay the courier when your order arrives'**
  String get cashPayHint;

  /// No description provided for @giftCardRedeemHint.
  ///
  /// In en, this message translates to:
  /// **'Redeem a gift card balance'**
  String get giftCardRedeemHint;

  /// No description provided for @previousPaymentFailed.
  ///
  /// In en, this message translates to:
  /// **'Previous payment failed. You can retry.'**
  String get previousPaymentFailed;

  /// No description provided for @reviewYourOrder.
  ///
  /// In en, this message translates to:
  /// **'Review your order'**
  String get reviewYourOrder;

  /// No description provided for @contactlessHint.
  ///
  /// In en, this message translates to:
  /// **'Leave at the door — no handoff needed.'**
  String get contactlessHint;

  /// No description provided for @giftOrderHint.
  ///
  /// In en, this message translates to:
  /// **'Hide prices on the receipt for the recipient.'**
  String get giftOrderHint;

  /// No description provided for @companyInvoice.
  ///
  /// In en, this message translates to:
  /// **'Company invoice'**
  String get companyInvoice;

  /// No description provided for @companyInvoiceHint.
  ///
  /// In en, this message translates to:
  /// **'Request a corporate invoice for this order.'**
  String get companyInvoiceHint;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'tr'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'tr':
      return AppLocalizationsTr();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
