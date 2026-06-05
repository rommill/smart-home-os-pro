/**
 * Climate Automation Utilities
 * Manages AC status checks and structural badge generation.
 */

export function getAcStatus(roomTemp, roomData) {
  // If backend provides a state, use it. Otherwise fall back to auto-on if temp >= 25.0°C
  if (roomData.ac_status !== undefined && roomData.ac_status !== null) {
    return roomData.ac_status;
  }
  return parseFloat(roomTemp) >= 25.0;
}

export function generateAcBadgeHtml(isAcOn) {
  const activeClass =
    "bg-sky-500/10 text-sky-500 animate-pulse border border-sky-500/20";
  const inactiveClass =
    "bg-slate-100 dark:bg-slate-800 text-slate-400 border border-transparent";

  return `
    <span class="ac-badge flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-bold uppercase tracking-wider ${isAcOn ? activeClass : inactiveClass}">
        AC: ${isAcOn ? "ON ❄️" : "OFF"}
    </span>
  `;
}
