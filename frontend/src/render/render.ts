import { t, getCurrentLang } from "../i18n/i18n";
import { RoomData } from "../types/telemetry";
import { createRoomCardHtml } from "./cardTemplate";
import { createDeviceCardHtml, DeviceData } from "../components/deviceCard";
import { toggleAuthView, updateStatus } from "./utils";

export { toggleAuthView, updateStatus };

interface DashboardPayload {
  rooms: RoomData[];
  devices: DeviceData[];
}

// Кэш устройств для фоновой проверки таймаутов
let lastKnownDevices: DeviceData[] = [];

/**
 * Вспомогательная функция для обновления карточек физических датчиков в DOM
 */
function updateDeviceCards(
  grid: HTMLDivElement,
  devices: DeviceData[],
  currentLang: string,
): Set<string> {
  const deviceCardIds = new Set<string>();

  devices.forEach((device) => {
    if (!device || !device.device_id) return;

    const safeDeviceId = device.device_id.replace(/:/g, "-");
    const cardId = `device-card-${safeDeviceId}`;
    deviceCardIds.add(cardId);
    let existingCard = document.getElementById(cardId);

    const tempContainer = document.createElement("div");
    tempContainer.innerHTML = createDeviceCardHtml(device, currentLang).trim();
    const newCardElement = tempContainer.firstElementChild as HTMLElement;

    if (existingCard && newCardElement) {
      // Синхронизируем CSS-классы для плавного перехода opacity/grayscale
      if (existingCard.className !== newCardElement.className) {
        existingCard.className = newCardElement.className;
      }
      if (existingCard.innerHTML !== newCardElement.innerHTML) {
        existingCard.innerHTML = newCardElement.innerHTML;
      }
    } else if (newCardElement) {
      newCardElement.id = cardId;
      grid.appendChild(newCardElement);
    }
  });

  return deviceCardIds;
}

/**
 * Рендерит карточки комнат и физических устройств с абсолютной целостностью элементов.
 */
export function renderRooms(payload: DashboardPayload | RoomData[]): void {
  const grid = document.getElementById("rooms-grid") as HTMLDivElement | null;
  if (!grid) return;

  let rooms: RoomData[] = [];
  let devices: DeviceData[] = [];

  if (Array.isArray(payload)) {
    rooms = payload;
  } else if (payload && typeof payload === "object") {
    rooms = payload.rooms || [];
    devices = payload.devices || [];
  }

  // Сохраняем в кэш для фонового таймера
  lastKnownDevices = devices;

  if (rooms.length === 0 && devices.length === 0) {
    grid.innerHTML = `<div class="col-span-full text-center text-slate-500 py-10">${t("noRooms")}</div>`;
    return;
  }

  if (grid.querySelector(".col-span-full")) {
    grid.innerHTML = "";
  }

  const currentLang = getCurrentLang();
  const existingCardIds = new Set<string>();

  // 1. Рендерим виртуальные комнаты (с ползунками из PostgreSQL)
  rooms.forEach((room) => {
    if (!room) return;
    const numericId = Number(room.id);
    if (isNaN(numericId) || numericId <= 0) return;

    const cardId = `room-card-${numericId}`;
    existingCardIds.add(cardId);
    let existingCard = document.getElementById(cardId);

    const tempContainer = document.createElement("div");
    tempContainer.innerHTML = createRoomCardHtml(
      { ...room, id: numericId },
      currentLang,
      false,
    ).trim();
    const newCardElement = tempContainer.firstElementChild as HTMLElement;

    if (existingCard && newCardElement) {
      if (existingCard.innerHTML !== newCardElement.innerHTML) {
        existingCard.innerHTML = newCardElement.innerHTML;
      }
    } else if (newCardElement) {
      newCardElement.id = cardId;
      grid.appendChild(newCardElement);
    }
  });

  // 2. Рендерим физические датчики через вспомогательную функцию
  const deviceCardIds = updateDeviceCards(grid, devices, currentLang);
  deviceCardIds.forEach((id) => existingCardIds.add(id));

  // 3. Автоматически удаляем карточки, которых больше нет в актуальных данных
  Array.from(grid.children).forEach((child) => {
    if (
      child.id &&
      (child.id.startsWith("room-card-") ||
        child.id.startsWith("device-card-")) &&
      !existingCardIds.has(child.id)
    ) {
      grid.removeChild(child);
    }
  });
}

// Фоновый тикер каждые 10 секунд: перепроверяет таймстауты и гасит карточку без новых пакетов от сервера
setInterval(() => {
  const grid = document.getElementById("rooms-grid") as HTMLDivElement | null;
  if (!grid || lastKnownDevices.length === 0) return;
  updateDeviceCards(grid, lastKnownDevices, getCurrentLang());
}, 10_000);
