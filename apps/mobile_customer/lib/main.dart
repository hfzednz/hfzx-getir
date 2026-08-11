import 'bootstrap/bootstrap.dart';

Future<void> main() async {
  final bootstrapResult = await bootstrap();
  await runNexoraApp(bootstrapResult);
}
