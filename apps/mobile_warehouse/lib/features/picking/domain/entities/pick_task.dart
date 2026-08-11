import 'package:equatable/equatable.dart';

enum PickTaskStatus {
  queued,
  claimed,
  inProgress,
  shortPick,
  picked,
  staged,
  exception;

  static PickTaskStatus fromString(String? raw) {
    final v = (raw ?? '').toLowerCase().replaceAll('-', '_');
    return switch (v) {
      'queued' => PickTaskStatus.queued,
      'claimed' => PickTaskStatus.claimed,
      'in_progress' || 'inprogress' => PickTaskStatus.inProgress,
      'short_pick' || 'shortpick' => PickTaskStatus.shortPick,
      'picked' => PickTaskStatus.picked,
      'staged' => PickTaskStatus.staged,
      'exception' => PickTaskStatus.exception,
      _ => PickTaskStatus.queued,
    };
  }

  String get wireName => switch (this) {
        PickTaskStatus.inProgress => 'in_progress',
        PickTaskStatus.shortPick => 'short_pick',
        _ => name,
      };
}

class PickLine extends Equatable {
  const PickLine({
    required this.id,
    required this.sku,
    required this.barcode,
    required this.bin,
    required this.qty,
    this.pickedQty = 0,
    this.zone,
    this.pathStep,
    this.shorted = false,
    this.substitutionSku,
    this.name,
  });

  final String id;
  final String sku;
  final String barcode;
  final String bin;
  final int qty;
  final int pickedQty;
  final String? zone;
  final int? pathStep;
  final bool shorted;
  final String? substitutionSku;
  final String? name;

  bool get isComplete => pickedQty >= qty || shorted;

  factory PickLine.fromJson(Map<String, dynamic> json) {
    return PickLine(
      id: json['id']?.toString() ?? '',
      sku: json['sku']?.toString() ?? '',
      barcode: json['barcode']?.toString() ?? '',
      bin: json['bin']?.toString() ?? json['bin_code']?.toString() ?? '',
      qty: (json['qty'] as num?)?.toInt() ?? 0,
      pickedQty: (json['picked_qty'] as num?)?.toInt() ?? 0,
      zone: json['zone']?.toString(),
      pathStep: (json['path_step'] as num?)?.toInt(),
      shorted: json['shorted'] as bool? ?? false,
      substitutionSku: json['substitution_sku']?.toString(),
      name: json['name']?.toString() ?? json['product_name']?.toString(),
    );
  }

  @override
  List<Object?> get props =>
      [id, sku, barcode, bin, qty, pickedQty, zone, pathStep, shorted];
}

class PickTask extends Equatable {
  const PickTask({
    required this.id,
    required this.orderId,
    required this.status,
    required this.lines,
    this.priority = 0,
    this.zone,
    this.claimedBy,
    this.pathHint,
  });

  final String id;
  final String orderId;
  final PickTaskStatus status;
  final List<PickLine> lines;
  final int priority;
  final String? zone;
  final String? claimedBy;
  final String? pathHint;

  List<PickLine> get pathOrderedLines {
    final sorted = [...lines];
    sorted.sort((a, b) {
      final as = a.pathStep ?? 9999;
      final bs = b.pathStep ?? 9999;
      return as.compareTo(bs);
    });
    return sorted;
  }

  factory PickTask.fromJson(Map<String, dynamic> json) {
    final linesJson = json['lines'] as List? ?? const [];
    return PickTask(
      id: json['id']?.toString() ?? '',
      orderId: json['order_id']?.toString() ?? '',
      status: PickTaskStatus.fromString(json['status']?.toString()),
      lines: linesJson
          .map((e) => PickLine.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
      priority: (json['priority'] as num?)?.toInt() ?? 0,
      zone: json['zone']?.toString(),
      claimedBy: json['claimed_by']?.toString(),
      pathHint: json['path_hint']?.toString(),
    );
  }

  @override
  List<Object?> get props =>
      [id, orderId, status, lines, priority, zone, claimedBy];
}
