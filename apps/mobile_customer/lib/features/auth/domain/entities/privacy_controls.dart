import 'package:equatable/equatable.dart';

class PrivacyControls extends Equatable {
  const PrivacyControls({
    this.marketingEmail = false,
    this.marketingPush = false,
    this.marketingSms = false,
    this.personalization = true,
    this.analyticsOptOut = false,
    this.shareWithPartners = false,
  });

  final bool marketingEmail;
  final bool marketingPush;
  final bool marketingSms;
  final bool personalization;
  final bool analyticsOptOut;
  final bool shareWithPartners;

  factory PrivacyControls.fromJson(Map<String, dynamic> json) =>
      PrivacyControls(
        marketingEmail: json['marketing_email'] == true,
        marketingPush: json['marketing_push'] == true,
        marketingSms: json['marketing_sms'] == true,
        personalization: json['personalization'] != false,
        analyticsOptOut: json['analytics_opt_out'] == true,
        shareWithPartners: json['share_with_partners'] == true,
      );

  Map<String, dynamic> toJson() => {
        'marketing_email': marketingEmail,
        'marketing_push': marketingPush,
        'marketing_sms': marketingSms,
        'personalization': personalization,
        'analytics_opt_out': analyticsOptOut,
        'share_with_partners': shareWithPartners,
      };

  PrivacyControls copyWith({
    bool? marketingEmail,
    bool? marketingPush,
    bool? marketingSms,
    bool? personalization,
    bool? analyticsOptOut,
    bool? shareWithPartners,
  }) {
    return PrivacyControls(
      marketingEmail: marketingEmail ?? this.marketingEmail,
      marketingPush: marketingPush ?? this.marketingPush,
      marketingSms: marketingSms ?? this.marketingSms,
      personalization: personalization ?? this.personalization,
      analyticsOptOut: analyticsOptOut ?? this.analyticsOptOut,
      shareWithPartners: shareWithPartners ?? this.shareWithPartners,
    );
  }

  @override
  List<Object?> get props => [
        marketingEmail,
        marketingPush,
        marketingSms,
        personalization,
        analyticsOptOut,
        shareWithPartners,
      ];
}
