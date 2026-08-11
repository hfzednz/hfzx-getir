import 'package:nexora_core/nexora_core.dart';
import '../entities/warehouse_layout.dart';

abstract class MapRepository {
  Future<Result<WarehouseLayout>> fetchLayout();
}
