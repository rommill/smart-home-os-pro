export function generateSystemBadgesHtml(
  currentTemp: string | undefined | null,
  targetTemp: number,
  isOffline: boolean,
  currentLang: string,
): string {
  if (isOffline || !currentTemp) return "";

  const temp = parseFloat(currentTemp);
  if (temp >= targetTemp + 0.5) {
    return `
      <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-sky-500/10 text-sky-500 border border-sky-500/20 animate-pulse">
        ${currentLang === "ru" ? "ОХЛАЖДЕНИЕ" : "AC ACTIVE"}
      </span>`;
  }
  if (temp <= targetTemp - 0.5) {
    return `
      <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-amber-500/10 text-amber-500 border border-amber-500/20 animate-pulse">
        ${currentLang === "ru" ? "ОТОПЛЕНИЕ" : "HEATING"}
      </span>`;
  }

  return `
    <span class="inline-flex items-center text-[9px] font-bold text-emerald-500 dark:text-emerald-400 uppercase tracking-wider">
      • ${currentLang === "ru" ? "ИДЕАЛЬНО" : "STABLE"}
    </span>`;
}
