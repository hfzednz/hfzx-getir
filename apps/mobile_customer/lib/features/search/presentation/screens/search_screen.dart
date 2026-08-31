import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:speech_to_text/speech_to_text.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../di/providers.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../domain/entities/search_entity.dart';
import '../../../product/domain/entities/product_entity.dart';
import '../providers/search_providers.dart';
import '../../../../shared/errors/error_copy.dart';
import '../../../../shared/widgets/error_view.dart';

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({super.key});

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen> {
  final _controller = TextEditingController();
  final _speech = SpeechToText();
  final _imagePicker = ImagePicker();

  String _query = '';
  bool _listening = false;
  bool _imageSearching = false;
  List<ProductSummary>? _imageResults;

  @override
  void dispose() {
    _controller.dispose();
    _speech.stop();
    super.dispose();
  }

  SearchQuery get _searchQuery => SearchQuery(
        text: _query,
        filters: ref.read(searchFiltersProvider),
      );

  void _applyQuery(String value, {bool persist = false}) {
    _controller.text = value;
    setState(() {
      _query = value;
      _imageResults = null;
    });
    if (persist && value.trim().isNotEmpty) {
      ref.read(databaseProvider).addRecentSearch(value.trim());
    }
  }

  void _submitQuery(String value) {
    final trimmed = value.trim();
    _applyQuery(trimmed, persist: trimmed.isNotEmpty);
  }

  void _showFiltersSheet() {
    final filters = ref.read(searchFiltersProvider);
    NxSheet.show<void>(
      context: context,
      child: _SearchFiltersSheet(
        initial: filters,
        onApply: (updated) {
          ref.read(searchFiltersProvider.notifier).state = updated;
          setState(() {});
          Navigator.of(context).pop();
        },
      ),
    );
  }

  Future<void> _startVoiceSearch() async {
    final l10n = AppLocalizations.of(context);
    final localeTag = Localizations.localeOf(context).toLanguageTag();
    final available = await _speech.initialize(
      onError: (error) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Voice search error: ${error.errorMsg}')),
        );
        setState(() => _listening = false);
      },
      onStatus: (status) {
        if (status == 'done' || status == 'notListening') {
          if (mounted) setState(() => _listening = false);
        }
      },
    );

    if (!available) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Microphone permission denied or speech recognition unavailable.',
          ),
        ),
      );
      return;
    }

    if (!mounted) return;
    setState(() => _listening = true);
    await _speech.listen(
      onResult: (result) {
        final text = result.recognizedWords.trim();
        if (text.isEmpty) return;
        _applyQuery(text);
        if (result.finalResult) {
          _speech.stop();
          if (mounted) setState(() => _listening = false);
        }
      },
      listenOptions: SpeechListenOptions(
        partialResults: true,
        listenFor: const Duration(seconds: 12),
        pauseFor: const Duration(seconds: 3),
        localeId: localeTag,
      ),
    );

    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('${l10n.voiceSearch}: listening…')),
    );
  }

  Future<void> _stopVoiceSearch() async {
    await _speech.stop();
    if (mounted) setState(() => _listening = false);
  }

  Future<void> _startImageSearch() async {
    final l10n = AppLocalizations.of(context);
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: Text(l10n.chooseFromGallery),
              onTap: () => Navigator.pop(ctx, ImageSource.gallery),
            ),
            ListTile(
              leading: const Icon(Icons.photo_camera_outlined),
              title: Text(l10n.takePhoto),
              onTap: () => Navigator.pop(ctx, ImageSource.camera),
            ),
          ],
        ),
      ),
    );
    if (source == null || !mounted) return;

    final file = await _imagePicker.pickImage(
      source: source,
      maxWidth: 1600,
      imageQuality: 85,
    );
    if (file == null || !mounted) return;

    setState(() => _imageSearching = true);
    try {
      final bytes = await file.readAsBytes();
      final result = await ref.read(searchRepositoryProvider).imageSearch(
            bytes,
            filename: file.name,
          );
      if (!mounted) return;

      result.fold(
        onSuccess: (searchResult) {
          final label = searchResult.query.trim().isNotEmpty
              ? searchResult.query
              : (searchResult.suggestions.isNotEmpty
                  ? searchResult.suggestions.first.text
                  : l10n.imageSearch);
          _controller.text = label;
          setState(() {
            _query = label;
            _imageResults = searchResult.items;
            _imageSearching = false;
          });
        },
        onFailure: (e) {
          setState(() => _imageSearching = false);
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(e.message)),
          );
        },
      );
    } catch (e) {
      if (!mounted) return;
      setState(() => _imageSearching = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Image search failed: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final results = ref.watch(searchResultsProvider(_searchQuery));
    final suggestions = ref.watch(searchSuggestionsProvider(_query));
    final showingImageResults = _imageResults != null;

    return Scaffold(
      appBar: NxTopBar(title: l10n.searchTitle),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Row(
              children: [
                Expanded(
                  child: NxSearchField(
                    controller: _controller,
                    hint: l10n.searchHint,
                    onChanged: (v) => setState(() {
                      _query = v;
                      _imageResults = null;
                    }),
                    onSubmitted: _submitQuery,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.tune),
                  tooltip: l10n.filters,
                  onPressed: _showFiltersSheet,
                ),
                IconButton(
                  icon: Icon(_listening ? Icons.mic : Icons.mic_none),
                  tooltip: l10n.voiceSearch,
                  color: _listening ? context.nxColors.textBrand : null,
                  onPressed: _listening ? _stopVoiceSearch : _startVoiceSearch,
                ),
                IconButton(
                  icon: _imageSearching
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: NxSpinner(),
                        )
                      : const Icon(Icons.image_search),
                  tooltip: l10n.imageSearch,
                  onPressed: _imageSearching ? null : _startImageSearch,
                ),
                IconButton(
                  icon: const Icon(Icons.qr_code_scanner),
                  tooltip: l10n.barcodeScan,
                  onPressed: () => context.push(RouteNames.barcodeScanner),
                ),
              ],
            ),
          ),
          if (!showingImageResults && _query.trim().isEmpty)
            Expanded(
              child: _RecentAndTrending(
                onSelect: (q) => _applyQuery(q, persist: true),
              ),
            )
          else if (!showingImageResults && _query.trim().length < 2)
            suggestions.when(
              data: (items) => items.isEmpty
                  ? const SizedBox.shrink()
                  : Expanded(
                      child: ListView(
                        children: items
                            .map(
                              (s) => ListTile(
                                leading: const Icon(Icons.search),
                                title: Text(s.text),
                                onTap: () => _applyQuery(s.text, persist: true),
                              ),
                            )
                            .toList(),
                      ),
                    ),
              loading: () => const SizedBox.shrink(),
              error: (_, __) => const SizedBox.shrink(),
            ),
          if (showingImageResults)
            Expanded(
              child: _ProductResultsList(items: _imageResults!),
            ),
          if (!showingImageResults && _query.trim().length >= 2)
            Expanded(
              child: results.when(
                data: (items) => _ProductResultsList(items: items),
                loading: () => const Center(child: NxSpinner()),
                error: (e, _) => ErrorView(message: localizedCustomerError(context, e)),
              ),
            ),
        ],
      ),
    );
  }
}

class _RecentAndTrending extends ConsumerWidget {
  const _RecentAndTrending({required this.onSelect});

  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final recentAsync = ref.watch(recentSearchesProvider);
    final trendingAsync = ref.watch(trendingSearchesProvider);

    return ListView(
      padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
      children: [
        Text('Recent', style: NxTypography.headlineSm),
        const SizedBox(height: NxSpacing.s2),
        recentAsync.when(
          data: (items) {
            if (items.isEmpty) {
              return Text(
                'No recent searches',
                style: NxTypography.captionMd.copyWith(
                  color: context.nxColors.textSecondary,
                ),
              );
            }
            return Column(
              children: items
                  .map(
                    (r) => ListTile(
                      contentPadding: EdgeInsets.zero,
                      leading: const Icon(Icons.history),
                      title: Text(r.query),
                      onTap: () => onSelect(r.query),
                    ),
                  )
                  .toList(),
            );
          },
          loading: () => const Center(child: NxSpinner()),
          error: (_, __) => const SizedBox.shrink(),
        ),
        const SizedBox(height: NxSpacing.s4),
        Text('Trending', style: NxTypography.headlineSm),
        const SizedBox(height: NxSpacing.s2),
        trendingAsync.when(
          data: (items) {
            if (items.isEmpty) {
              return Text(
                'No trending searches',
                style: NxTypography.captionMd.copyWith(
                  color: context.nxColors.textSecondary,
                ),
              );
            }
            return Column(
              children: items
                  .map(
                    (s) => ListTile(
                      contentPadding: EdgeInsets.zero,
                      leading: const Icon(Icons.trending_up),
                      title: Text(s.text),
                      onTap: () => onSelect(s.text),
                    ),
                  )
                  .toList(),
            );
          },
          loading: () => const Center(child: NxSpinner()),
          error: (_, __) => const SizedBox.shrink(),
        ),
      ],
    );
  }
}

class _ProductResultsList extends StatelessWidget {
  const _ProductResultsList({required this.items});

  final List<ProductSummary> items;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(child: Text('No products found'));
    }
    return ListView.builder(
      itemCount: items.length,
      itemBuilder: (context, i) {
        final item = items[i];
        final price = Money(minorUnits: item.priceMinor, currency: item.currency);
        return ListTile(
          leading: item.imageUrl != null
              ? Image.network(item.imageUrl!, width: 48, height: 48, fit: BoxFit.cover)
              : null,
          title: Text(item.title),
          subtitle: Text(Formatters.money(price)),
          onTap: () => context.push('/p/${item.id}'),
        );
      },
    );
  }
}

class _SearchFiltersSheet extends StatefulWidget {
  const _SearchFiltersSheet({required this.initial, required this.onApply});

  final SearchFilters initial;
  final ValueChanged<SearchFilters> onApply;

  @override
  State<_SearchFiltersSheet> createState() => _SearchFiltersSheetState();
}

class _SearchFiltersSheetState extends State<_SearchFiltersSheet> {
  late SearchSort _sort;
  ProductStockStatus? _availability;
  final _brandsController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _sort = widget.initial.sort;
    _availability = widget.initial.availability;
    _brandsController.text = widget.initial.brands.join(', ');
  }

  @override
  void dispose() {
    _brandsController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.all(NxSpacing.s4),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(l10n.filters, style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s4),
          NxField(label: l10n.brandsComma, controller: _brandsController),
          const SizedBox(height: NxSpacing.s3),
          DropdownButtonFormField<SearchSort>(
            value: _sort,
            decoration: InputDecoration(labelText: l10n.sort),
            items: SearchSort.values
                .map((s) => DropdownMenuItem(value: s, child: Text(searchSortToJson(s))))
                .toList(),
            onChanged: (v) => setState(() => _sort = v ?? SearchSort.relevance),
          ),
          const SizedBox(height: NxSpacing.s3),
          DropdownButtonFormField<ProductStockStatus?>(
            value: _availability,
            decoration: InputDecoration(labelText: l10n.availabilityFilter),
            items: [
              DropdownMenuItem(value: null, child: Text(l10n.anyAvailability)),
              DropdownMenuItem(value: ProductStockStatus.inStock, child: Text(l10n.inStock)),
              DropdownMenuItem(value: ProductStockStatus.low, child: Text(l10n.lowStock)),
            ],
            onChanged: (v) => setState(() => _availability = v),
          ),
          const SizedBox(height: NxSpacing.s6),
          NxButton(
            label: l10n.applyFilters,
            expand: true,
            onPressed: () {
              widget.onApply(
                widget.initial.copyWith(
                  brands: _brandsController.text
                      .split(',')
                      .map((e) => e.trim())
                      .where((e) => e.isNotEmpty)
                      .toList(),
                  sort: _sort,
                  availability: _availability,
                  clearAvailability: _availability == null,
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}
