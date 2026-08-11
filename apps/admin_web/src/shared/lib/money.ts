/**
 * Format minor currency units (e.g. kuruş / cents) for display.
 */
export function formatMinorUnits(
  amount: number,
  currency: string,
  locale = "tr-TR",
): string {
  const major = amount / 100;
  try {
    return new Intl.NumberFormat(locale, {
      style: "currency",
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(major);
  } catch {
    return `${major.toFixed(2)} ${currency}`;
  }
}
