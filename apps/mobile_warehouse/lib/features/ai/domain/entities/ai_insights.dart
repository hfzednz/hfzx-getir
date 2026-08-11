import 'package:equatable/equatable.dart';

class AiHubInsights extends Equatable {
  const AiHubInsights({
    this.demandForecast = const [],
    this.pickPathTip,
    this.restockSuggestions = const [],
  });
  final List<String> demandForecast;
  final String? pickPathTip;
  final List<String> restockSuggestions;
  factory AiHubInsights.fromJson(Map<String, dynamic> json) {
    return AiHubInsights(
      demandForecast: ((json['demand_forecast'] as List?) ?? const []).map((e) => e.toString()).toList(),
      pickPathTip: json['pick_path_tip']?.toString(),
      restockSuggestions: ((json['restock_suggestions'] as List?) ?? const []).map((e) => e.toString()).toList(),
    );
  }
  @override
  List<Object?> get props => [demandForecast, pickPathTip, restockSuggestions];
}
