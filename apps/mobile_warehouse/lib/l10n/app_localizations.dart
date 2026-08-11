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
  /// **'NEXORA Warehouse'**
  String get appTitle;

  /// No description provided for @homeTitle.
  ///
  /// In en, this message translates to:
  /// **'Dashboard'**
  String get homeTitle;

  /// No description provided for @pickingTitle.
  ///
  /// In en, this message translates to:
  /// **'Picking'**
  String get pickingTitle;

  /// No description provided for @packingTitle.
  ///
  /// In en, this message translates to:
  /// **'Packing'**
  String get packingTitle;

  /// No description provided for @dispatchTitle.
  ///
  /// In en, this message translates to:
  /// **'Dispatch'**
  String get dispatchTitle;

  /// No description provided for @inventoryTitle.
  ///
  /// In en, this message translates to:
  /// **'Inventory'**
  String get inventoryTitle;

  /// No description provided for @moreTitle.
  ///
  /// In en, this message translates to:
  /// **'More'**
  String get moreTitle;

  /// No description provided for @claimTask.
  ///
  /// In en, this message translates to:
  /// **'Claim'**
  String get claimTask;

  /// No description provided for @startPick.
  ///
  /// In en, this message translates to:
  /// **'Start picking'**
  String get startPick;

  /// No description provided for @scanBarcode.
  ///
  /// In en, this message translates to:
  /// **'Scan barcode'**
  String get scanBarcode;

  /// No description provided for @confirmPack.
  ///
  /// In en, this message translates to:
  /// **'Confirm pack'**
  String get confirmPack;

  /// No description provided for @handoff.
  ///
  /// In en, this message translates to:
  /// **'Handoff'**
  String get handoff;

  /// No description provided for @cycleCount.
  ///
  /// In en, this message translates to:
  /// **'Cycle count'**
  String get cycleCount;

  /// No description provided for @lowStock.
  ///
  /// In en, this message translates to:
  /// **'Low stock'**
  String get lowStock;

  /// No description provided for @outOfStock.
  ///
  /// In en, this message translates to:
  /// **'Out of stock'**
  String get outOfStock;

  /// No description provided for @expired.
  ///
  /// In en, this message translates to:
  /// **'Expired'**
  String get expired;

  /// No description provided for @clockIn.
  ///
  /// In en, this message translates to:
  /// **'Clock in'**
  String get clockIn;

  /// No description provided for @clockOut.
  ///
  /// In en, this message translates to:
  /// **'Clock out'**
  String get clockOut;

  /// No description provided for @offlineMessage.
  ///
  /// In en, this message translates to:
  /// **'Offline — scans will sync when connected'**
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

  /// No description provided for @emptyTitle.
  ///
  /// In en, this message translates to:
  /// **'Queue empty'**
  String get emptyTitle;

  /// No description provided for @emptyMessage.
  ///
  /// In en, this message translates to:
  /// **'No tasks right now.'**
  String get emptyMessage;

  /// No description provided for @weightCheck.
  ///
  /// In en, this message translates to:
  /// **'Weight check'**
  String get weightCheck;

  /// No description provided for @printLabel.
  ///
  /// In en, this message translates to:
  /// **'Print label'**
  String get printLabel;

  /// No description provided for @exception.
  ///
  /// In en, this message translates to:
  /// **'Exception'**
  String get exception;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @support.
  ///
  /// In en, this message translates to:
  /// **'Support'**
  String get support;

  /// No description provided for @reports.
  ///
  /// In en, this message translates to:
  /// **'Reports'**
  String get reports;

  /// No description provided for @transfers.
  ///
  /// In en, this message translates to:
  /// **'Transfers'**
  String get transfers;

  /// No description provided for @returns.
  ///
  /// In en, this message translates to:
  /// **'Returns'**
  String get returns;

  /// No description provided for @quality.
  ///
  /// In en, this message translates to:
  /// **'Quality'**
  String get quality;

  /// No description provided for @purchasing.
  ///
  /// In en, this message translates to:
  /// **'Purchasing'**
  String get purchasing;

  /// No description provided for @warehouseMap.
  ///
  /// In en, this message translates to:
  /// **'Warehouse map'**
  String get warehouseMap;

  /// No description provided for @aiAssist.
  ///
  /// In en, this message translates to:
  /// **'AI assist'**
  String get aiAssist;
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
