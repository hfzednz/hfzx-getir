"""Optional Python research helpers for innovation-service experiments."""

from __future__ import annotations


def score_hypothesis(accuracy: float, trl: int) -> float:
    """Simple research success score in 0..100."""
    return min(100.0, max(0.0, accuracy * 80.0 + trl * 2.0))
