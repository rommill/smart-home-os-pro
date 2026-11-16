import { t } from "../i18n/i18n.js";

export function renderClimateControlPanel(
  roomId: number,
  targetTemp: number,
): string {
  const labelText: string = t("targetTemp");

  return `
    <div class="climate-control overflow-hidden transition-all duration-500 ease-in-out border-t border-slate-100 dark:border-slate-800/50 mt-2">
        <div class="pt-3">
            <div class="flex justify-between items-center mb-1.5">
                <span data-i18n="targetTemp" class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">${labelText}</span>
                <span class="text-sm font-mono font-bold text-sky-500 dark:text-sky-400" id="target-val-${roomId}">${targetTemp}°C</span>
            </div>
            <input 
                type="range" 
                min="18" 
                max="28" 
                step="0.5" 
                value="${targetTemp}" 
                class="w-full h-2 bg-slate-200 dark:bg-slate-800 rounded-lg appearance-none cursor-pointer accent-sky-500"
                onclick="event.stopPropagation();" 
                oninput="document.getElementById('target-val-${roomId}').innerText = this.value + '°C'; this.setAttribute('value', this.value);"
                data-room-id="${roomId}"
            />
            <div class="flex justify-between text-[10px] text-slate-400 font-mono mt-1">
                <span>18°C</span>
                <span>23°C</span>
                <span>28°C</span>
            </div>
        </div>
    </div>
  `;
}
