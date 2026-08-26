import { t } from "../i18n/i18n";
import { RoomData } from "../types/telemetry";
import { renderClimateControlPanel } from "./climate";
import { generateSystemBadgesHtml } from "./badges";
import { getLocalizedRoomName } from "./helpers";
import { getRoomIcon, getTemperatureStyles } from "./utils";

export function createRoomCardHtml(
  room: RoomData,
  currentLang: string,
  isFirstRender: boolean,
): string {
  const roomName = getLocalizedRoomName(room, currentLang);
  const isOffline = !room.temperature || room.temperature === "N/A";
  const tempValue = isOffline ? "N/A" : room.temperature;

  const time =
    room.last_update && room.last_update !== "0001-01-01T00:00:00Z"
      ? new Date(room.last_update).toLocaleTimeString(
          currentLang === "ru" ? "ru-RU" : "et-EE",
        )
      : t("noData");

  const targetTemp = room.target_temperature ?? 23.0;
  const styles = getTemperatureStyles(room.temperature, isOffline);

  return `
    <div class="room-card ${isFirstRender ? "opacity-0" : "opacity-100"} bg-white dark:bg-slate-900/60 p-6 rounded-2xl shadow-md border border-slate-200 dark:border-slate-800/80 hover:border-emerald-500 transition-all duration-300 backdrop-blur-xl group cursor-pointer" data-room-id="${room.id}">
        <div class="flex justify-between items-start mb-4">
            <div>
                <h2 class="room-title-text text-lg font-bold text-slate-800 dark:text-slate-100">${roomName}</h2>
                <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[10px] uppercase text-slate-400 font-semibold">ID: 00${room.id}</span>
                    <span class="ac-badge-container">${generateSystemBadgesHtml(room.temperature, targetTemp, isOffline, currentLang)}</span>
                </div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-100 dark:bg-slate-800/80 text-slate-500">
                ${getRoomIcon(room.id)}
            </div>
        </div>
        <div class="flex items-end justify-between mb-4">
            <div class="flex items-baseline gap-1">
                <span class="temp-value-display text-4xl font-extrabold font-mono ${styles.colorClass}">${tempValue}</span>
                ${isOffline ? "" : `<span class="temp-unit text-xl font-bold text-slate-500">°C</span>`}
            </div>
            <div class="flex flex-col items-end">
                <span class="static-updated-label text-[10px] text-slate-600 font-bold uppercase">${t("updated")}</span>
                <span class="time-display text-xs font-mono font-semibold text-slate-700 dark:text-slate-400">${time}</span>
            </div>
        </div>
        <div class="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden mb-4">
            <div class="progress-bar h-full bg-gradient-to-r ${styles.barColor} transition-all duration-1000" style="width: ${styles.percent}%"></div>
        </div>
        <div class="climate-panel-container">
            ${renderClimateControlPanel(room.id, targetTemp)}
        </div>
    </div>
  `;
}
