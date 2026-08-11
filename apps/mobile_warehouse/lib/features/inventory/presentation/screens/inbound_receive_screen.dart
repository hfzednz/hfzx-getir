import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../providers/inventory_providers.dart';

/// Stub inbound receive screen (putaway wiring later).
class InboundReceiveScreen extends ConsumerStatefulWidget {
  const InboundReceiveScreen({super.key});

  @override
  ConsumerState<InboundReceiveScreen> createState() => _InboundReceiveScreenState();
}

class _InboundReceiveScreenState extends ConsumerState<InboundReceiveScreen> {
  final _refCtrl = TextEditingController();
  final _skuCtrl = TextEditingController();
  final _qtyCtrl = TextEditingController();

  @override
  void dispose() {
    _refCtrl.dispose();
    _skuCtrl.dispose();
    _qtyCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Inbound receive'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s3),
        children: [
          TextField(controller: _refCtrl, decoration: const InputDecoration(labelText: 'ASN / PO ref')),
          TextField(controller: _skuCtrl, decoration: const InputDecoration(labelText: 'SKU')),
          TextField(controller: _qtyCtrl, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: 'Qty')),
          const SizedBox(height: NxSpacing.s3),
          NxButton(
            label: 'Receive line (stub)',
            expand: true,
            onPressed: () async {
              final qty = int.tryParse(_qtyCtrl.text) ?? 0;
              final r = await ref.read(inventoryActionsProvider).receiveInbound(
                    reference: _refCtrl.text.trim(),
                    lines: [
                      {'sku': _skuCtrl.text.trim(), 'qty': qty},
                    ],
                  );
              if (!context.mounted) return;
              r.fold(
                onSuccess: (_) => ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Receive submitted')),
                ),
                onFailure: (e) => ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text(e.message)),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}
