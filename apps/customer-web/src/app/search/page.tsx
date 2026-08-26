"use client";

import { Suspense } from "react";
import SearchInner from "./search-inner";

export default function SearchPage() {
  return (
    <Suspense fallback={<p>Loading search…</p>}>
      <SearchInner />
    </Suspense>
  );
}
