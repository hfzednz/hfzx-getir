// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'app_database.dart';

// ignore_for_file: type=lint
class $CartItemsTable extends CartItems
    with TableInfo<$CartItemsTable, CartItem> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $CartItemsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _productIdMeta = const VerificationMeta(
    'productId',
  );
  @override
  late final GeneratedColumn<String> productId = GeneratedColumn<String>(
    'product_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _variantIdMeta = const VerificationMeta(
    'variantId',
  );
  @override
  late final GeneratedColumn<String> variantId = GeneratedColumn<String>(
    'variant_id',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _titleMeta = const VerificationMeta('title');
  @override
  late final GeneratedColumn<String> title = GeneratedColumn<String>(
    'title',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _imageUrlMeta = const VerificationMeta(
    'imageUrl',
  );
  @override
  late final GeneratedColumn<String> imageUrl = GeneratedColumn<String>(
    'image_url',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _quantityMeta = const VerificationMeta(
    'quantity',
  );
  @override
  late final GeneratedColumn<int> quantity = GeneratedColumn<int>(
    'quantity',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(1),
  );
  static const VerificationMeta _unitPriceMinorMeta = const VerificationMeta(
    'unitPriceMinor',
  );
  @override
  late final GeneratedColumn<int> unitPriceMinor = GeneratedColumn<int>(
    'unit_price_minor',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _currencyMeta = const VerificationMeta(
    'currency',
  );
  @override
  late final GeneratedColumn<String> currency = GeneratedColumn<String>(
    'currency',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
    defaultValue: const Constant('TRY'),
  );
  static const VerificationMeta _notesMeta = const VerificationMeta('notes');
  @override
  late final GeneratedColumn<String> notes = GeneratedColumn<String>(
    'notes',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _updatedAtMeta = const VerificationMeta(
    'updatedAt',
  );
  @override
  late final GeneratedColumn<DateTime> updatedAt = GeneratedColumn<DateTime>(
    'updated_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    defaultValue: currentDateAndTime,
  );
  static const VerificationMeta _pendingSyncMeta = const VerificationMeta(
    'pendingSync',
  );
  @override
  late final GeneratedColumn<bool> pendingSync = GeneratedColumn<bool>(
    'pending_sync',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'CHECK ("pending_sync" IN (0, 1))',
    ),
    defaultValue: const Constant(false),
  );
  @override
  List<GeneratedColumn> get $columns => [
    productId,
    variantId,
    title,
    imageUrl,
    quantity,
    unitPriceMinor,
    currency,
    notes,
    updatedAt,
    pendingSync,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'cart_items';
  @override
  VerificationContext validateIntegrity(
    Insertable<CartItem> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('product_id')) {
      context.handle(
        _productIdMeta,
        productId.isAcceptableOrUnknown(data['product_id']!, _productIdMeta),
      );
    } else if (isInserting) {
      context.missing(_productIdMeta);
    }
    if (data.containsKey('variant_id')) {
      context.handle(
        _variantIdMeta,
        variantId.isAcceptableOrUnknown(data['variant_id']!, _variantIdMeta),
      );
    }
    if (data.containsKey('title')) {
      context.handle(
        _titleMeta,
        title.isAcceptableOrUnknown(data['title']!, _titleMeta),
      );
    } else if (isInserting) {
      context.missing(_titleMeta);
    }
    if (data.containsKey('image_url')) {
      context.handle(
        _imageUrlMeta,
        imageUrl.isAcceptableOrUnknown(data['image_url']!, _imageUrlMeta),
      );
    }
    if (data.containsKey('quantity')) {
      context.handle(
        _quantityMeta,
        quantity.isAcceptableOrUnknown(data['quantity']!, _quantityMeta),
      );
    }
    if (data.containsKey('unit_price_minor')) {
      context.handle(
        _unitPriceMinorMeta,
        unitPriceMinor.isAcceptableOrUnknown(
          data['unit_price_minor']!,
          _unitPriceMinorMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_unitPriceMinorMeta);
    }
    if (data.containsKey('currency')) {
      context.handle(
        _currencyMeta,
        currency.isAcceptableOrUnknown(data['currency']!, _currencyMeta),
      );
    }
    if (data.containsKey('notes')) {
      context.handle(
        _notesMeta,
        notes.isAcceptableOrUnknown(data['notes']!, _notesMeta),
      );
    }
    if (data.containsKey('updated_at')) {
      context.handle(
        _updatedAtMeta,
        updatedAt.isAcceptableOrUnknown(data['updated_at']!, _updatedAtMeta),
      );
    }
    if (data.containsKey('pending_sync')) {
      context.handle(
        _pendingSyncMeta,
        pendingSync.isAcceptableOrUnknown(
          data['pending_sync']!,
          _pendingSyncMeta,
        ),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {productId, variantId};
  @override
  CartItem map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return CartItem(
      productId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}product_id'],
      )!,
      variantId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}variant_id'],
      ),
      title: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}title'],
      )!,
      imageUrl: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}image_url'],
      ),
      quantity: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}quantity'],
      )!,
      unitPriceMinor: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}unit_price_minor'],
      )!,
      currency: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}currency'],
      )!,
      notes: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}notes'],
      ),
      updatedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}updated_at'],
      )!,
      pendingSync: attachedDatabase.typeMapping.read(
        DriftSqlType.bool,
        data['${effectivePrefix}pending_sync'],
      )!,
    );
  }

  @override
  $CartItemsTable createAlias(String alias) {
    return $CartItemsTable(attachedDatabase, alias);
  }
}

class CartItem extends DataClass implements Insertable<CartItem> {
  final String productId;
  final String? variantId;
  final String title;
  final String? imageUrl;
  final int quantity;
  final int unitPriceMinor;
  final String currency;
  final String? notes;
  final DateTime updatedAt;
  final bool pendingSync;
  const CartItem({
    required this.productId,
    this.variantId,
    required this.title,
    this.imageUrl,
    required this.quantity,
    required this.unitPriceMinor,
    required this.currency,
    this.notes,
    required this.updatedAt,
    required this.pendingSync,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['product_id'] = Variable<String>(productId);
    if (!nullToAbsent || variantId != null) {
      map['variant_id'] = Variable<String>(variantId);
    }
    map['title'] = Variable<String>(title);
    if (!nullToAbsent || imageUrl != null) {
      map['image_url'] = Variable<String>(imageUrl);
    }
    map['quantity'] = Variable<int>(quantity);
    map['unit_price_minor'] = Variable<int>(unitPriceMinor);
    map['currency'] = Variable<String>(currency);
    if (!nullToAbsent || notes != null) {
      map['notes'] = Variable<String>(notes);
    }
    map['updated_at'] = Variable<DateTime>(updatedAt);
    map['pending_sync'] = Variable<bool>(pendingSync);
    return map;
  }

  CartItemsCompanion toCompanion(bool nullToAbsent) {
    return CartItemsCompanion(
      productId: Value(productId),
      variantId: variantId == null && nullToAbsent
          ? const Value.absent()
          : Value(variantId),
      title: Value(title),
      imageUrl: imageUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(imageUrl),
      quantity: Value(quantity),
      unitPriceMinor: Value(unitPriceMinor),
      currency: Value(currency),
      notes: notes == null && nullToAbsent
          ? const Value.absent()
          : Value(notes),
      updatedAt: Value(updatedAt),
      pendingSync: Value(pendingSync),
    );
  }

  factory CartItem.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return CartItem(
      productId: serializer.fromJson<String>(json['productId']),
      variantId: serializer.fromJson<String?>(json['variantId']),
      title: serializer.fromJson<String>(json['title']),
      imageUrl: serializer.fromJson<String?>(json['imageUrl']),
      quantity: serializer.fromJson<int>(json['quantity']),
      unitPriceMinor: serializer.fromJson<int>(json['unitPriceMinor']),
      currency: serializer.fromJson<String>(json['currency']),
      notes: serializer.fromJson<String?>(json['notes']),
      updatedAt: serializer.fromJson<DateTime>(json['updatedAt']),
      pendingSync: serializer.fromJson<bool>(json['pendingSync']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'productId': serializer.toJson<String>(productId),
      'variantId': serializer.toJson<String?>(variantId),
      'title': serializer.toJson<String>(title),
      'imageUrl': serializer.toJson<String?>(imageUrl),
      'quantity': serializer.toJson<int>(quantity),
      'unitPriceMinor': serializer.toJson<int>(unitPriceMinor),
      'currency': serializer.toJson<String>(currency),
      'notes': serializer.toJson<String?>(notes),
      'updatedAt': serializer.toJson<DateTime>(updatedAt),
      'pendingSync': serializer.toJson<bool>(pendingSync),
    };
  }

  CartItem copyWith({
    String? productId,
    Value<String?> variantId = const Value.absent(),
    String? title,
    Value<String?> imageUrl = const Value.absent(),
    int? quantity,
    int? unitPriceMinor,
    String? currency,
    Value<String?> notes = const Value.absent(),
    DateTime? updatedAt,
    bool? pendingSync,
  }) => CartItem(
    productId: productId ?? this.productId,
    variantId: variantId.present ? variantId.value : this.variantId,
    title: title ?? this.title,
    imageUrl: imageUrl.present ? imageUrl.value : this.imageUrl,
    quantity: quantity ?? this.quantity,
    unitPriceMinor: unitPriceMinor ?? this.unitPriceMinor,
    currency: currency ?? this.currency,
    notes: notes.present ? notes.value : this.notes,
    updatedAt: updatedAt ?? this.updatedAt,
    pendingSync: pendingSync ?? this.pendingSync,
  );
  CartItem copyWithCompanion(CartItemsCompanion data) {
    return CartItem(
      productId: data.productId.present ? data.productId.value : this.productId,
      variantId: data.variantId.present ? data.variantId.value : this.variantId,
      title: data.title.present ? data.title.value : this.title,
      imageUrl: data.imageUrl.present ? data.imageUrl.value : this.imageUrl,
      quantity: data.quantity.present ? data.quantity.value : this.quantity,
      unitPriceMinor: data.unitPriceMinor.present
          ? data.unitPriceMinor.value
          : this.unitPriceMinor,
      currency: data.currency.present ? data.currency.value : this.currency,
      notes: data.notes.present ? data.notes.value : this.notes,
      updatedAt: data.updatedAt.present ? data.updatedAt.value : this.updatedAt,
      pendingSync: data.pendingSync.present
          ? data.pendingSync.value
          : this.pendingSync,
    );
  }

  @override
  String toString() {
    return (StringBuffer('CartItem(')
          ..write('productId: $productId, ')
          ..write('variantId: $variantId, ')
          ..write('title: $title, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('quantity: $quantity, ')
          ..write('unitPriceMinor: $unitPriceMinor, ')
          ..write('currency: $currency, ')
          ..write('notes: $notes, ')
          ..write('updatedAt: $updatedAt, ')
          ..write('pendingSync: $pendingSync')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    productId,
    variantId,
    title,
    imageUrl,
    quantity,
    unitPriceMinor,
    currency,
    notes,
    updatedAt,
    pendingSync,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is CartItem &&
          other.productId == this.productId &&
          other.variantId == this.variantId &&
          other.title == this.title &&
          other.imageUrl == this.imageUrl &&
          other.quantity == this.quantity &&
          other.unitPriceMinor == this.unitPriceMinor &&
          other.currency == this.currency &&
          other.notes == this.notes &&
          other.updatedAt == this.updatedAt &&
          other.pendingSync == this.pendingSync);
}

class CartItemsCompanion extends UpdateCompanion<CartItem> {
  final Value<String> productId;
  final Value<String?> variantId;
  final Value<String> title;
  final Value<String?> imageUrl;
  final Value<int> quantity;
  final Value<int> unitPriceMinor;
  final Value<String> currency;
  final Value<String?> notes;
  final Value<DateTime> updatedAt;
  final Value<bool> pendingSync;
  final Value<int> rowid;
  const CartItemsCompanion({
    this.productId = const Value.absent(),
    this.variantId = const Value.absent(),
    this.title = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.quantity = const Value.absent(),
    this.unitPriceMinor = const Value.absent(),
    this.currency = const Value.absent(),
    this.notes = const Value.absent(),
    this.updatedAt = const Value.absent(),
    this.pendingSync = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  CartItemsCompanion.insert({
    required String productId,
    this.variantId = const Value.absent(),
    required String title,
    this.imageUrl = const Value.absent(),
    this.quantity = const Value.absent(),
    required int unitPriceMinor,
    this.currency = const Value.absent(),
    this.notes = const Value.absent(),
    this.updatedAt = const Value.absent(),
    this.pendingSync = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : productId = Value(productId),
       title = Value(title),
       unitPriceMinor = Value(unitPriceMinor);
  static Insertable<CartItem> custom({
    Expression<String>? productId,
    Expression<String>? variantId,
    Expression<String>? title,
    Expression<String>? imageUrl,
    Expression<int>? quantity,
    Expression<int>? unitPriceMinor,
    Expression<String>? currency,
    Expression<String>? notes,
    Expression<DateTime>? updatedAt,
    Expression<bool>? pendingSync,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (productId != null) 'product_id': productId,
      if (variantId != null) 'variant_id': variantId,
      if (title != null) 'title': title,
      if (imageUrl != null) 'image_url': imageUrl,
      if (quantity != null) 'quantity': quantity,
      if (unitPriceMinor != null) 'unit_price_minor': unitPriceMinor,
      if (currency != null) 'currency': currency,
      if (notes != null) 'notes': notes,
      if (updatedAt != null) 'updated_at': updatedAt,
      if (pendingSync != null) 'pending_sync': pendingSync,
      if (rowid != null) 'rowid': rowid,
    });
  }

  CartItemsCompanion copyWith({
    Value<String>? productId,
    Value<String?>? variantId,
    Value<String>? title,
    Value<String?>? imageUrl,
    Value<int>? quantity,
    Value<int>? unitPriceMinor,
    Value<String>? currency,
    Value<String?>? notes,
    Value<DateTime>? updatedAt,
    Value<bool>? pendingSync,
    Value<int>? rowid,
  }) {
    return CartItemsCompanion(
      productId: productId ?? this.productId,
      variantId: variantId ?? this.variantId,
      title: title ?? this.title,
      imageUrl: imageUrl ?? this.imageUrl,
      quantity: quantity ?? this.quantity,
      unitPriceMinor: unitPriceMinor ?? this.unitPriceMinor,
      currency: currency ?? this.currency,
      notes: notes ?? this.notes,
      updatedAt: updatedAt ?? this.updatedAt,
      pendingSync: pendingSync ?? this.pendingSync,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (productId.present) {
      map['product_id'] = Variable<String>(productId.value);
    }
    if (variantId.present) {
      map['variant_id'] = Variable<String>(variantId.value);
    }
    if (title.present) {
      map['title'] = Variable<String>(title.value);
    }
    if (imageUrl.present) {
      map['image_url'] = Variable<String>(imageUrl.value);
    }
    if (quantity.present) {
      map['quantity'] = Variable<int>(quantity.value);
    }
    if (unitPriceMinor.present) {
      map['unit_price_minor'] = Variable<int>(unitPriceMinor.value);
    }
    if (currency.present) {
      map['currency'] = Variable<String>(currency.value);
    }
    if (notes.present) {
      map['notes'] = Variable<String>(notes.value);
    }
    if (updatedAt.present) {
      map['updated_at'] = Variable<DateTime>(updatedAt.value);
    }
    if (pendingSync.present) {
      map['pending_sync'] = Variable<bool>(pendingSync.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('CartItemsCompanion(')
          ..write('productId: $productId, ')
          ..write('variantId: $variantId, ')
          ..write('title: $title, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('quantity: $quantity, ')
          ..write('unitPriceMinor: $unitPriceMinor, ')
          ..write('currency: $currency, ')
          ..write('notes: $notes, ')
          ..write('updatedAt: $updatedAt, ')
          ..write('pendingSync: $pendingSync, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $FavoritesTable extends Favorites
    with TableInfo<$FavoritesTable, Favorite> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $FavoritesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _productIdMeta = const VerificationMeta(
    'productId',
  );
  @override
  late final GeneratedColumn<String> productId = GeneratedColumn<String>(
    'product_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _titleMeta = const VerificationMeta('title');
  @override
  late final GeneratedColumn<String> title = GeneratedColumn<String>(
    'title',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _imageUrlMeta = const VerificationMeta(
    'imageUrl',
  );
  @override
  late final GeneratedColumn<String> imageUrl = GeneratedColumn<String>(
    'image_url',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _unitPriceMinorMeta = const VerificationMeta(
    'unitPriceMinor',
  );
  @override
  late final GeneratedColumn<int> unitPriceMinor = GeneratedColumn<int>(
    'unit_price_minor',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _addedAtMeta = const VerificationMeta(
    'addedAt',
  );
  @override
  late final GeneratedColumn<DateTime> addedAt = GeneratedColumn<DateTime>(
    'added_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    defaultValue: currentDateAndTime,
  );
  @override
  List<GeneratedColumn> get $columns => [
    productId,
    title,
    imageUrl,
    unitPriceMinor,
    addedAt,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'favorites';
  @override
  VerificationContext validateIntegrity(
    Insertable<Favorite> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('product_id')) {
      context.handle(
        _productIdMeta,
        productId.isAcceptableOrUnknown(data['product_id']!, _productIdMeta),
      );
    } else if (isInserting) {
      context.missing(_productIdMeta);
    }
    if (data.containsKey('title')) {
      context.handle(
        _titleMeta,
        title.isAcceptableOrUnknown(data['title']!, _titleMeta),
      );
    } else if (isInserting) {
      context.missing(_titleMeta);
    }
    if (data.containsKey('image_url')) {
      context.handle(
        _imageUrlMeta,
        imageUrl.isAcceptableOrUnknown(data['image_url']!, _imageUrlMeta),
      );
    }
    if (data.containsKey('unit_price_minor')) {
      context.handle(
        _unitPriceMinorMeta,
        unitPriceMinor.isAcceptableOrUnknown(
          data['unit_price_minor']!,
          _unitPriceMinorMeta,
        ),
      );
    }
    if (data.containsKey('added_at')) {
      context.handle(
        _addedAtMeta,
        addedAt.isAcceptableOrUnknown(data['added_at']!, _addedAtMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {productId};
  @override
  Favorite map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return Favorite(
      productId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}product_id'],
      )!,
      title: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}title'],
      )!,
      imageUrl: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}image_url'],
      ),
      unitPriceMinor: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}unit_price_minor'],
      ),
      addedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}added_at'],
      )!,
    );
  }

  @override
  $FavoritesTable createAlias(String alias) {
    return $FavoritesTable(attachedDatabase, alias);
  }
}

class Favorite extends DataClass implements Insertable<Favorite> {
  final String productId;
  final String title;
  final String? imageUrl;
  final int? unitPriceMinor;
  final DateTime addedAt;
  const Favorite({
    required this.productId,
    required this.title,
    this.imageUrl,
    this.unitPriceMinor,
    required this.addedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['product_id'] = Variable<String>(productId);
    map['title'] = Variable<String>(title);
    if (!nullToAbsent || imageUrl != null) {
      map['image_url'] = Variable<String>(imageUrl);
    }
    if (!nullToAbsent || unitPriceMinor != null) {
      map['unit_price_minor'] = Variable<int>(unitPriceMinor);
    }
    map['added_at'] = Variable<DateTime>(addedAt);
    return map;
  }

  FavoritesCompanion toCompanion(bool nullToAbsent) {
    return FavoritesCompanion(
      productId: Value(productId),
      title: Value(title),
      imageUrl: imageUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(imageUrl),
      unitPriceMinor: unitPriceMinor == null && nullToAbsent
          ? const Value.absent()
          : Value(unitPriceMinor),
      addedAt: Value(addedAt),
    );
  }

  factory Favorite.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return Favorite(
      productId: serializer.fromJson<String>(json['productId']),
      title: serializer.fromJson<String>(json['title']),
      imageUrl: serializer.fromJson<String?>(json['imageUrl']),
      unitPriceMinor: serializer.fromJson<int?>(json['unitPriceMinor']),
      addedAt: serializer.fromJson<DateTime>(json['addedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'productId': serializer.toJson<String>(productId),
      'title': serializer.toJson<String>(title),
      'imageUrl': serializer.toJson<String?>(imageUrl),
      'unitPriceMinor': serializer.toJson<int?>(unitPriceMinor),
      'addedAt': serializer.toJson<DateTime>(addedAt),
    };
  }

  Favorite copyWith({
    String? productId,
    String? title,
    Value<String?> imageUrl = const Value.absent(),
    Value<int?> unitPriceMinor = const Value.absent(),
    DateTime? addedAt,
  }) => Favorite(
    productId: productId ?? this.productId,
    title: title ?? this.title,
    imageUrl: imageUrl.present ? imageUrl.value : this.imageUrl,
    unitPriceMinor: unitPriceMinor.present
        ? unitPriceMinor.value
        : this.unitPriceMinor,
    addedAt: addedAt ?? this.addedAt,
  );
  Favorite copyWithCompanion(FavoritesCompanion data) {
    return Favorite(
      productId: data.productId.present ? data.productId.value : this.productId,
      title: data.title.present ? data.title.value : this.title,
      imageUrl: data.imageUrl.present ? data.imageUrl.value : this.imageUrl,
      unitPriceMinor: data.unitPriceMinor.present
          ? data.unitPriceMinor.value
          : this.unitPriceMinor,
      addedAt: data.addedAt.present ? data.addedAt.value : this.addedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('Favorite(')
          ..write('productId: $productId, ')
          ..write('title: $title, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('unitPriceMinor: $unitPriceMinor, ')
          ..write('addedAt: $addedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode =>
      Object.hash(productId, title, imageUrl, unitPriceMinor, addedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is Favorite &&
          other.productId == this.productId &&
          other.title == this.title &&
          other.imageUrl == this.imageUrl &&
          other.unitPriceMinor == this.unitPriceMinor &&
          other.addedAt == this.addedAt);
}

class FavoritesCompanion extends UpdateCompanion<Favorite> {
  final Value<String> productId;
  final Value<String> title;
  final Value<String?> imageUrl;
  final Value<int?> unitPriceMinor;
  final Value<DateTime> addedAt;
  final Value<int> rowid;
  const FavoritesCompanion({
    this.productId = const Value.absent(),
    this.title = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.unitPriceMinor = const Value.absent(),
    this.addedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  FavoritesCompanion.insert({
    required String productId,
    required String title,
    this.imageUrl = const Value.absent(),
    this.unitPriceMinor = const Value.absent(),
    this.addedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : productId = Value(productId),
       title = Value(title);
  static Insertable<Favorite> custom({
    Expression<String>? productId,
    Expression<String>? title,
    Expression<String>? imageUrl,
    Expression<int>? unitPriceMinor,
    Expression<DateTime>? addedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (productId != null) 'product_id': productId,
      if (title != null) 'title': title,
      if (imageUrl != null) 'image_url': imageUrl,
      if (unitPriceMinor != null) 'unit_price_minor': unitPriceMinor,
      if (addedAt != null) 'added_at': addedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  FavoritesCompanion copyWith({
    Value<String>? productId,
    Value<String>? title,
    Value<String?>? imageUrl,
    Value<int?>? unitPriceMinor,
    Value<DateTime>? addedAt,
    Value<int>? rowid,
  }) {
    return FavoritesCompanion(
      productId: productId ?? this.productId,
      title: title ?? this.title,
      imageUrl: imageUrl ?? this.imageUrl,
      unitPriceMinor: unitPriceMinor ?? this.unitPriceMinor,
      addedAt: addedAt ?? this.addedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (productId.present) {
      map['product_id'] = Variable<String>(productId.value);
    }
    if (title.present) {
      map['title'] = Variable<String>(title.value);
    }
    if (imageUrl.present) {
      map['image_url'] = Variable<String>(imageUrl.value);
    }
    if (unitPriceMinor.present) {
      map['unit_price_minor'] = Variable<int>(unitPriceMinor.value);
    }
    if (addedAt.present) {
      map['added_at'] = Variable<DateTime>(addedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('FavoritesCompanion(')
          ..write('productId: $productId, ')
          ..write('title: $title, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('unitPriceMinor: $unitPriceMinor, ')
          ..write('addedAt: $addedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $FavoriteEntriesTable extends FavoriteEntries
    with TableInfo<$FavoriteEntriesTable, FavoriteEntryRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $FavoriteEntriesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _entryIdMeta = const VerificationMeta(
    'entryId',
  );
  @override
  late final GeneratedColumn<String> entryId = GeneratedColumn<String>(
    'entry_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _typeMeta = const VerificationMeta('type');
  @override
  late final GeneratedColumn<String> type = GeneratedColumn<String>(
    'type',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _targetIdMeta = const VerificationMeta(
    'targetId',
  );
  @override
  late final GeneratedColumn<String> targetId = GeneratedColumn<String>(
    'target_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _titleMeta = const VerificationMeta('title');
  @override
  late final GeneratedColumn<String> title = GeneratedColumn<String>(
    'title',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _subtitleMeta = const VerificationMeta(
    'subtitle',
  );
  @override
  late final GeneratedColumn<String> subtitle = GeneratedColumn<String>(
    'subtitle',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
    defaultValue: const Constant(''),
  );
  static const VerificationMeta _imageUrlMeta = const VerificationMeta(
    'imageUrl',
  );
  @override
  late final GeneratedColumn<String> imageUrl = GeneratedColumn<String>(
    'image_url',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _metadataJsonMeta = const VerificationMeta(
    'metadataJson',
  );
  @override
  late final GeneratedColumn<String> metadataJson = GeneratedColumn<String>(
    'metadata_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
    defaultValue: const Constant('{}'),
  );
  static const VerificationMeta _addedAtMeta = const VerificationMeta(
    'addedAt',
  );
  @override
  late final GeneratedColumn<DateTime> addedAt = GeneratedColumn<DateTime>(
    'added_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    defaultValue: currentDateAndTime,
  );
  static const VerificationMeta _pendingSyncMeta = const VerificationMeta(
    'pendingSync',
  );
  @override
  late final GeneratedColumn<bool> pendingSync = GeneratedColumn<bool>(
    'pending_sync',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'CHECK ("pending_sync" IN (0, 1))',
    ),
    defaultValue: const Constant(false),
  );
  @override
  List<GeneratedColumn> get $columns => [
    entryId,
    type,
    targetId,
    title,
    subtitle,
    imageUrl,
    metadataJson,
    addedAt,
    pendingSync,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'favorite_entries';
  @override
  VerificationContext validateIntegrity(
    Insertable<FavoriteEntryRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('entry_id')) {
      context.handle(
        _entryIdMeta,
        entryId.isAcceptableOrUnknown(data['entry_id']!, _entryIdMeta),
      );
    } else if (isInserting) {
      context.missing(_entryIdMeta);
    }
    if (data.containsKey('type')) {
      context.handle(
        _typeMeta,
        type.isAcceptableOrUnknown(data['type']!, _typeMeta),
      );
    } else if (isInserting) {
      context.missing(_typeMeta);
    }
    if (data.containsKey('target_id')) {
      context.handle(
        _targetIdMeta,
        targetId.isAcceptableOrUnknown(data['target_id']!, _targetIdMeta),
      );
    } else if (isInserting) {
      context.missing(_targetIdMeta);
    }
    if (data.containsKey('title')) {
      context.handle(
        _titleMeta,
        title.isAcceptableOrUnknown(data['title']!, _titleMeta),
      );
    } else if (isInserting) {
      context.missing(_titleMeta);
    }
    if (data.containsKey('subtitle')) {
      context.handle(
        _subtitleMeta,
        subtitle.isAcceptableOrUnknown(data['subtitle']!, _subtitleMeta),
      );
    }
    if (data.containsKey('image_url')) {
      context.handle(
        _imageUrlMeta,
        imageUrl.isAcceptableOrUnknown(data['image_url']!, _imageUrlMeta),
      );
    }
    if (data.containsKey('metadata_json')) {
      context.handle(
        _metadataJsonMeta,
        metadataJson.isAcceptableOrUnknown(
          data['metadata_json']!,
          _metadataJsonMeta,
        ),
      );
    }
    if (data.containsKey('added_at')) {
      context.handle(
        _addedAtMeta,
        addedAt.isAcceptableOrUnknown(data['added_at']!, _addedAtMeta),
      );
    }
    if (data.containsKey('pending_sync')) {
      context.handle(
        _pendingSyncMeta,
        pendingSync.isAcceptableOrUnknown(
          data['pending_sync']!,
          _pendingSyncMeta,
        ),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {entryId};
  @override
  FavoriteEntryRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return FavoriteEntryRow(
      entryId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}entry_id'],
      )!,
      type: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}type'],
      )!,
      targetId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}target_id'],
      )!,
      title: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}title'],
      )!,
      subtitle: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}subtitle'],
      )!,
      imageUrl: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}image_url'],
      ),
      metadataJson: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}metadata_json'],
      )!,
      addedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}added_at'],
      )!,
      pendingSync: attachedDatabase.typeMapping.read(
        DriftSqlType.bool,
        data['${effectivePrefix}pending_sync'],
      )!,
    );
  }

  @override
  $FavoriteEntriesTable createAlias(String alias) {
    return $FavoriteEntriesTable(attachedDatabase, alias);
  }
}

class FavoriteEntryRow extends DataClass
    implements Insertable<FavoriteEntryRow> {
  final String entryId;
  final String type;
  final String targetId;
  final String title;
  final String subtitle;
  final String? imageUrl;
  final String metadataJson;
  final DateTime addedAt;
  final bool pendingSync;
  const FavoriteEntryRow({
    required this.entryId,
    required this.type,
    required this.targetId,
    required this.title,
    required this.subtitle,
    this.imageUrl,
    required this.metadataJson,
    required this.addedAt,
    required this.pendingSync,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['entry_id'] = Variable<String>(entryId);
    map['type'] = Variable<String>(type);
    map['target_id'] = Variable<String>(targetId);
    map['title'] = Variable<String>(title);
    map['subtitle'] = Variable<String>(subtitle);
    if (!nullToAbsent || imageUrl != null) {
      map['image_url'] = Variable<String>(imageUrl);
    }
    map['metadata_json'] = Variable<String>(metadataJson);
    map['added_at'] = Variable<DateTime>(addedAt);
    map['pending_sync'] = Variable<bool>(pendingSync);
    return map;
  }

  FavoriteEntriesCompanion toCompanion(bool nullToAbsent) {
    return FavoriteEntriesCompanion(
      entryId: Value(entryId),
      type: Value(type),
      targetId: Value(targetId),
      title: Value(title),
      subtitle: Value(subtitle),
      imageUrl: imageUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(imageUrl),
      metadataJson: Value(metadataJson),
      addedAt: Value(addedAt),
      pendingSync: Value(pendingSync),
    );
  }

  factory FavoriteEntryRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return FavoriteEntryRow(
      entryId: serializer.fromJson<String>(json['entryId']),
      type: serializer.fromJson<String>(json['type']),
      targetId: serializer.fromJson<String>(json['targetId']),
      title: serializer.fromJson<String>(json['title']),
      subtitle: serializer.fromJson<String>(json['subtitle']),
      imageUrl: serializer.fromJson<String?>(json['imageUrl']),
      metadataJson: serializer.fromJson<String>(json['metadataJson']),
      addedAt: serializer.fromJson<DateTime>(json['addedAt']),
      pendingSync: serializer.fromJson<bool>(json['pendingSync']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'entryId': serializer.toJson<String>(entryId),
      'type': serializer.toJson<String>(type),
      'targetId': serializer.toJson<String>(targetId),
      'title': serializer.toJson<String>(title),
      'subtitle': serializer.toJson<String>(subtitle),
      'imageUrl': serializer.toJson<String?>(imageUrl),
      'metadataJson': serializer.toJson<String>(metadataJson),
      'addedAt': serializer.toJson<DateTime>(addedAt),
      'pendingSync': serializer.toJson<bool>(pendingSync),
    };
  }

  FavoriteEntryRow copyWith({
    String? entryId,
    String? type,
    String? targetId,
    String? title,
    String? subtitle,
    Value<String?> imageUrl = const Value.absent(),
    String? metadataJson,
    DateTime? addedAt,
    bool? pendingSync,
  }) => FavoriteEntryRow(
    entryId: entryId ?? this.entryId,
    type: type ?? this.type,
    targetId: targetId ?? this.targetId,
    title: title ?? this.title,
    subtitle: subtitle ?? this.subtitle,
    imageUrl: imageUrl.present ? imageUrl.value : this.imageUrl,
    metadataJson: metadataJson ?? this.metadataJson,
    addedAt: addedAt ?? this.addedAt,
    pendingSync: pendingSync ?? this.pendingSync,
  );
  FavoriteEntryRow copyWithCompanion(FavoriteEntriesCompanion data) {
    return FavoriteEntryRow(
      entryId: data.entryId.present ? data.entryId.value : this.entryId,
      type: data.type.present ? data.type.value : this.type,
      targetId: data.targetId.present ? data.targetId.value : this.targetId,
      title: data.title.present ? data.title.value : this.title,
      subtitle: data.subtitle.present ? data.subtitle.value : this.subtitle,
      imageUrl: data.imageUrl.present ? data.imageUrl.value : this.imageUrl,
      metadataJson: data.metadataJson.present
          ? data.metadataJson.value
          : this.metadataJson,
      addedAt: data.addedAt.present ? data.addedAt.value : this.addedAt,
      pendingSync: data.pendingSync.present
          ? data.pendingSync.value
          : this.pendingSync,
    );
  }

  @override
  String toString() {
    return (StringBuffer('FavoriteEntryRow(')
          ..write('entryId: $entryId, ')
          ..write('type: $type, ')
          ..write('targetId: $targetId, ')
          ..write('title: $title, ')
          ..write('subtitle: $subtitle, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('metadataJson: $metadataJson, ')
          ..write('addedAt: $addedAt, ')
          ..write('pendingSync: $pendingSync')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    entryId,
    type,
    targetId,
    title,
    subtitle,
    imageUrl,
    metadataJson,
    addedAt,
    pendingSync,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is FavoriteEntryRow &&
          other.entryId == this.entryId &&
          other.type == this.type &&
          other.targetId == this.targetId &&
          other.title == this.title &&
          other.subtitle == this.subtitle &&
          other.imageUrl == this.imageUrl &&
          other.metadataJson == this.metadataJson &&
          other.addedAt == this.addedAt &&
          other.pendingSync == this.pendingSync);
}

class FavoriteEntriesCompanion extends UpdateCompanion<FavoriteEntryRow> {
  final Value<String> entryId;
  final Value<String> type;
  final Value<String> targetId;
  final Value<String> title;
  final Value<String> subtitle;
  final Value<String?> imageUrl;
  final Value<String> metadataJson;
  final Value<DateTime> addedAt;
  final Value<bool> pendingSync;
  final Value<int> rowid;
  const FavoriteEntriesCompanion({
    this.entryId = const Value.absent(),
    this.type = const Value.absent(),
    this.targetId = const Value.absent(),
    this.title = const Value.absent(),
    this.subtitle = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.metadataJson = const Value.absent(),
    this.addedAt = const Value.absent(),
    this.pendingSync = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  FavoriteEntriesCompanion.insert({
    required String entryId,
    required String type,
    required String targetId,
    required String title,
    this.subtitle = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.metadataJson = const Value.absent(),
    this.addedAt = const Value.absent(),
    this.pendingSync = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : entryId = Value(entryId),
       type = Value(type),
       targetId = Value(targetId),
       title = Value(title);
  static Insertable<FavoriteEntryRow> custom({
    Expression<String>? entryId,
    Expression<String>? type,
    Expression<String>? targetId,
    Expression<String>? title,
    Expression<String>? subtitle,
    Expression<String>? imageUrl,
    Expression<String>? metadataJson,
    Expression<DateTime>? addedAt,
    Expression<bool>? pendingSync,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (entryId != null) 'entry_id': entryId,
      if (type != null) 'type': type,
      if (targetId != null) 'target_id': targetId,
      if (title != null) 'title': title,
      if (subtitle != null) 'subtitle': subtitle,
      if (imageUrl != null) 'image_url': imageUrl,
      if (metadataJson != null) 'metadata_json': metadataJson,
      if (addedAt != null) 'added_at': addedAt,
      if (pendingSync != null) 'pending_sync': pendingSync,
      if (rowid != null) 'rowid': rowid,
    });
  }

  FavoriteEntriesCompanion copyWith({
    Value<String>? entryId,
    Value<String>? type,
    Value<String>? targetId,
    Value<String>? title,
    Value<String>? subtitle,
    Value<String?>? imageUrl,
    Value<String>? metadataJson,
    Value<DateTime>? addedAt,
    Value<bool>? pendingSync,
    Value<int>? rowid,
  }) {
    return FavoriteEntriesCompanion(
      entryId: entryId ?? this.entryId,
      type: type ?? this.type,
      targetId: targetId ?? this.targetId,
      title: title ?? this.title,
      subtitle: subtitle ?? this.subtitle,
      imageUrl: imageUrl ?? this.imageUrl,
      metadataJson: metadataJson ?? this.metadataJson,
      addedAt: addedAt ?? this.addedAt,
      pendingSync: pendingSync ?? this.pendingSync,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (entryId.present) {
      map['entry_id'] = Variable<String>(entryId.value);
    }
    if (type.present) {
      map['type'] = Variable<String>(type.value);
    }
    if (targetId.present) {
      map['target_id'] = Variable<String>(targetId.value);
    }
    if (title.present) {
      map['title'] = Variable<String>(title.value);
    }
    if (subtitle.present) {
      map['subtitle'] = Variable<String>(subtitle.value);
    }
    if (imageUrl.present) {
      map['image_url'] = Variable<String>(imageUrl.value);
    }
    if (metadataJson.present) {
      map['metadata_json'] = Variable<String>(metadataJson.value);
    }
    if (addedAt.present) {
      map['added_at'] = Variable<DateTime>(addedAt.value);
    }
    if (pendingSync.present) {
      map['pending_sync'] = Variable<bool>(pendingSync.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('FavoriteEntriesCompanion(')
          ..write('entryId: $entryId, ')
          ..write('type: $type, ')
          ..write('targetId: $targetId, ')
          ..write('title: $title, ')
          ..write('subtitle: $subtitle, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('metadataJson: $metadataJson, ')
          ..write('addedAt: $addedAt, ')
          ..write('pendingSync: $pendingSync, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $RecentSearchesTable extends RecentSearches
    with TableInfo<$RecentSearchesTable, RecentSearche> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $RecentSearchesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
    'id',
    aliasedName,
    false,
    hasAutoIncrement: true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'PRIMARY KEY AUTOINCREMENT',
    ),
  );
  static const VerificationMeta _queryMeta = const VerificationMeta('query');
  @override
  late final GeneratedColumn<String> query = GeneratedColumn<String>(
    'query',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _searchedAtMeta = const VerificationMeta(
    'searchedAt',
  );
  @override
  late final GeneratedColumn<DateTime> searchedAt = GeneratedColumn<DateTime>(
    'searched_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    defaultValue: currentDateAndTime,
  );
  @override
  List<GeneratedColumn> get $columns => [id, query, searchedAt];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'recent_searches';
  @override
  VerificationContext validateIntegrity(
    Insertable<RecentSearche> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('query')) {
      context.handle(
        _queryMeta,
        query.isAcceptableOrUnknown(data['query']!, _queryMeta),
      );
    } else if (isInserting) {
      context.missing(_queryMeta);
    }
    if (data.containsKey('searched_at')) {
      context.handle(
        _searchedAtMeta,
        searchedAt.isAcceptableOrUnknown(data['searched_at']!, _searchedAtMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  RecentSearche map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return RecentSearche(
      id: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}id'],
      )!,
      query: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}query'],
      )!,
      searchedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}searched_at'],
      )!,
    );
  }

  @override
  $RecentSearchesTable createAlias(String alias) {
    return $RecentSearchesTable(attachedDatabase, alias);
  }
}

class RecentSearche extends DataClass implements Insertable<RecentSearche> {
  final int id;
  final String query;
  final DateTime searchedAt;
  const RecentSearche({
    required this.id,
    required this.query,
    required this.searchedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['query'] = Variable<String>(query);
    map['searched_at'] = Variable<DateTime>(searchedAt);
    return map;
  }

  RecentSearchesCompanion toCompanion(bool nullToAbsent) {
    return RecentSearchesCompanion(
      id: Value(id),
      query: Value(query),
      searchedAt: Value(searchedAt),
    );
  }

  factory RecentSearche.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return RecentSearche(
      id: serializer.fromJson<int>(json['id']),
      query: serializer.fromJson<String>(json['query']),
      searchedAt: serializer.fromJson<DateTime>(json['searchedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'query': serializer.toJson<String>(query),
      'searchedAt': serializer.toJson<DateTime>(searchedAt),
    };
  }

  RecentSearche copyWith({int? id, String? query, DateTime? searchedAt}) =>
      RecentSearche(
        id: id ?? this.id,
        query: query ?? this.query,
        searchedAt: searchedAt ?? this.searchedAt,
      );
  RecentSearche copyWithCompanion(RecentSearchesCompanion data) {
    return RecentSearche(
      id: data.id.present ? data.id.value : this.id,
      query: data.query.present ? data.query.value : this.query,
      searchedAt: data.searchedAt.present
          ? data.searchedAt.value
          : this.searchedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('RecentSearche(')
          ..write('id: $id, ')
          ..write('query: $query, ')
          ..write('searchedAt: $searchedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, query, searchedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is RecentSearche &&
          other.id == this.id &&
          other.query == this.query &&
          other.searchedAt == this.searchedAt);
}

class RecentSearchesCompanion extends UpdateCompanion<RecentSearche> {
  final Value<int> id;
  final Value<String> query;
  final Value<DateTime> searchedAt;
  const RecentSearchesCompanion({
    this.id = const Value.absent(),
    this.query = const Value.absent(),
    this.searchedAt = const Value.absent(),
  });
  RecentSearchesCompanion.insert({
    this.id = const Value.absent(),
    required String query,
    this.searchedAt = const Value.absent(),
  }) : query = Value(query);
  static Insertable<RecentSearche> custom({
    Expression<int>? id,
    Expression<String>? query,
    Expression<DateTime>? searchedAt,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (query != null) 'query': query,
      if (searchedAt != null) 'searched_at': searchedAt,
    });
  }

  RecentSearchesCompanion copyWith({
    Value<int>? id,
    Value<String>? query,
    Value<DateTime>? searchedAt,
  }) {
    return RecentSearchesCompanion(
      id: id ?? this.id,
      query: query ?? this.query,
      searchedAt: searchedAt ?? this.searchedAt,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (query.present) {
      map['query'] = Variable<String>(query.value);
    }
    if (searchedAt.present) {
      map['searched_at'] = Variable<DateTime>(searchedAt.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('RecentSearchesCompanion(')
          ..write('id: $id, ')
          ..write('query: $query, ')
          ..write('searchedAt: $searchedAt')
          ..write(')'))
        .toString();
  }
}

class $CachedOrdersTable extends CachedOrders
    with TableInfo<$CachedOrdersTable, CachedOrder> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $CachedOrdersTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _orderIdMeta = const VerificationMeta(
    'orderId',
  );
  @override
  late final GeneratedColumn<String> orderId = GeneratedColumn<String>(
    'order_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
    'status',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _payloadJsonMeta = const VerificationMeta(
    'payloadJson',
  );
  @override
  late final GeneratedColumn<String> payloadJson = GeneratedColumn<String>(
    'payload_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _cachedAtMeta = const VerificationMeta(
    'cachedAt',
  );
  @override
  late final GeneratedColumn<DateTime> cachedAt = GeneratedColumn<DateTime>(
    'cached_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    defaultValue: currentDateAndTime,
  );
  @override
  List<GeneratedColumn> get $columns => [
    orderId,
    status,
    payloadJson,
    cachedAt,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'cached_orders';
  @override
  VerificationContext validateIntegrity(
    Insertable<CachedOrder> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('order_id')) {
      context.handle(
        _orderIdMeta,
        orderId.isAcceptableOrUnknown(data['order_id']!, _orderIdMeta),
      );
    } else if (isInserting) {
      context.missing(_orderIdMeta);
    }
    if (data.containsKey('status')) {
      context.handle(
        _statusMeta,
        status.isAcceptableOrUnknown(data['status']!, _statusMeta),
      );
    } else if (isInserting) {
      context.missing(_statusMeta);
    }
    if (data.containsKey('payload_json')) {
      context.handle(
        _payloadJsonMeta,
        payloadJson.isAcceptableOrUnknown(
          data['payload_json']!,
          _payloadJsonMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_payloadJsonMeta);
    }
    if (data.containsKey('cached_at')) {
      context.handle(
        _cachedAtMeta,
        cachedAt.isAcceptableOrUnknown(data['cached_at']!, _cachedAtMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {orderId};
  @override
  CachedOrder map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return CachedOrder(
      orderId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}order_id'],
      )!,
      status: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}status'],
      )!,
      payloadJson: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}payload_json'],
      )!,
      cachedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}cached_at'],
      )!,
    );
  }

  @override
  $CachedOrdersTable createAlias(String alias) {
    return $CachedOrdersTable(attachedDatabase, alias);
  }
}

class CachedOrder extends DataClass implements Insertable<CachedOrder> {
  final String orderId;
  final String status;
  final String payloadJson;
  final DateTime cachedAt;
  const CachedOrder({
    required this.orderId,
    required this.status,
    required this.payloadJson,
    required this.cachedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['order_id'] = Variable<String>(orderId);
    map['status'] = Variable<String>(status);
    map['payload_json'] = Variable<String>(payloadJson);
    map['cached_at'] = Variable<DateTime>(cachedAt);
    return map;
  }

  CachedOrdersCompanion toCompanion(bool nullToAbsent) {
    return CachedOrdersCompanion(
      orderId: Value(orderId),
      status: Value(status),
      payloadJson: Value(payloadJson),
      cachedAt: Value(cachedAt),
    );
  }

  factory CachedOrder.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return CachedOrder(
      orderId: serializer.fromJson<String>(json['orderId']),
      status: serializer.fromJson<String>(json['status']),
      payloadJson: serializer.fromJson<String>(json['payloadJson']),
      cachedAt: serializer.fromJson<DateTime>(json['cachedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'orderId': serializer.toJson<String>(orderId),
      'status': serializer.toJson<String>(status),
      'payloadJson': serializer.toJson<String>(payloadJson),
      'cachedAt': serializer.toJson<DateTime>(cachedAt),
    };
  }

  CachedOrder copyWith({
    String? orderId,
    String? status,
    String? payloadJson,
    DateTime? cachedAt,
  }) => CachedOrder(
    orderId: orderId ?? this.orderId,
    status: status ?? this.status,
    payloadJson: payloadJson ?? this.payloadJson,
    cachedAt: cachedAt ?? this.cachedAt,
  );
  CachedOrder copyWithCompanion(CachedOrdersCompanion data) {
    return CachedOrder(
      orderId: data.orderId.present ? data.orderId.value : this.orderId,
      status: data.status.present ? data.status.value : this.status,
      payloadJson: data.payloadJson.present
          ? data.payloadJson.value
          : this.payloadJson,
      cachedAt: data.cachedAt.present ? data.cachedAt.value : this.cachedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('CachedOrder(')
          ..write('orderId: $orderId, ')
          ..write('status: $status, ')
          ..write('payloadJson: $payloadJson, ')
          ..write('cachedAt: $cachedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(orderId, status, payloadJson, cachedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is CachedOrder &&
          other.orderId == this.orderId &&
          other.status == this.status &&
          other.payloadJson == this.payloadJson &&
          other.cachedAt == this.cachedAt);
}

class CachedOrdersCompanion extends UpdateCompanion<CachedOrder> {
  final Value<String> orderId;
  final Value<String> status;
  final Value<String> payloadJson;
  final Value<DateTime> cachedAt;
  final Value<int> rowid;
  const CachedOrdersCompanion({
    this.orderId = const Value.absent(),
    this.status = const Value.absent(),
    this.payloadJson = const Value.absent(),
    this.cachedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  CachedOrdersCompanion.insert({
    required String orderId,
    required String status,
    required String payloadJson,
    this.cachedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : orderId = Value(orderId),
       status = Value(status),
       payloadJson = Value(payloadJson);
  static Insertable<CachedOrder> custom({
    Expression<String>? orderId,
    Expression<String>? status,
    Expression<String>? payloadJson,
    Expression<DateTime>? cachedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (orderId != null) 'order_id': orderId,
      if (status != null) 'status': status,
      if (payloadJson != null) 'payload_json': payloadJson,
      if (cachedAt != null) 'cached_at': cachedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  CachedOrdersCompanion copyWith({
    Value<String>? orderId,
    Value<String>? status,
    Value<String>? payloadJson,
    Value<DateTime>? cachedAt,
    Value<int>? rowid,
  }) {
    return CachedOrdersCompanion(
      orderId: orderId ?? this.orderId,
      status: status ?? this.status,
      payloadJson: payloadJson ?? this.payloadJson,
      cachedAt: cachedAt ?? this.cachedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (orderId.present) {
      map['order_id'] = Variable<String>(orderId.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (payloadJson.present) {
      map['payload_json'] = Variable<String>(payloadJson.value);
    }
    if (cachedAt.present) {
      map['cached_at'] = Variable<DateTime>(cachedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('CachedOrdersCompanion(')
          ..write('orderId: $orderId, ')
          ..write('status: $status, ')
          ..write('payloadJson: $payloadJson, ')
          ..write('cachedAt: $cachedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $SyncQueueTable extends SyncQueue
    with TableInfo<$SyncQueueTable, SyncQueueData> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $SyncQueueTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _clientMutationIdMeta = const VerificationMeta(
    'clientMutationId',
  );
  @override
  late final GeneratedColumn<String> clientMutationId = GeneratedColumn<String>(
    'client_mutation_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _methodMeta = const VerificationMeta('method');
  @override
  late final GeneratedColumn<String> method = GeneratedColumn<String>(
    'method',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _pathMeta = const VerificationMeta('path');
  @override
  late final GeneratedColumn<String> path = GeneratedColumn<String>(
    'path',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _bodyJsonMeta = const VerificationMeta(
    'bodyJson',
  );
  @override
  late final GeneratedColumn<String> bodyJson = GeneratedColumn<String>(
    'body_json',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _idempotencyKeyMeta = const VerificationMeta(
    'idempotencyKey',
  );
  @override
  late final GeneratedColumn<String> idempotencyKey = GeneratedColumn<String>(
    'idempotency_key',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _retryCountMeta = const VerificationMeta(
    'retryCount',
  );
  @override
  late final GeneratedColumn<int> retryCount = GeneratedColumn<int>(
    'retry_count',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _createdAtMeta = const VerificationMeta(
    'createdAt',
  );
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
    'created_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    defaultValue: currentDateAndTime,
  );
  static const VerificationMeta _nextAttemptAtMeta = const VerificationMeta(
    'nextAttemptAt',
  );
  @override
  late final GeneratedColumn<DateTime> nextAttemptAt =
      GeneratedColumn<DateTime>(
        'next_attempt_at',
        aliasedName,
        true,
        type: DriftSqlType.dateTime,
        requiredDuringInsert: false,
      );
  static const VerificationMeta _lastErrorMeta = const VerificationMeta(
    'lastError',
  );
  @override
  late final GeneratedColumn<String> lastError = GeneratedColumn<String>(
    'last_error',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  @override
  List<GeneratedColumn> get $columns => [
    clientMutationId,
    method,
    path,
    bodyJson,
    idempotencyKey,
    retryCount,
    createdAt,
    nextAttemptAt,
    lastError,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'sync_queue';
  @override
  VerificationContext validateIntegrity(
    Insertable<SyncQueueData> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('client_mutation_id')) {
      context.handle(
        _clientMutationIdMeta,
        clientMutationId.isAcceptableOrUnknown(
          data['client_mutation_id']!,
          _clientMutationIdMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_clientMutationIdMeta);
    }
    if (data.containsKey('method')) {
      context.handle(
        _methodMeta,
        method.isAcceptableOrUnknown(data['method']!, _methodMeta),
      );
    } else if (isInserting) {
      context.missing(_methodMeta);
    }
    if (data.containsKey('path')) {
      context.handle(
        _pathMeta,
        path.isAcceptableOrUnknown(data['path']!, _pathMeta),
      );
    } else if (isInserting) {
      context.missing(_pathMeta);
    }
    if (data.containsKey('body_json')) {
      context.handle(
        _bodyJsonMeta,
        bodyJson.isAcceptableOrUnknown(data['body_json']!, _bodyJsonMeta),
      );
    }
    if (data.containsKey('idempotency_key')) {
      context.handle(
        _idempotencyKeyMeta,
        idempotencyKey.isAcceptableOrUnknown(
          data['idempotency_key']!,
          _idempotencyKeyMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_idempotencyKeyMeta);
    }
    if (data.containsKey('retry_count')) {
      context.handle(
        _retryCountMeta,
        retryCount.isAcceptableOrUnknown(data['retry_count']!, _retryCountMeta),
      );
    }
    if (data.containsKey('created_at')) {
      context.handle(
        _createdAtMeta,
        createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta),
      );
    }
    if (data.containsKey('next_attempt_at')) {
      context.handle(
        _nextAttemptAtMeta,
        nextAttemptAt.isAcceptableOrUnknown(
          data['next_attempt_at']!,
          _nextAttemptAtMeta,
        ),
      );
    }
    if (data.containsKey('last_error')) {
      context.handle(
        _lastErrorMeta,
        lastError.isAcceptableOrUnknown(data['last_error']!, _lastErrorMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {clientMutationId};
  @override
  SyncQueueData map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return SyncQueueData(
      clientMutationId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}client_mutation_id'],
      )!,
      method: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}method'],
      )!,
      path: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}path'],
      )!,
      bodyJson: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}body_json'],
      ),
      idempotencyKey: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}idempotency_key'],
      )!,
      retryCount: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}retry_count'],
      )!,
      createdAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}created_at'],
      )!,
      nextAttemptAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}next_attempt_at'],
      ),
      lastError: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}last_error'],
      ),
    );
  }

  @override
  $SyncQueueTable createAlias(String alias) {
    return $SyncQueueTable(attachedDatabase, alias);
  }
}

class SyncQueueData extends DataClass implements Insertable<SyncQueueData> {
  final String clientMutationId;
  final String method;
  final String path;
  final String? bodyJson;
  final String idempotencyKey;
  final int retryCount;
  final DateTime createdAt;
  final DateTime? nextAttemptAt;
  final String? lastError;
  const SyncQueueData({
    required this.clientMutationId,
    required this.method,
    required this.path,
    this.bodyJson,
    required this.idempotencyKey,
    required this.retryCount,
    required this.createdAt,
    this.nextAttemptAt,
    this.lastError,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['client_mutation_id'] = Variable<String>(clientMutationId);
    map['method'] = Variable<String>(method);
    map['path'] = Variable<String>(path);
    if (!nullToAbsent || bodyJson != null) {
      map['body_json'] = Variable<String>(bodyJson);
    }
    map['idempotency_key'] = Variable<String>(idempotencyKey);
    map['retry_count'] = Variable<int>(retryCount);
    map['created_at'] = Variable<DateTime>(createdAt);
    if (!nullToAbsent || nextAttemptAt != null) {
      map['next_attempt_at'] = Variable<DateTime>(nextAttemptAt);
    }
    if (!nullToAbsent || lastError != null) {
      map['last_error'] = Variable<String>(lastError);
    }
    return map;
  }

  SyncQueueCompanion toCompanion(bool nullToAbsent) {
    return SyncQueueCompanion(
      clientMutationId: Value(clientMutationId),
      method: Value(method),
      path: Value(path),
      bodyJson: bodyJson == null && nullToAbsent
          ? const Value.absent()
          : Value(bodyJson),
      idempotencyKey: Value(idempotencyKey),
      retryCount: Value(retryCount),
      createdAt: Value(createdAt),
      nextAttemptAt: nextAttemptAt == null && nullToAbsent
          ? const Value.absent()
          : Value(nextAttemptAt),
      lastError: lastError == null && nullToAbsent
          ? const Value.absent()
          : Value(lastError),
    );
  }

  factory SyncQueueData.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return SyncQueueData(
      clientMutationId: serializer.fromJson<String>(json['clientMutationId']),
      method: serializer.fromJson<String>(json['method']),
      path: serializer.fromJson<String>(json['path']),
      bodyJson: serializer.fromJson<String?>(json['bodyJson']),
      idempotencyKey: serializer.fromJson<String>(json['idempotencyKey']),
      retryCount: serializer.fromJson<int>(json['retryCount']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      nextAttemptAt: serializer.fromJson<DateTime?>(json['nextAttemptAt']),
      lastError: serializer.fromJson<String?>(json['lastError']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'clientMutationId': serializer.toJson<String>(clientMutationId),
      'method': serializer.toJson<String>(method),
      'path': serializer.toJson<String>(path),
      'bodyJson': serializer.toJson<String?>(bodyJson),
      'idempotencyKey': serializer.toJson<String>(idempotencyKey),
      'retryCount': serializer.toJson<int>(retryCount),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'nextAttemptAt': serializer.toJson<DateTime?>(nextAttemptAt),
      'lastError': serializer.toJson<String?>(lastError),
    };
  }

  SyncQueueData copyWith({
    String? clientMutationId,
    String? method,
    String? path,
    Value<String?> bodyJson = const Value.absent(),
    String? idempotencyKey,
    int? retryCount,
    DateTime? createdAt,
    Value<DateTime?> nextAttemptAt = const Value.absent(),
    Value<String?> lastError = const Value.absent(),
  }) => SyncQueueData(
    clientMutationId: clientMutationId ?? this.clientMutationId,
    method: method ?? this.method,
    path: path ?? this.path,
    bodyJson: bodyJson.present ? bodyJson.value : this.bodyJson,
    idempotencyKey: idempotencyKey ?? this.idempotencyKey,
    retryCount: retryCount ?? this.retryCount,
    createdAt: createdAt ?? this.createdAt,
    nextAttemptAt: nextAttemptAt.present
        ? nextAttemptAt.value
        : this.nextAttemptAt,
    lastError: lastError.present ? lastError.value : this.lastError,
  );
  SyncQueueData copyWithCompanion(SyncQueueCompanion data) {
    return SyncQueueData(
      clientMutationId: data.clientMutationId.present
          ? data.clientMutationId.value
          : this.clientMutationId,
      method: data.method.present ? data.method.value : this.method,
      path: data.path.present ? data.path.value : this.path,
      bodyJson: data.bodyJson.present ? data.bodyJson.value : this.bodyJson,
      idempotencyKey: data.idempotencyKey.present
          ? data.idempotencyKey.value
          : this.idempotencyKey,
      retryCount: data.retryCount.present
          ? data.retryCount.value
          : this.retryCount,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      nextAttemptAt: data.nextAttemptAt.present
          ? data.nextAttemptAt.value
          : this.nextAttemptAt,
      lastError: data.lastError.present ? data.lastError.value : this.lastError,
    );
  }

  @override
  String toString() {
    return (StringBuffer('SyncQueueData(')
          ..write('clientMutationId: $clientMutationId, ')
          ..write('method: $method, ')
          ..write('path: $path, ')
          ..write('bodyJson: $bodyJson, ')
          ..write('idempotencyKey: $idempotencyKey, ')
          ..write('retryCount: $retryCount, ')
          ..write('createdAt: $createdAt, ')
          ..write('nextAttemptAt: $nextAttemptAt, ')
          ..write('lastError: $lastError')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    clientMutationId,
    method,
    path,
    bodyJson,
    idempotencyKey,
    retryCount,
    createdAt,
    nextAttemptAt,
    lastError,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is SyncQueueData &&
          other.clientMutationId == this.clientMutationId &&
          other.method == this.method &&
          other.path == this.path &&
          other.bodyJson == this.bodyJson &&
          other.idempotencyKey == this.idempotencyKey &&
          other.retryCount == this.retryCount &&
          other.createdAt == this.createdAt &&
          other.nextAttemptAt == this.nextAttemptAt &&
          other.lastError == this.lastError);
}

class SyncQueueCompanion extends UpdateCompanion<SyncQueueData> {
  final Value<String> clientMutationId;
  final Value<String> method;
  final Value<String> path;
  final Value<String?> bodyJson;
  final Value<String> idempotencyKey;
  final Value<int> retryCount;
  final Value<DateTime> createdAt;
  final Value<DateTime?> nextAttemptAt;
  final Value<String?> lastError;
  final Value<int> rowid;
  const SyncQueueCompanion({
    this.clientMutationId = const Value.absent(),
    this.method = const Value.absent(),
    this.path = const Value.absent(),
    this.bodyJson = const Value.absent(),
    this.idempotencyKey = const Value.absent(),
    this.retryCount = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.nextAttemptAt = const Value.absent(),
    this.lastError = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  SyncQueueCompanion.insert({
    required String clientMutationId,
    required String method,
    required String path,
    this.bodyJson = const Value.absent(),
    required String idempotencyKey,
    this.retryCount = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.nextAttemptAt = const Value.absent(),
    this.lastError = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : clientMutationId = Value(clientMutationId),
       method = Value(method),
       path = Value(path),
       idempotencyKey = Value(idempotencyKey);
  static Insertable<SyncQueueData> custom({
    Expression<String>? clientMutationId,
    Expression<String>? method,
    Expression<String>? path,
    Expression<String>? bodyJson,
    Expression<String>? idempotencyKey,
    Expression<int>? retryCount,
    Expression<DateTime>? createdAt,
    Expression<DateTime>? nextAttemptAt,
    Expression<String>? lastError,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (clientMutationId != null) 'client_mutation_id': clientMutationId,
      if (method != null) 'method': method,
      if (path != null) 'path': path,
      if (bodyJson != null) 'body_json': bodyJson,
      if (idempotencyKey != null) 'idempotency_key': idempotencyKey,
      if (retryCount != null) 'retry_count': retryCount,
      if (createdAt != null) 'created_at': createdAt,
      if (nextAttemptAt != null) 'next_attempt_at': nextAttemptAt,
      if (lastError != null) 'last_error': lastError,
      if (rowid != null) 'rowid': rowid,
    });
  }

  SyncQueueCompanion copyWith({
    Value<String>? clientMutationId,
    Value<String>? method,
    Value<String>? path,
    Value<String?>? bodyJson,
    Value<String>? idempotencyKey,
    Value<int>? retryCount,
    Value<DateTime>? createdAt,
    Value<DateTime?>? nextAttemptAt,
    Value<String?>? lastError,
    Value<int>? rowid,
  }) {
    return SyncQueueCompanion(
      clientMutationId: clientMutationId ?? this.clientMutationId,
      method: method ?? this.method,
      path: path ?? this.path,
      bodyJson: bodyJson ?? this.bodyJson,
      idempotencyKey: idempotencyKey ?? this.idempotencyKey,
      retryCount: retryCount ?? this.retryCount,
      createdAt: createdAt ?? this.createdAt,
      nextAttemptAt: nextAttemptAt ?? this.nextAttemptAt,
      lastError: lastError ?? this.lastError,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (clientMutationId.present) {
      map['client_mutation_id'] = Variable<String>(clientMutationId.value);
    }
    if (method.present) {
      map['method'] = Variable<String>(method.value);
    }
    if (path.present) {
      map['path'] = Variable<String>(path.value);
    }
    if (bodyJson.present) {
      map['body_json'] = Variable<String>(bodyJson.value);
    }
    if (idempotencyKey.present) {
      map['idempotency_key'] = Variable<String>(idempotencyKey.value);
    }
    if (retryCount.present) {
      map['retry_count'] = Variable<int>(retryCount.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (nextAttemptAt.present) {
      map['next_attempt_at'] = Variable<DateTime>(nextAttemptAt.value);
    }
    if (lastError.present) {
      map['last_error'] = Variable<String>(lastError.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('SyncQueueCompanion(')
          ..write('clientMutationId: $clientMutationId, ')
          ..write('method: $method, ')
          ..write('path: $path, ')
          ..write('bodyJson: $bodyJson, ')
          ..write('idempotencyKey: $idempotencyKey, ')
          ..write('retryCount: $retryCount, ')
          ..write('createdAt: $createdAt, ')
          ..write('nextAttemptAt: $nextAttemptAt, ')
          ..write('lastError: $lastError, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

abstract class _$AppDatabase extends GeneratedDatabase {
  _$AppDatabase(QueryExecutor e) : super(e);
  $AppDatabaseManager get managers => $AppDatabaseManager(this);
  late final $CartItemsTable cartItems = $CartItemsTable(this);
  late final $FavoritesTable favorites = $FavoritesTable(this);
  late final $FavoriteEntriesTable favoriteEntries = $FavoriteEntriesTable(
    this,
  );
  late final $RecentSearchesTable recentSearches = $RecentSearchesTable(this);
  late final $CachedOrdersTable cachedOrders = $CachedOrdersTable(this);
  late final $SyncQueueTable syncQueue = $SyncQueueTable(this);
  @override
  Iterable<TableInfo<Table, Object?>> get allTables =>
      allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities => [
    cartItems,
    favorites,
    favoriteEntries,
    recentSearches,
    cachedOrders,
    syncQueue,
  ];
}

typedef $$CartItemsTableCreateCompanionBuilder =
    CartItemsCompanion Function({
      required String productId,
      Value<String?> variantId,
      required String title,
      Value<String?> imageUrl,
      Value<int> quantity,
      required int unitPriceMinor,
      Value<String> currency,
      Value<String?> notes,
      Value<DateTime> updatedAt,
      Value<bool> pendingSync,
      Value<int> rowid,
    });
typedef $$CartItemsTableUpdateCompanionBuilder =
    CartItemsCompanion Function({
      Value<String> productId,
      Value<String?> variantId,
      Value<String> title,
      Value<String?> imageUrl,
      Value<int> quantity,
      Value<int> unitPriceMinor,
      Value<String> currency,
      Value<String?> notes,
      Value<DateTime> updatedAt,
      Value<bool> pendingSync,
      Value<int> rowid,
    });

class $$CartItemsTableFilterComposer
    extends Composer<_$AppDatabase, $CartItemsTable> {
  $$CartItemsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get productId => $composableBuilder(
    column: $table.productId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get variantId => $composableBuilder(
    column: $table.variantId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get title => $composableBuilder(
    column: $table.title,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get imageUrl => $composableBuilder(
    column: $table.imageUrl,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get quantity => $composableBuilder(
    column: $table.quantity,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get unitPriceMinor => $composableBuilder(
    column: $table.unitPriceMinor,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get currency => $composableBuilder(
    column: $table.currency,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get notes => $composableBuilder(
    column: $table.notes,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get updatedAt => $composableBuilder(
    column: $table.updatedAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<bool> get pendingSync => $composableBuilder(
    column: $table.pendingSync,
    builder: (column) => ColumnFilters(column),
  );
}

class $$CartItemsTableOrderingComposer
    extends Composer<_$AppDatabase, $CartItemsTable> {
  $$CartItemsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get productId => $composableBuilder(
    column: $table.productId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get variantId => $composableBuilder(
    column: $table.variantId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get title => $composableBuilder(
    column: $table.title,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get imageUrl => $composableBuilder(
    column: $table.imageUrl,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get quantity => $composableBuilder(
    column: $table.quantity,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get unitPriceMinor => $composableBuilder(
    column: $table.unitPriceMinor,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get currency => $composableBuilder(
    column: $table.currency,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get notes => $composableBuilder(
    column: $table.notes,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get updatedAt => $composableBuilder(
    column: $table.updatedAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<bool> get pendingSync => $composableBuilder(
    column: $table.pendingSync,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$CartItemsTableAnnotationComposer
    extends Composer<_$AppDatabase, $CartItemsTable> {
  $$CartItemsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get productId =>
      $composableBuilder(column: $table.productId, builder: (column) => column);

  GeneratedColumn<String> get variantId =>
      $composableBuilder(column: $table.variantId, builder: (column) => column);

  GeneratedColumn<String> get title =>
      $composableBuilder(column: $table.title, builder: (column) => column);

  GeneratedColumn<String> get imageUrl =>
      $composableBuilder(column: $table.imageUrl, builder: (column) => column);

  GeneratedColumn<int> get quantity =>
      $composableBuilder(column: $table.quantity, builder: (column) => column);

  GeneratedColumn<int> get unitPriceMinor => $composableBuilder(
    column: $table.unitPriceMinor,
    builder: (column) => column,
  );

  GeneratedColumn<String> get currency =>
      $composableBuilder(column: $table.currency, builder: (column) => column);

  GeneratedColumn<String> get notes =>
      $composableBuilder(column: $table.notes, builder: (column) => column);

  GeneratedColumn<DateTime> get updatedAt =>
      $composableBuilder(column: $table.updatedAt, builder: (column) => column);

  GeneratedColumn<bool> get pendingSync => $composableBuilder(
    column: $table.pendingSync,
    builder: (column) => column,
  );
}

class $$CartItemsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $CartItemsTable,
          CartItem,
          $$CartItemsTableFilterComposer,
          $$CartItemsTableOrderingComposer,
          $$CartItemsTableAnnotationComposer,
          $$CartItemsTableCreateCompanionBuilder,
          $$CartItemsTableUpdateCompanionBuilder,
          (CartItem, BaseReferences<_$AppDatabase, $CartItemsTable, CartItem>),
          CartItem,
          PrefetchHooks Function()
        > {
  $$CartItemsTableTableManager(_$AppDatabase db, $CartItemsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$CartItemsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$CartItemsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$CartItemsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> productId = const Value.absent(),
                Value<String?> variantId = const Value.absent(),
                Value<String> title = const Value.absent(),
                Value<String?> imageUrl = const Value.absent(),
                Value<int> quantity = const Value.absent(),
                Value<int> unitPriceMinor = const Value.absent(),
                Value<String> currency = const Value.absent(),
                Value<String?> notes = const Value.absent(),
                Value<DateTime> updatedAt = const Value.absent(),
                Value<bool> pendingSync = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => CartItemsCompanion(
                productId: productId,
                variantId: variantId,
                title: title,
                imageUrl: imageUrl,
                quantity: quantity,
                unitPriceMinor: unitPriceMinor,
                currency: currency,
                notes: notes,
                updatedAt: updatedAt,
                pendingSync: pendingSync,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String productId,
                Value<String?> variantId = const Value.absent(),
                required String title,
                Value<String?> imageUrl = const Value.absent(),
                Value<int> quantity = const Value.absent(),
                required int unitPriceMinor,
                Value<String> currency = const Value.absent(),
                Value<String?> notes = const Value.absent(),
                Value<DateTime> updatedAt = const Value.absent(),
                Value<bool> pendingSync = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => CartItemsCompanion.insert(
                productId: productId,
                variantId: variantId,
                title: title,
                imageUrl: imageUrl,
                quantity: quantity,
                unitPriceMinor: unitPriceMinor,
                currency: currency,
                notes: notes,
                updatedAt: updatedAt,
                pendingSync: pendingSync,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$CartItemsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $CartItemsTable,
      CartItem,
      $$CartItemsTableFilterComposer,
      $$CartItemsTableOrderingComposer,
      $$CartItemsTableAnnotationComposer,
      $$CartItemsTableCreateCompanionBuilder,
      $$CartItemsTableUpdateCompanionBuilder,
      (CartItem, BaseReferences<_$AppDatabase, $CartItemsTable, CartItem>),
      CartItem,
      PrefetchHooks Function()
    >;
typedef $$FavoritesTableCreateCompanionBuilder =
    FavoritesCompanion Function({
      required String productId,
      required String title,
      Value<String?> imageUrl,
      Value<int?> unitPriceMinor,
      Value<DateTime> addedAt,
      Value<int> rowid,
    });
typedef $$FavoritesTableUpdateCompanionBuilder =
    FavoritesCompanion Function({
      Value<String> productId,
      Value<String> title,
      Value<String?> imageUrl,
      Value<int?> unitPriceMinor,
      Value<DateTime> addedAt,
      Value<int> rowid,
    });

class $$FavoritesTableFilterComposer
    extends Composer<_$AppDatabase, $FavoritesTable> {
  $$FavoritesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get productId => $composableBuilder(
    column: $table.productId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get title => $composableBuilder(
    column: $table.title,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get imageUrl => $composableBuilder(
    column: $table.imageUrl,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get unitPriceMinor => $composableBuilder(
    column: $table.unitPriceMinor,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get addedAt => $composableBuilder(
    column: $table.addedAt,
    builder: (column) => ColumnFilters(column),
  );
}

class $$FavoritesTableOrderingComposer
    extends Composer<_$AppDatabase, $FavoritesTable> {
  $$FavoritesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get productId => $composableBuilder(
    column: $table.productId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get title => $composableBuilder(
    column: $table.title,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get imageUrl => $composableBuilder(
    column: $table.imageUrl,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get unitPriceMinor => $composableBuilder(
    column: $table.unitPriceMinor,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get addedAt => $composableBuilder(
    column: $table.addedAt,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$FavoritesTableAnnotationComposer
    extends Composer<_$AppDatabase, $FavoritesTable> {
  $$FavoritesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get productId =>
      $composableBuilder(column: $table.productId, builder: (column) => column);

  GeneratedColumn<String> get title =>
      $composableBuilder(column: $table.title, builder: (column) => column);

  GeneratedColumn<String> get imageUrl =>
      $composableBuilder(column: $table.imageUrl, builder: (column) => column);

  GeneratedColumn<int> get unitPriceMinor => $composableBuilder(
    column: $table.unitPriceMinor,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get addedAt =>
      $composableBuilder(column: $table.addedAt, builder: (column) => column);
}

class $$FavoritesTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $FavoritesTable,
          Favorite,
          $$FavoritesTableFilterComposer,
          $$FavoritesTableOrderingComposer,
          $$FavoritesTableAnnotationComposer,
          $$FavoritesTableCreateCompanionBuilder,
          $$FavoritesTableUpdateCompanionBuilder,
          (Favorite, BaseReferences<_$AppDatabase, $FavoritesTable, Favorite>),
          Favorite,
          PrefetchHooks Function()
        > {
  $$FavoritesTableTableManager(_$AppDatabase db, $FavoritesTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$FavoritesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$FavoritesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$FavoritesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> productId = const Value.absent(),
                Value<String> title = const Value.absent(),
                Value<String?> imageUrl = const Value.absent(),
                Value<int?> unitPriceMinor = const Value.absent(),
                Value<DateTime> addedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => FavoritesCompanion(
                productId: productId,
                title: title,
                imageUrl: imageUrl,
                unitPriceMinor: unitPriceMinor,
                addedAt: addedAt,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String productId,
                required String title,
                Value<String?> imageUrl = const Value.absent(),
                Value<int?> unitPriceMinor = const Value.absent(),
                Value<DateTime> addedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => FavoritesCompanion.insert(
                productId: productId,
                title: title,
                imageUrl: imageUrl,
                unitPriceMinor: unitPriceMinor,
                addedAt: addedAt,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$FavoritesTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $FavoritesTable,
      Favorite,
      $$FavoritesTableFilterComposer,
      $$FavoritesTableOrderingComposer,
      $$FavoritesTableAnnotationComposer,
      $$FavoritesTableCreateCompanionBuilder,
      $$FavoritesTableUpdateCompanionBuilder,
      (Favorite, BaseReferences<_$AppDatabase, $FavoritesTable, Favorite>),
      Favorite,
      PrefetchHooks Function()
    >;
typedef $$FavoriteEntriesTableCreateCompanionBuilder =
    FavoriteEntriesCompanion Function({
      required String entryId,
      required String type,
      required String targetId,
      required String title,
      Value<String> subtitle,
      Value<String?> imageUrl,
      Value<String> metadataJson,
      Value<DateTime> addedAt,
      Value<bool> pendingSync,
      Value<int> rowid,
    });
typedef $$FavoriteEntriesTableUpdateCompanionBuilder =
    FavoriteEntriesCompanion Function({
      Value<String> entryId,
      Value<String> type,
      Value<String> targetId,
      Value<String> title,
      Value<String> subtitle,
      Value<String?> imageUrl,
      Value<String> metadataJson,
      Value<DateTime> addedAt,
      Value<bool> pendingSync,
      Value<int> rowid,
    });

class $$FavoriteEntriesTableFilterComposer
    extends Composer<_$AppDatabase, $FavoriteEntriesTable> {
  $$FavoriteEntriesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get entryId => $composableBuilder(
    column: $table.entryId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get type => $composableBuilder(
    column: $table.type,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get targetId => $composableBuilder(
    column: $table.targetId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get title => $composableBuilder(
    column: $table.title,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get subtitle => $composableBuilder(
    column: $table.subtitle,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get imageUrl => $composableBuilder(
    column: $table.imageUrl,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get metadataJson => $composableBuilder(
    column: $table.metadataJson,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get addedAt => $composableBuilder(
    column: $table.addedAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<bool> get pendingSync => $composableBuilder(
    column: $table.pendingSync,
    builder: (column) => ColumnFilters(column),
  );
}

class $$FavoriteEntriesTableOrderingComposer
    extends Composer<_$AppDatabase, $FavoriteEntriesTable> {
  $$FavoriteEntriesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get entryId => $composableBuilder(
    column: $table.entryId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get type => $composableBuilder(
    column: $table.type,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get targetId => $composableBuilder(
    column: $table.targetId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get title => $composableBuilder(
    column: $table.title,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get subtitle => $composableBuilder(
    column: $table.subtitle,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get imageUrl => $composableBuilder(
    column: $table.imageUrl,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get metadataJson => $composableBuilder(
    column: $table.metadataJson,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get addedAt => $composableBuilder(
    column: $table.addedAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<bool> get pendingSync => $composableBuilder(
    column: $table.pendingSync,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$FavoriteEntriesTableAnnotationComposer
    extends Composer<_$AppDatabase, $FavoriteEntriesTable> {
  $$FavoriteEntriesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get entryId =>
      $composableBuilder(column: $table.entryId, builder: (column) => column);

  GeneratedColumn<String> get type =>
      $composableBuilder(column: $table.type, builder: (column) => column);

  GeneratedColumn<String> get targetId =>
      $composableBuilder(column: $table.targetId, builder: (column) => column);

  GeneratedColumn<String> get title =>
      $composableBuilder(column: $table.title, builder: (column) => column);

  GeneratedColumn<String> get subtitle =>
      $composableBuilder(column: $table.subtitle, builder: (column) => column);

  GeneratedColumn<String> get imageUrl =>
      $composableBuilder(column: $table.imageUrl, builder: (column) => column);

  GeneratedColumn<String> get metadataJson => $composableBuilder(
    column: $table.metadataJson,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get addedAt =>
      $composableBuilder(column: $table.addedAt, builder: (column) => column);

  GeneratedColumn<bool> get pendingSync => $composableBuilder(
    column: $table.pendingSync,
    builder: (column) => column,
  );
}

class $$FavoriteEntriesTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $FavoriteEntriesTable,
          FavoriteEntryRow,
          $$FavoriteEntriesTableFilterComposer,
          $$FavoriteEntriesTableOrderingComposer,
          $$FavoriteEntriesTableAnnotationComposer,
          $$FavoriteEntriesTableCreateCompanionBuilder,
          $$FavoriteEntriesTableUpdateCompanionBuilder,
          (
            FavoriteEntryRow,
            BaseReferences<
              _$AppDatabase,
              $FavoriteEntriesTable,
              FavoriteEntryRow
            >,
          ),
          FavoriteEntryRow,
          PrefetchHooks Function()
        > {
  $$FavoriteEntriesTableTableManager(
    _$AppDatabase db,
    $FavoriteEntriesTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$FavoriteEntriesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$FavoriteEntriesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$FavoriteEntriesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> entryId = const Value.absent(),
                Value<String> type = const Value.absent(),
                Value<String> targetId = const Value.absent(),
                Value<String> title = const Value.absent(),
                Value<String> subtitle = const Value.absent(),
                Value<String?> imageUrl = const Value.absent(),
                Value<String> metadataJson = const Value.absent(),
                Value<DateTime> addedAt = const Value.absent(),
                Value<bool> pendingSync = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => FavoriteEntriesCompanion(
                entryId: entryId,
                type: type,
                targetId: targetId,
                title: title,
                subtitle: subtitle,
                imageUrl: imageUrl,
                metadataJson: metadataJson,
                addedAt: addedAt,
                pendingSync: pendingSync,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String entryId,
                required String type,
                required String targetId,
                required String title,
                Value<String> subtitle = const Value.absent(),
                Value<String?> imageUrl = const Value.absent(),
                Value<String> metadataJson = const Value.absent(),
                Value<DateTime> addedAt = const Value.absent(),
                Value<bool> pendingSync = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => FavoriteEntriesCompanion.insert(
                entryId: entryId,
                type: type,
                targetId: targetId,
                title: title,
                subtitle: subtitle,
                imageUrl: imageUrl,
                metadataJson: metadataJson,
                addedAt: addedAt,
                pendingSync: pendingSync,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$FavoriteEntriesTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $FavoriteEntriesTable,
      FavoriteEntryRow,
      $$FavoriteEntriesTableFilterComposer,
      $$FavoriteEntriesTableOrderingComposer,
      $$FavoriteEntriesTableAnnotationComposer,
      $$FavoriteEntriesTableCreateCompanionBuilder,
      $$FavoriteEntriesTableUpdateCompanionBuilder,
      (
        FavoriteEntryRow,
        BaseReferences<_$AppDatabase, $FavoriteEntriesTable, FavoriteEntryRow>,
      ),
      FavoriteEntryRow,
      PrefetchHooks Function()
    >;
typedef $$RecentSearchesTableCreateCompanionBuilder =
    RecentSearchesCompanion Function({
      Value<int> id,
      required String query,
      Value<DateTime> searchedAt,
    });
typedef $$RecentSearchesTableUpdateCompanionBuilder =
    RecentSearchesCompanion Function({
      Value<int> id,
      Value<String> query,
      Value<DateTime> searchedAt,
    });

class $$RecentSearchesTableFilterComposer
    extends Composer<_$AppDatabase, $RecentSearchesTable> {
  $$RecentSearchesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get query => $composableBuilder(
    column: $table.query,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get searchedAt => $composableBuilder(
    column: $table.searchedAt,
    builder: (column) => ColumnFilters(column),
  );
}

class $$RecentSearchesTableOrderingComposer
    extends Composer<_$AppDatabase, $RecentSearchesTable> {
  $$RecentSearchesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get query => $composableBuilder(
    column: $table.query,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get searchedAt => $composableBuilder(
    column: $table.searchedAt,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$RecentSearchesTableAnnotationComposer
    extends Composer<_$AppDatabase, $RecentSearchesTable> {
  $$RecentSearchesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get query =>
      $composableBuilder(column: $table.query, builder: (column) => column);

  GeneratedColumn<DateTime> get searchedAt => $composableBuilder(
    column: $table.searchedAt,
    builder: (column) => column,
  );
}

class $$RecentSearchesTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $RecentSearchesTable,
          RecentSearche,
          $$RecentSearchesTableFilterComposer,
          $$RecentSearchesTableOrderingComposer,
          $$RecentSearchesTableAnnotationComposer,
          $$RecentSearchesTableCreateCompanionBuilder,
          $$RecentSearchesTableUpdateCompanionBuilder,
          (
            RecentSearche,
            BaseReferences<_$AppDatabase, $RecentSearchesTable, RecentSearche>,
          ),
          RecentSearche,
          PrefetchHooks Function()
        > {
  $$RecentSearchesTableTableManager(
    _$AppDatabase db,
    $RecentSearchesTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$RecentSearchesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$RecentSearchesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$RecentSearchesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                Value<String> query = const Value.absent(),
                Value<DateTime> searchedAt = const Value.absent(),
              }) => RecentSearchesCompanion(
                id: id,
                query: query,
                searchedAt: searchedAt,
              ),
          createCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                required String query,
                Value<DateTime> searchedAt = const Value.absent(),
              }) => RecentSearchesCompanion.insert(
                id: id,
                query: query,
                searchedAt: searchedAt,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$RecentSearchesTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $RecentSearchesTable,
      RecentSearche,
      $$RecentSearchesTableFilterComposer,
      $$RecentSearchesTableOrderingComposer,
      $$RecentSearchesTableAnnotationComposer,
      $$RecentSearchesTableCreateCompanionBuilder,
      $$RecentSearchesTableUpdateCompanionBuilder,
      (
        RecentSearche,
        BaseReferences<_$AppDatabase, $RecentSearchesTable, RecentSearche>,
      ),
      RecentSearche,
      PrefetchHooks Function()
    >;
typedef $$CachedOrdersTableCreateCompanionBuilder =
    CachedOrdersCompanion Function({
      required String orderId,
      required String status,
      required String payloadJson,
      Value<DateTime> cachedAt,
      Value<int> rowid,
    });
typedef $$CachedOrdersTableUpdateCompanionBuilder =
    CachedOrdersCompanion Function({
      Value<String> orderId,
      Value<String> status,
      Value<String> payloadJson,
      Value<DateTime> cachedAt,
      Value<int> rowid,
    });

class $$CachedOrdersTableFilterComposer
    extends Composer<_$AppDatabase, $CachedOrdersTable> {
  $$CachedOrdersTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get orderId => $composableBuilder(
    column: $table.orderId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get status => $composableBuilder(
    column: $table.status,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get payloadJson => $composableBuilder(
    column: $table.payloadJson,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get cachedAt => $composableBuilder(
    column: $table.cachedAt,
    builder: (column) => ColumnFilters(column),
  );
}

class $$CachedOrdersTableOrderingComposer
    extends Composer<_$AppDatabase, $CachedOrdersTable> {
  $$CachedOrdersTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get orderId => $composableBuilder(
    column: $table.orderId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get status => $composableBuilder(
    column: $table.status,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get payloadJson => $composableBuilder(
    column: $table.payloadJson,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get cachedAt => $composableBuilder(
    column: $table.cachedAt,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$CachedOrdersTableAnnotationComposer
    extends Composer<_$AppDatabase, $CachedOrdersTable> {
  $$CachedOrdersTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get orderId =>
      $composableBuilder(column: $table.orderId, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<String> get payloadJson => $composableBuilder(
    column: $table.payloadJson,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get cachedAt =>
      $composableBuilder(column: $table.cachedAt, builder: (column) => column);
}

class $$CachedOrdersTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $CachedOrdersTable,
          CachedOrder,
          $$CachedOrdersTableFilterComposer,
          $$CachedOrdersTableOrderingComposer,
          $$CachedOrdersTableAnnotationComposer,
          $$CachedOrdersTableCreateCompanionBuilder,
          $$CachedOrdersTableUpdateCompanionBuilder,
          (
            CachedOrder,
            BaseReferences<_$AppDatabase, $CachedOrdersTable, CachedOrder>,
          ),
          CachedOrder,
          PrefetchHooks Function()
        > {
  $$CachedOrdersTableTableManager(_$AppDatabase db, $CachedOrdersTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$CachedOrdersTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$CachedOrdersTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$CachedOrdersTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> orderId = const Value.absent(),
                Value<String> status = const Value.absent(),
                Value<String> payloadJson = const Value.absent(),
                Value<DateTime> cachedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => CachedOrdersCompanion(
                orderId: orderId,
                status: status,
                payloadJson: payloadJson,
                cachedAt: cachedAt,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String orderId,
                required String status,
                required String payloadJson,
                Value<DateTime> cachedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => CachedOrdersCompanion.insert(
                orderId: orderId,
                status: status,
                payloadJson: payloadJson,
                cachedAt: cachedAt,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$CachedOrdersTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $CachedOrdersTable,
      CachedOrder,
      $$CachedOrdersTableFilterComposer,
      $$CachedOrdersTableOrderingComposer,
      $$CachedOrdersTableAnnotationComposer,
      $$CachedOrdersTableCreateCompanionBuilder,
      $$CachedOrdersTableUpdateCompanionBuilder,
      (
        CachedOrder,
        BaseReferences<_$AppDatabase, $CachedOrdersTable, CachedOrder>,
      ),
      CachedOrder,
      PrefetchHooks Function()
    >;
typedef $$SyncQueueTableCreateCompanionBuilder =
    SyncQueueCompanion Function({
      required String clientMutationId,
      required String method,
      required String path,
      Value<String?> bodyJson,
      required String idempotencyKey,
      Value<int> retryCount,
      Value<DateTime> createdAt,
      Value<DateTime?> nextAttemptAt,
      Value<String?> lastError,
      Value<int> rowid,
    });
typedef $$SyncQueueTableUpdateCompanionBuilder =
    SyncQueueCompanion Function({
      Value<String> clientMutationId,
      Value<String> method,
      Value<String> path,
      Value<String?> bodyJson,
      Value<String> idempotencyKey,
      Value<int> retryCount,
      Value<DateTime> createdAt,
      Value<DateTime?> nextAttemptAt,
      Value<String?> lastError,
      Value<int> rowid,
    });

class $$SyncQueueTableFilterComposer
    extends Composer<_$AppDatabase, $SyncQueueTable> {
  $$SyncQueueTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get clientMutationId => $composableBuilder(
    column: $table.clientMutationId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get method => $composableBuilder(
    column: $table.method,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get path => $composableBuilder(
    column: $table.path,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get bodyJson => $composableBuilder(
    column: $table.bodyJson,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get idempotencyKey => $composableBuilder(
    column: $table.idempotencyKey,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get retryCount => $composableBuilder(
    column: $table.retryCount,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
    column: $table.createdAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get nextAttemptAt => $composableBuilder(
    column: $table.nextAttemptAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get lastError => $composableBuilder(
    column: $table.lastError,
    builder: (column) => ColumnFilters(column),
  );
}

class $$SyncQueueTableOrderingComposer
    extends Composer<_$AppDatabase, $SyncQueueTable> {
  $$SyncQueueTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get clientMutationId => $composableBuilder(
    column: $table.clientMutationId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get method => $composableBuilder(
    column: $table.method,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get path => $composableBuilder(
    column: $table.path,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get bodyJson => $composableBuilder(
    column: $table.bodyJson,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get idempotencyKey => $composableBuilder(
    column: $table.idempotencyKey,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get retryCount => $composableBuilder(
    column: $table.retryCount,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
    column: $table.createdAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get nextAttemptAt => $composableBuilder(
    column: $table.nextAttemptAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get lastError => $composableBuilder(
    column: $table.lastError,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$SyncQueueTableAnnotationComposer
    extends Composer<_$AppDatabase, $SyncQueueTable> {
  $$SyncQueueTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get clientMutationId => $composableBuilder(
    column: $table.clientMutationId,
    builder: (column) => column,
  );

  GeneratedColumn<String> get method =>
      $composableBuilder(column: $table.method, builder: (column) => column);

  GeneratedColumn<String> get path =>
      $composableBuilder(column: $table.path, builder: (column) => column);

  GeneratedColumn<String> get bodyJson =>
      $composableBuilder(column: $table.bodyJson, builder: (column) => column);

  GeneratedColumn<String> get idempotencyKey => $composableBuilder(
    column: $table.idempotencyKey,
    builder: (column) => column,
  );

  GeneratedColumn<int> get retryCount => $composableBuilder(
    column: $table.retryCount,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<DateTime> get nextAttemptAt => $composableBuilder(
    column: $table.nextAttemptAt,
    builder: (column) => column,
  );

  GeneratedColumn<String> get lastError =>
      $composableBuilder(column: $table.lastError, builder: (column) => column);
}

class $$SyncQueueTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $SyncQueueTable,
          SyncQueueData,
          $$SyncQueueTableFilterComposer,
          $$SyncQueueTableOrderingComposer,
          $$SyncQueueTableAnnotationComposer,
          $$SyncQueueTableCreateCompanionBuilder,
          $$SyncQueueTableUpdateCompanionBuilder,
          (
            SyncQueueData,
            BaseReferences<_$AppDatabase, $SyncQueueTable, SyncQueueData>,
          ),
          SyncQueueData,
          PrefetchHooks Function()
        > {
  $$SyncQueueTableTableManager(_$AppDatabase db, $SyncQueueTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$SyncQueueTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$SyncQueueTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$SyncQueueTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> clientMutationId = const Value.absent(),
                Value<String> method = const Value.absent(),
                Value<String> path = const Value.absent(),
                Value<String?> bodyJson = const Value.absent(),
                Value<String> idempotencyKey = const Value.absent(),
                Value<int> retryCount = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<DateTime?> nextAttemptAt = const Value.absent(),
                Value<String?> lastError = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => SyncQueueCompanion(
                clientMutationId: clientMutationId,
                method: method,
                path: path,
                bodyJson: bodyJson,
                idempotencyKey: idempotencyKey,
                retryCount: retryCount,
                createdAt: createdAt,
                nextAttemptAt: nextAttemptAt,
                lastError: lastError,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String clientMutationId,
                required String method,
                required String path,
                Value<String?> bodyJson = const Value.absent(),
                required String idempotencyKey,
                Value<int> retryCount = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<DateTime?> nextAttemptAt = const Value.absent(),
                Value<String?> lastError = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => SyncQueueCompanion.insert(
                clientMutationId: clientMutationId,
                method: method,
                path: path,
                bodyJson: bodyJson,
                idempotencyKey: idempotencyKey,
                retryCount: retryCount,
                createdAt: createdAt,
                nextAttemptAt: nextAttemptAt,
                lastError: lastError,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$SyncQueueTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $SyncQueueTable,
      SyncQueueData,
      $$SyncQueueTableFilterComposer,
      $$SyncQueueTableOrderingComposer,
      $$SyncQueueTableAnnotationComposer,
      $$SyncQueueTableCreateCompanionBuilder,
      $$SyncQueueTableUpdateCompanionBuilder,
      (
        SyncQueueData,
        BaseReferences<_$AppDatabase, $SyncQueueTable, SyncQueueData>,
      ),
      SyncQueueData,
      PrefetchHooks Function()
    >;

class $AppDatabaseManager {
  final _$AppDatabase _db;
  $AppDatabaseManager(this._db);
  $$CartItemsTableTableManager get cartItems =>
      $$CartItemsTableTableManager(_db, _db.cartItems);
  $$FavoritesTableTableManager get favorites =>
      $$FavoritesTableTableManager(_db, _db.favorites);
  $$FavoriteEntriesTableTableManager get favoriteEntries =>
      $$FavoriteEntriesTableTableManager(_db, _db.favoriteEntries);
  $$RecentSearchesTableTableManager get recentSearches =>
      $$RecentSearchesTableTableManager(_db, _db.recentSearches);
  $$CachedOrdersTableTableManager get cachedOrders =>
      $$CachedOrdersTableTableManager(_db, _db.cachedOrders);
  $$SyncQueueTableTableManager get syncQueue =>
      $$SyncQueueTableTableManager(_db, _db.syncQueue);
}
