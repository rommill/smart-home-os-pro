import { t, getCurrentLang } from "./i18n.js";
import { animateCards } from "./theme.js";
import { getRoomIcon } from "./utils/icons.js";
import { getTemperatureStyles } from "./utils/weather.js";
import { renderClimateControlPanel } from "./utils/slider.js";

export { toggleAuthView } from "./utils/authUi.js";
export { updateStatus } from "./utils/statusUi.js";

/**
 * Generates climate system operational state badges (AC / Heater / Idle)
 * Based on Variant 2: Ideal Thermostat Balancing Engine
 */
function generateSystemBadgesHtml(
  currentTemp,
  targetTemp,
  isOffline,
  currentLang,
) {
  if (isOffline) return "";

  const temp = parseFloat(currentTemp);
  const target = parseFloat(targetTemp);
  let badges = "";

  // 1. Air Conditioning: Actual temperature is higher than target + 0.5°C
  if (temp >= target + 0.5) {
    badges += `
      <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-sky-500/10 text-sky-500 border border-sky-500/20 animate-pulse">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707M16.243 17.657l.707.707M6.343 6.343l.707-.707M14 12a2 2 0 11-4 0 2 2 0 014 0z"/></svg>
        ${currentLang === "ru" ? "ОХЛАЖДЕНИЕ" : "AC ACTIVE"}
      </span>
    `;
  }
  // 2. Heating Subsystem: Actual temperature is lower than target - 0.5°C
  else if (temp <= target - 0.5) {
    badges += `
      <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-bold bg-amber-500/10 text-amber-500 border border-amber-500/20 animate-pulse">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A8 8 0 0117.657 18.657z"/></svg>
        ${currentLang === "ru" ? "ОТОПЛЕНИЕ" : "HEATING"}
      </span>
    `;
  }

  return (
    badges ||
    `
    <span class="inline-flex items-center text-[9px] font-bold text-emerald-500 dark:text-emerald-400 uppercase tracking-wider">
      • ${currentLang === "ru" ? "ИДЕАЛЬНО" : "STABLE"}
    </span>
  `
  );
} // <-- ВОТ ЭТА СКОБКА БЫЛА ПРОПУЩЕНА! Теперь синтаксис в порядке.

/**
 * Extracts and returns localized room string signatures
 */
function getLocalizedRoomName(room, currentLang) {
  if (room.name && typeof room.name === "object") {
    return room.name[currentLang] || room.name["en"] || "Unknown Room";
  }
  return room.name || "Unknown Room";
}

/**
 * HTML Template Generator for individual room cards
 */
function createRoomCardHtml(room, currentLang, isFirstRender) {
  const roomName = getLocalizedRoomName(room, currentLang);
  const isOffline =
    room.temperature === undefined ||
    room.temperature === null ||
    room.temperature === "N/A";
  const tempValue = isOffline ? "N/A" : room.temperature;

  const time =
    room.last_update && room.last_update !== "0001-01-01T00:00:00Z"
      ? new Date(room.last_update).toLocaleTimeString(
          currentLang === "ru" ? "ru-RU" : "et-EE",
        )
      : t("noData");

  const targetTemp =
    room.target_temperature !== undefined
      ? room.target_temperature
      : room.TargetTemperature || 23.0;

  let styles = getTemperatureStyles(room.temperature, isOffline);
  if (!isOffline) {
    const currentFloat = parseFloat(room.temperature);
    const targetFloat = parseFloat(targetTemp);
    if (currentFloat >= targetFloat + 0.5) {
      styles.barColor = "from-sky-400 to-blue-500";
    } else if (currentFloat <= targetFloat - 0.5) {
      styles.barColor = "from-orange-400 to-red-500";
    }
  }

  const cardOpacity = isFirstRender ? "opacity-0" : "opacity-100";
  const cardClass = `room-card ${cardOpacity} bg-white dark:bg-slate-900/60 p-6 rounded-2xl shadow-md dark:shadow-2xl border border-slate-200 dark:border-slate-800/80 hover:border-emerald-500 dark:hover:border-emerald-500/50 transition-all duration-300 backdrop-blur-xl group cursor-pointer`;

  const unitColorClass = styles.colorClass
    ? styles.colorClass.split(" ")[0] + "/70"
    : "text-slate-500";

  return `
    <div class="${cardClass}" data-room-id="${room.id}" onclick="this.classList.toggle('is-active')">
        <div class="flex justify-between items-start mb-4">
            <div>
                <h2 class="room-title-text text-lg font-bold text-slate-800 dark:text-slate-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors" data-raw-name='${JSON.stringify(room.name)}'>${roomName}</h2>
                <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[10px] uppercase tracking-wider text-slate-400 dark:text-slate-500 font-semibold">ID: 00${room.id}</span>
                    <span class="ac-badge-container">${generateSystemBadgesHtml(room.temperature, targetTemp, isOffline, currentLang)}</span>
                </div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-100 dark:bg-slate-800/80 text-slate-500 dark:text-slate-400 group-hover:bg-emerald-500/10 group-hover:text-emerald-500 transition-all duration-300">
                ${getRoomIcon(room.id)}
            </div>
        </div>

        <div class="flex items-end justify-between mb-4">
            <div class="flex items-baseline gap-1">
                <span class="temp-value-display text-4xl font-extrabold font-mono tracking-tight ${styles.colorClass || ""}">${tempValue}</span>
                ${isOffline ? "" : `<span class="temp-unit text-xl font-bold ${unitColorClass}">°C</span>`}
            </div>
            <div class="flex flex-col items-end">
                <span class="static-updated-label text-[10px] text-slate-600 dark:text-slate-400/80 font-bold uppercase tracking-wider">${t("updated")}</span>
                <span class="time-display text-xs font-mono font-semibold text-slate-700 dark:text-slate-400">${time}</span>
            </div>
        </div>

        <div class="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden mb-4">
            <div class="progress-bar h-full bg-gradient-to-r ${styles.barColor} transition-all duration-1000 ease-out" style="width: ${styles.percent}%"></div>
        </div>

        <div class="climate-panel-container">
            ${renderClimateControlPanel(room.id, targetTemp)}
        </div>
    </div>
  `;
}

/**
 * Smart Grid Rendering Process
 */
export function renderRooms(data) {
  const grid = document.getElementById("rooms-grid");
  if (!grid) return;

  if (!data || !Array.isArray(data) || data.length === 0) {
    grid.innerHTML = `<div class="col-span-full text-center text-slate-500 py-10">${t("noRooms")}</div>`;
    return;
  }

  const currentLang = getCurrentLang();
  const existingCards = grid.querySelectorAll(".room-card");

  if (existingCards.length === 0) {
    grid.innerHTML = data
      .map((room) => createRoomCardHtml(room, currentLang, true))
      .join("");
    animateCards();
    return;
  }

  data.forEach((room) => {
    const card = grid.querySelector(`.room-card[data-room-id="${room.id}"]`);
    if (!card) return;

    const isOffline =
      room.temperature === undefined ||
      room.temperature === null ||
      room.temperature === "N/A";
    const tempValue = isOffline ? "N/A" : room.temperature;
    const targetTemp =
      room.target_temperature !== undefined
        ? room.target_temperature
        : room.TargetTemperature || 23.0;

    let styles = getTemperatureStyles(room.temperature, isOffline);
    if (!isOffline) {
      const currentFloat = parseFloat(room.temperature);
      const targetFloat = parseFloat(targetTemp);
      if (currentFloat >= targetFloat + 0.5) {
        styles.barColor = "from-sky-400 to-blue-500";
      } else if (currentFloat <= targetFloat - 0.5) {
        styles.barColor = "from-orange-400 to-red-500";
      }
    }

    const time =
      room.last_update && room.last_update !== "0001-01-01T00:00:00Z"
        ? new Date(room.last_update).toLocaleTimeString(
            currentLang === "ru" ? "ru-RU" : "et-EE",
          )
        : t("noData");

    const titleTextNode = card.querySelector(".room-title-text");
    if (titleTextNode && titleTextNode.getAttribute("data-raw-name")) {
      try {
        const rawAttr = titleTextNode.getAttribute("data-raw-name");
        if (rawAttr.startsWith("{")) {
          const rawNameMap = JSON.parse(rawAttr);
          titleTextNode.textContent =
            rawNameMap[currentLang] ||
            rawNameMap["en"] ||
            titleTextNode.textContent;
        } else {
          titleTextNode.textContent = rawAttr.replace(/['"]+/g, "");
        }
      } catch (e) {}
    }

    const updatedStaticLabel = card.querySelector(".static-updated-label");
    if (updatedStaticLabel) updatedStaticLabel.textContent = t("updated");

    const tempDisplay = card.querySelector(".temp-value-display");
    if (tempDisplay && tempDisplay.textContent !== tempValue) {
      tempDisplay.textContent = tempValue;
      tempDisplay.className = `temp-value-display text-4xl font-extrabold font-mono tracking-tight ${styles.colorClass || ""}`;
    }

    const timeDisplay = card.querySelector(".time-display");
    if (timeDisplay) timeDisplay.textContent = time;

    const progressBar = card.querySelector(".progress-bar");
    if (progressBar) {
      progressBar.style.width = `${styles.percent}%`;
      progressBar.className = `progress-bar h-full bg-gradient-to-r ${styles.barColor} transition-all duration-1000 ease-out`;
    }

    const acContainer = card.querySelector(".ac-badge-container");
    if (acContainer)
      acContainer.innerHTML = generateSystemBadgesHtml(
        room.temperature,
        targetTemp,
        isOffline,
        currentLang,
      );

    if (!card.classList.contains("is-active")) {
      const targetTextNode = card.querySelector(`#target-val-${room.id}`);
      const inputSlider = card.querySelector(`input[type="range"]`);

      if (targetTextNode) targetTextNode.textContent = `${targetTemp}°C`;
      if (inputSlider) {
        inputSlider.value = targetTemp;
        inputSlider.setAttribute("value", targetTemp);
      }
    }
  });
}
