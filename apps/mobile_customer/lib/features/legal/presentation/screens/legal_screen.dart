import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/legal_entity.dart';
import '../providers/legal_providers.dart';

class LegalScreen extends ConsumerWidget {
  const LegalScreen({super.key, this.doc});

  final String? doc;

  String get _docKey => (doc ?? 'terms').toLowerCase();

  String _title(AppLocalizations l10n) => switch (_docKey) {
        'privacy' => l10n.privacyPolicy,
        'cookies' => 'Cookie policy',
        'terms' => l10n.termsOfService,
        _ => l10n.legalTitle,
      };

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncDoc = ref.watch(legalDocumentProvider(_docKey));

    return Scaffold(
      appBar: NxTopBar(title: _title(l10n)),
      body: AsyncValueWidget(
        value: asyncDoc,
        data: (document) {
          final body = _resolveBody(document);
          final heading =
              document.title.isNotEmpty ? document.title : _title(l10n);
          return RefreshIndicator(
            onRefresh: () async =>
                ref.invalidate(legalDocumentProvider(_docKey)),
            child: SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(heading, style: NxTypography.headlineSm),
                  const SizedBox(height: NxSpacing.s4),
                  Text(body, style: NxTypography.bodyMd),
                ],
              ),
            ),
          );
        },
        error: (e, _) => SingleChildScrollView(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(_title(l10n), style: NxTypography.headlineSm),
              const SizedBox(height: NxSpacing.s3),
              NxButton(
                label: l10n.retry,
                onPressed: () =>
                    ref.invalidate(legalDocumentProvider(_docKey)),
              ),
              const SizedBox(height: NxSpacing.s4),
              Text(_fallbackPolicy(_docKey), style: NxTypography.bodyMd),
            ],
          ),
        ),
      ),
    );
  }

  String _resolveBody(LegalDocument document) {
    final payload = document.payload;
    final content = payload['content']?.toString() ??
        payload['body']?.toString() ??
        payload['html']?.toString() ??
        payload['text']?.toString();
    if (content != null && content.trim().isNotEmpty) {
      return content;
    }
    return _fallbackPolicy(_docKey);
  }
}

String _fallbackPolicy(String docKey) {
  switch (docKey) {
    case 'privacy':
      return '''
NEXORA Privacy Policy

1. Who we are
NEXORA operates a quick-commerce service. This policy explains how we collect, use, and protect personal data when you use our apps and websites.

2. Data we collect
• Account details such as name, phone, email, and delivery addresses
• Order history, payment references (not full card numbers), and support messages
• Device and usage data needed for security, crash reporting, and service improvement
• Location when you enable it to show delivery ETAs and available products

3. How we use data
We use data to fulfill orders, prevent fraud, personalize catalog availability by city, improve the product, and communicate about deliveries or account changes.

4. Sharing
We share data with delivery partners, payment processors, and infrastructure providers only as needed to operate the service. We do not sell personal data.

5. Your choices
You can update privacy controls in Settings, request data access or deletion from account settings, and revoke device sessions at any time.

6. Retention
We keep order and account records for as long as required by law and operational needs, then delete or anonymize them.

7. Contact
For privacy questions, contact Support from the Help screen in the app.
''';
    case 'cookies':
      return '''
NEXORA Cookie Policy

1. What cookies and similar tech we use
We use cookies, local storage, and similar identifiers on web surfaces and SDKs in the mobile app to keep you signed in, remember preferences, measure performance, and improve reliability.

2. Essential
Required for authentication, security, cart continuity, and core checkout flows. These cannot be disabled while using the service.

3. Preferences
Remember language, theme, city, and accessibility choices so the experience stays consistent across sessions.

4. Analytics
Help us understand feature usage and diagnose errors. You can limit non-essential analytics via Privacy controls where available.

5. Managing preferences
Use in-app Privacy controls and your device or browser settings to manage storage and tracking permissions.

6. Updates
We may update this policy when our tooling changes. Continued use after updates means you accept the revised policy.
''';
    default:
      return '''
NEXORA Terms of Service

1. Agreement
By creating an account or placing an order with NEXORA you agree to these Terms of Service.

2. Service
NEXORA offers on-demand retail delivery subject to city availability, stock, and delivery capacity. Product availability and ETAs can change.

3. Accounts
You are responsible for accurate account information and for keeping login methods secure. Do not share OTP codes or biometric access with others.

4. Orders and pricing
Prices, fees, and promotions are shown before you place an order. Once confirmed, orders are binding subject to cancellation rules shown in the app.

5. Delivery
You must provide a reachable address and any access instructions. Failed delivery attempts due to incomplete address details may incur additional fees where disclosed.

6. Acceptable use
Do not misuse the platform, attempt unauthorized access, or place fraudulent orders. We may suspend accounts that violate these terms.

7. Liability
To the extent permitted by law, NEXORA is not liable for indirect damages arising from delivery delays beyond our control or third-party payment outages.

8. Changes
We may update these terms. Material changes will be surfaced in the app. Continued use constitutes acceptance.

9. Contact
Questions about these terms can be sent through in-app Support.
''';
  }
}
