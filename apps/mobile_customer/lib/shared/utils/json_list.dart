/// Unwraps `{items|hits|data: [...]}` or a raw JSON array into maps.
List<Map<String, dynamic>> jsonObjectList(dynamic json) {
  final raw = json is Map
      ? (json['items'] ??
          json['Items'] ??
          json['hits'] ??
          json['Hits'] ??
          json['products'] ??
          json['Products'] ??
          json['data'] ??
          json['orders'] ??
          json['addresses'] ??
          json['categories'] ??
          json['stores'] ??
          json['favorites'] ??
          json['notifications'] ??
          json['tickets'] ??
          json['faq'] ??
          json['coupons'])
      : json;
  if (raw is! List) return const [];
  return [
    for (final item in raw)
      if (item is Map) Map<String, dynamic>.from(item),
  ];
}
