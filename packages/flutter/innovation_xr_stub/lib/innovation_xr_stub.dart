library innovation_xr_stub;

/// Loads an XR asset URI registered in innovation-service.
class XRExperienceLoader {
  const XRExperienceLoader();

  Future<void> open(String assetUri) async {
    if (assetUri.isEmpty) {
      throw ArgumentError('assetUri required');
    }
  }
}
