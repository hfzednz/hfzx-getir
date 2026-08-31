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
