# Mobile (Flutter) Optimization Guide

- Reduce rebuilds: const constructors, selective listenable
- Deferred loading for mini-apps (superapp_shell)
- Cold start: defer non-critical plugins
- Offline: mutation outbox already in nexora_core — keep
- Battery: batch location/telemetry; avoid tight timers
- Accessibility: WCAG contrast/touch targets per design system
