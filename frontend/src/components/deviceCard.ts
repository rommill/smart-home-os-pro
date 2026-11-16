import { t } from "../i18n/i18n";

// Порог потери связи: 90 секунд (3 пропущенных интервала по 30 сек)
const OFFLINE_THRESHOLD_MS = 90 * 1000;

export interface DeviceData {
  device_id: string;
  temperature: number;
  humidity: number;
  battery: number;
  sensor_type: string;
  rssi: number;
  last_update: string;
}

/**
 * Проверяет, считается ли датчик отвалившимся по времени последнего пакета
 */
export function isDeviceOffline(lastUpdate?: string): boolean {
  if (!lastUpdate) return true;
  const lastTime = new Date(lastUpdate).getTime();
  if (isNaN(lastTime)) return true;
  return Date.now() - lastTime > OFFLINE_THRESHOLD_MS;
}

/**
 * Рендерит чистую карточку физического BLE-датчика с поддержкой мягкого увядания.
 */
export function createDeviceCardHtml(
  device: DeviceData,
  currentLang: string,
): string {
  const offline = isDeviceOffline(device.last_update);

  const time = device.last_update
    ? new Date(device.last_update).toLocaleTimeString(
        currentLang === "ru" ? "ru-RU" : "et-EE",
      )
    : t("noData");

  // Динамические стили для активного и увядшего (оффлайн) состояний
  const cardStateClasses = offline
    ? "opacity-45 grayscale border-slate-700/60"
    : "opacity-100 border-emerald-500/40";

  const titleColor = offline ? "text-slate-400" : "text-emerald-400";

  const badgeHtml = offline
    ? `<span class="px-2 py-0.5 text-[10px] font-bold bg-amber-500/20 text-amber-300 rounded-full">⚠️ OFFLINE</span>`
    : `<span class="px-2 py-0.5 text-[10px] font-bold bg-emerald-500/20 text-emerald-300 rounded-full">LIVE / HARDWARE</span>`;

  const statusFooterHtml = offline
    ? `<span class="text-amber-400 font-semibold">⚠️ Нет связи</span>`
    : `<span class="text-emerald-400 font-semibold">● Online</span>`;

  return `
    <div class="device-card bg-slate-900/80 border p-6 rounded-2xl shadow-lg backdrop-blur-xl transition-all duration-700 ${cardStateClasses}">
        <div class="flex justify-between items-start mb-4">
            <div>
                <h2 class="text-lg font-bold ${titleColor} transition-colors duration-500">Xiaomi BLE Sensor</h2>
                <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[10px] uppercase text-slate-400 font-semibold">MAC: ${device.device_id}</span>
                    ${badgeHtml}
                </div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-800 ${titleColor} text-xs font-mono">
                📶 RSSI: ${offline ? "—" : `${device.rssi} dBm`}
            </div>
        </div>
        <div class="flex items-end justify-between mb-4">
            <div class="flex items-baseline gap-1">
                <span class="text-4xl font-extrabold font-mono ${titleColor} transition-colors duration-500">${device.temperature.toFixed(1)}</span>
                <span class="text-xl font-bold text-slate-400">°C</span>
            </div>
            <div class="flex flex-col items-end">
                <span class="text-[10px] text-slate-400 font-bold uppercase">Humidity: ${device.humidity.toFixed(0)}%</span>
                <span class="text-xs font-mono text-slate-400">${time}</span>
            </div>
        </div>
        <div class="flex items-center justify-between text-xs text-slate-400 pt-2 border-t border-slate-800">
            <span>Battery: ${device.battery}%</span>
            ${statusFooterHtml}
        </div>
    </div>
  `;
}
