import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import '../providers/transfers_providers.dart';

class CreateTransferScreen extends ConsumerStatefulWidget {
  const CreateTransferScreen({super.key});
  @override
  ConsumerState<CreateTransferScreen> createState() => _CreateTransferScreenState();
}

class _CreateTransferScreenState extends ConsumerState<CreateTransferScreen> {
  final _from = TextEditingController();
  final _to = TextEditingController();
  final _sku = TextEditingController();
  final _qty = TextEditingController();
  String _type = 'shelf';

  @override
  void dispose() {
    _from.dispose(); _to.dispose(); _sku.dispose(); _qty.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Create transfer'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s3),
        children: [
          DropdownButtonFormField<String>(
            initialValue: _type,
            items: const [
              DropdownMenuItem(value: 'shelf', child: Text('Shelf')),
              DropdownMenuItem(value: 'warehouse', child: Text('Warehouse')),
            ],
            onChanged: (v) => setState(() => _type = v ?? _type),
            decoration: const InputDecoration(labelText: 'Type'),
          ),
          TextField(controller: _from, decoration: const InputDecoration(labelText: 'From')),
          TextField(controller: _to, decoration: const InputDecoration(labelText: 'To')),
          TextField(controller: _sku, decoration: const InputDecoration(labelText: 'SKU')),
          TextField(controller: _qty, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: 'Qty')),
          const SizedBox(height: NxSpacing.s3),
          NxButton(
            label: 'Create',
            expand: true,
            onPressed: () async {
              final r = await ref.read(transfersActionsProvider).create({
                'type': _type,
                'from_location': _from.text.trim(),
                'to_location': _to.text.trim(),
                'sku': _sku.text.trim(),
                'qty': int.tryParse(_qty.text) ?? 0,
              });
              if (!context.mounted) return;
              r.fold(onSuccess: (_) => context.pop(), onFailure: (e) {
                ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
              });
            },
          ),
        ],
      ),
    );
  }
}
