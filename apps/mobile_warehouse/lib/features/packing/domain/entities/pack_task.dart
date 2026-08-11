import 'package:equatable/equatable.dart';

enum PackTaskStatus {
  readyToPack,
  packing,
  weighed,
  labeled,
  packed,
  qcHold,
  dispatchQueued;

  static PackTaskStatus fromString(String? raw) {
    final v = (raw ?? '').toLowerCase().replaceAll('-', '_');
    return switch (v) {
      'ready_to_pack' || 'readytopack' => PackTaskStatus.readyToPack,
      'packing' => PackTaskStatus.packing,
      'weighed' => PackTaskStatus.weighed,
      'labeled' => PackTaskStatus.labeled,
      'packed' => PackTaskStatus.packed,
      'qc_hold' || 'qchold' => PackTaskStatus.qcHold,
      'dispatch_queued' || 'dispatchqueued' => PackTaskStatus.dispatchQueued,
      _ => PackTaskStatus.readyToPack,
    };
  }

  String get wireName => switch (this) {
        PackTaskStatus.readyToPack => 'ready_to_pack',
        PackTaskStatus.qcHold => 'qc_hold',
        PackTaskStatus.dispatchQueued => 'dispatch_queued',
        _ => name,
      };
}

class PackMaterial extends Equatable {
  const PackMaterial({
    required this.code,
    required this.name,
    this.qty = 1,
  });

  final String code;
  final String name;
  final int qty;

  factory PackMaterial.fromJson(Map<String, dynamic> json) {
    return PackMaterial(
      code: json['code']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      qty: (json['qty'] as num?)?.toInt() ?? 1,
    );
  }

  @override
  List<Object?> get props => [code, name, qty];
}

class PackTask extends Equatable {
  const PackTask({
    required this.id,
    required this.orderId,
    required this.status,
    this.expectedWeightGrams = 0,
    this.actualWeightGrams,
    this.materials = const [],
    this.labelUrl,
    this.sealed = false,
    this.labelPrinted = false,
    this.itemCount = 0,
  });

  final String id;
  final String orderId;
  final PackTaskStatus status;
  final double expectedWeightGrams;
  final double? actualWeightGrams;
  final List<PackMaterial> materials;
  final String? labelUrl;
  final bool sealed;
  final bool labelPrinted;
  final int itemCount;

  factory PackTask.fromJson(Map<String, dynamic> json) {
    final mats = json['materials'] as List? ?? const [];
    return PackTask(
      id: json['id']?.toString() ?? '',
      orderId: json['order_id']?.toString() ?? '',
      status: PackTaskStatus.fromString(json['status']?.toString()),
      expectedWeightGrams:
          (json['expected_weight_grams'] as num?)?.toDouble() ?? 0,
      actualWeightGrams: (json['actual_weight_grams'] as num?)?.toDouble(),
      materials: mats
          .map((e) => PackMaterial.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
      labelUrl: json['label_url']?.toString(),
      sealed: json['sealed'] as bool? ?? false,
      labelPrinted: json['label_printed'] as bool? ?? false,
      itemCount: (json['item_count'] as num?)?.toInt() ?? 0,
    );
  }

  @override
  List<Object?> get props =>
      [id, orderId, status, expectedWeightGrams, actualWeightGrams, sealed];
}
