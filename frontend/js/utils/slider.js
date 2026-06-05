import { t } from "../i18n.js";

/**
 * Climate Control Slider Component
 * Generates interactive target temperature controls for a specific room.
 * * @param {number} roomId - Unique identifier of the room
 * @param {number} targetTemp - Target temperature from DB or state
 * @returns {string} HTML component string
 */
export function renderClimateControlPanel(roomId, targetTemp) {
  // Fetch translated string dynamically during HTML generation
  const labelText = t("targetTemp");

  return `
    <div class="climate-control max-h-0 group-[.is-active]:max-h-40 overflow-hidden transition-all duration-500 ease-in-out border-t border-transparent group-[.is-active]:border-slate-100 group-[.is-active]:dark:border-slate-800/50 group-[.is-active]:mt-2">
        <div class="pt-4">
            <div class="flex justify-between items-center mb-2">
                <span class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">${labelText}</span>
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
                onchange="event.stopPropagation(); window.updateTargetTemperature(${roomId}, this.value);"
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
