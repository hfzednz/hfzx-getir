// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'NEXORA Warehouse';

  @override
  String get homeTitle => 'Dashboard';

  @override
  String get pickingTitle => 'Picking';

  @override
  String get packingTitle => 'Packing';

  @override
  String get dispatchTitle => 'Dispatch';

  @override
  String get inventoryTitle => 'Inventory';

  @override
  String get moreTitle => 'More';

  @override
  String get claimTask => 'Claim';

  @override
  String get startPick => 'Start picking';

  @override
  String get scanBarcode => 'Scan barcode';

  @override
  String get confirmPack => 'Confirm pack';

  @override
  String get handoff => 'Handoff';

  @override
  String get cycleCount => 'Cycle count';

  @override
  String get lowStock => 'Low stock';

  @override
  String get outOfStock => 'Out of stock';

  @override
  String get expired => 'Expired';

  @override
  String get clockIn => 'Clock in';

  @override
  String get clockOut => 'Clock out';

  @override
  String get offlineMessage => 'Offline — scans will sync when connected';

  @override
  String get retry => 'Retry';

  @override
  String get continueLabel => 'Continue';

  @override
  String get phoneLogin => 'Phone number';

  @override
  String get sendOtp => 'Send code';

  @override
  String get verifyOtp => 'Verify';

  @override
  String get otpHint => '6-digit code';

  @override
  String get emptyTitle => 'Queue empty';

  @override
  String get emptyMessage => 'No tasks right now.';

  @override
  String get weightCheck => 'Weight check';

  @override
  String get printLabel => 'Print label';

  @override
  String get exception => 'Exception';

  @override
  String get settings => 'Settings';

  @override
  String get support => 'Support';

  @override
  String get reports => 'Reports';

  @override
  String get transfers => 'Transfers';

  @override
  String get returns => 'Returns';

  @override
  String get quality => 'Quality';

  @override
  String get purchasing => 'Purchasing';

  @override
  String get warehouseMap => 'Warehouse map';

  @override
  String get aiAssist => 'AI assist';
}
