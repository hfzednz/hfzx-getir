import '../../domain/entities/checkout_entity.dart';

class CheckoutSessionModel {
  const CheckoutSessionModel({required this.entity});

  final CheckoutSession entity;

  factory CheckoutSessionModel.fromJson(Map<String, dynamic> json) =>
      CheckoutSessionModel(entity: CheckoutSession.fromJson(json));

  CheckoutSession toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}

class CheckoutQuoteModel {
  const CheckoutQuoteModel({required this.entity});

  final CheckoutQuote entity;

  factory CheckoutQuoteModel.fromJson(Map<String, dynamic> json) =>
      CheckoutQuoteModel(entity: CheckoutQuote.fromJson(json));

  CheckoutQuote toEntity() => entity;
}

class SavedPaymentCardModel {
  const SavedPaymentCardModel({required this.entity});

  final SavedPaymentCard entity;

  factory SavedPaymentCardModel.fromJson(Map<String, dynamic> json) =>
      SavedPaymentCardModel(entity: SavedPaymentCard.fromJson(json));

  SavedPaymentCard toEntity() => entity;
}
