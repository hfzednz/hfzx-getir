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
  /// **'NEXORA Courier'**
  String get appTitle;

  /// No description provided for @homeTitle.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get homeTitle;

  /// No description provided for @offersTitle.
  ///
  /// In en, this message translates to:
  /// **'Offers'**
  String get offersTitle;

  /// No description provided for @deliveriesTitle.
  ///
  /// In en, this message translates to:
  /// **'Deliveries'**
  String get deliveriesTitle;

  /// No description provided for @earningsTitle.
  ///
  /// In en, this message translates to:
  /// **'Earnings'**
  String get earningsTitle;

  /// No description provided for @accountTitle.
  ///
  /// In en, this message translates to:
  /// **'Account'**
  String get accountTitle;

  /// No description provided for @goOnline.
  ///
  /// In en, this message translates to:
  /// **'Go online'**
  String get goOnline;

  /// No description provided for @goOffline.
  ///
  /// In en, this message translates to:
  /// **'Go offline'**
  String get goOffline;

  /// No description provided for @busyMode.
  ///
  /// In en, this message translates to:
  /// **'Busy'**
  String get busyMode;

  /// No description provided for @breakMode.
  ///
  /// In en, this message translates to:
  /// **'On break'**
  String get breakMode;

  /// No description provided for @emergencyMode.
  ///
  /// In en, this message translates to:
  /// **'Emergency'**
  String get emergencyMode;

  /// No description provided for @acceptOffer.
  ///
  /// In en, this message translates to:
  /// **'Accept'**
  String get acceptOffer;

  /// No description provided for @rejectOffer.
  ///
  /// In en, this message translates to:
  /// **'Reject'**
  String get rejectOffer;

  /// No description provided for @navigate.
  ///
  /// In en, this message translates to:
  /// **'Navigate'**
  String get navigate;

  /// No description provided for @confirmPickup.
  ///
  /// In en, this message translates to:
  /// **'Confirm pickup'**
  String get confirmPickup;

  /// No description provided for @confirmDelivery.
  ///
  /// In en, this message translates to:
  /// **'Confirm delivery'**
  String get confirmDelivery;

  /// No description provided for @proofOfDelivery.
  ///
  /// In en, this message translates to:
  /// **'Proof of delivery'**
  String get proofOfDelivery;

  /// No description provided for @scanQr.
  ///
  /// In en, this message translates to:
  /// **'Scan QR'**
  String get scanQr;

  /// No description provided for @startShift.
  ///
  /// In en, this message translates to:
  /// **'Start shift'**
  String get startShift;

  /// No description provided for @endShift.
  ///
  /// In en, this message translates to:
  /// **'End shift'**
  String get endShift;

  /// No description provided for @todayEarnings.
  ///
  /// In en, this message translates to:
  /// **'Today\'s earnings'**
  String get todayEarnings;

  /// No description provided for @performance.
  ///
  /// In en, this message translates to:
  /// **'Performance'**
  String get performance;

  /// No description provided for @support.
  ///
  /// In en, this message translates to:
  /// **'Support'**
  String get support;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @notifications.
  ///
  /// In en, this message translates to:
  /// **'Notifications'**
  String get notifications;

  /// No description provided for @offlineMessage.
  ///
  /// In en, this message translates to:
  /// **'You are offline — actions will sync when connected'**
  String get offlineMessage;

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

  /// No description provided for @phoneLogin.
  ///
  /// In en, this message translates to:
  /// **'Phone number'**
  String get phoneLogin;

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

  /// No description provided for @otpHint.
  ///
  /// In en, this message translates to:
  /// **'6-digit code'**
  String get otpHint;

  /// No description provided for @kycTitle.
  ///
  /// In en, this message translates to:
  /// **'Identity verification'**
  String get kycTitle;

  /// No description provided for @emptyTitle.
  ///
  /// In en, this message translates to:
  /// **'Nothing here yet'**
  String get emptyTitle;

  /// No description provided for @emptyMessage.
  ///
  /// In en, this message translates to:
  /// **'Check back soon.'**
  String get emptyMessage;

  /// No description provided for @failedDelivery.
  ///
  /// In en, this message translates to:
  /// **'Mark as failed'**
  String get failedDelivery;

  /// No description provided for @sos.
  ///
  /// In en, this message translates to:
  /// **'SOS'**
  String get sos;

  /// No description provided for @batteryLow.
  ///
  /// In en, this message translates to:
  /// **'Low battery — location updates slowed'**
  String get batteryLow;

  /// No description provided for @connectionPoor.
  ///
  /// In en, this message translates to:
  /// **'Poor connection'**
  String get connectionPoor;
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
