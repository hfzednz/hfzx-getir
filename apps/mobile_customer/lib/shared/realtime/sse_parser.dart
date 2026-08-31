/// Parses W3C Server-Sent Events frames from a text buffer.
class SseFrame {
  const SseFrame({this.id, this.event, this.data = ''});

  final String? id;
  final String? event;
  final String data;

  bool get isHeartbeat => data.isEmpty && event == null;
}

class SseParseResult {
  const SseParseResult({required this.frames, required this.rest});

  final List<SseFrame> frames;
  final String rest;
}

/// Splits complete SSE events (blank-line delimited) from [buffer].
SseParseResult parseSseBuffer(String buffer) {
  final frames = <SseFrame>[];
  var rest = buffer.replaceAll('\r\n', '\n');
  while (true) {
    final split = rest.indexOf('\n\n');
    if (split < 0) break;
    final block = rest.substring(0, split);
    rest = rest.substring(split + 2);
    final frame = _parseBlock(block);
    if (frame != null) frames.add(frame);
  }
  return SseParseResult(frames: frames, rest: rest);
}

SseFrame? _parseBlock(String block) {
  if (block.trim().isEmpty) return null;
  String? id;
  String? event;
  final data = StringBuffer();
  for (final raw in block.split('\n')) {
    final line = raw.trimRight();
    if (line.isEmpty || line.startsWith(':')) continue;
    final colon = line.indexOf(':');
    final field = colon < 0 ? line : line.substring(0, colon);
    var value = colon < 0 ? '' : line.substring(colon + 1);
    if (value.startsWith(' ')) value = value.substring(1);
    switch (field) {
      case 'id':
        id = value;
      case 'event':
        event = value;
      case 'data':
        if (data.isNotEmpty) data.write('\n');
        data.write(value);
    }
  }
  if (id == null && event == null && data.isEmpty) return null;
  return SseFrame(id: id, event: event, data: data.toString());
}
