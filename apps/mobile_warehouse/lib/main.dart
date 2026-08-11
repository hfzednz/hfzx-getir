import 'bootstrap/bootstrap.dart';

Future<void> main() async {
  final result = await bootstrap();
  await runNexoraWarehouseApp(result);
}
