import { t, getCurrentLang } from "../i18n/i18n";
import { RoomData } from "../types/telemetry";
import { createRoomCardHtml } from "./cardTemplate";
import { toggleAuthView, updateStatus } from "./utils";

export { toggleAuthView, updateStatus };

/**
 * Renders room cards cleanly with absolute strict single-element integrity.
 */
export function renderRooms(rawData: RoomData[]): void {
  const grid = document.getElementById("rooms-grid") as HTMLDivElement | null;
  if (!grid) return;

  if (!rawData || !Array.isArray(rawData) || rawData.length === 0) {
    grid.innerHTML = `<div class="col-span-full text-center text-slate-500 py-10">${t("noRooms")}</div>`;
    return;
  }

  const uniqueRoomsMap = new Map<number, RoomData>();
  rawData.forEach((room) => {
    if (room) {
      const numericId = Number(room.id);
      if (!isNaN(numericId) && numericId > 0) {
        uniqueRoomsMap.set(numericId, { ...room, id: numericId });
      }
    }
  });

  const rooms = Array.from(uniqueRoomsMap.values());
  const currentLang = getCurrentLang();

  if (grid.querySelector(".col-span-full")) {
    grid.innerHTML = "";
  }

  const existingCardIds = new Set<string>();

  rooms.forEach((room) => {
    const cardId = `room-card-${room.id}`;
    existingCardIds.add(cardId);
    let existingCard = document.getElementById(cardId);

    const tempContainer = document.createElement("div");
    tempContainer.innerHTML = createRoomCardHtml(
      room,
      currentLang,
      false,
    ).trim();
    const newCardElement = tempContainer.firstElementChild as HTMLElement;

    if (existingCard) {
      if (existingCard.innerHTML !== newCardElement.innerHTML) {
        existingCard.innerHTML = newCardElement.innerHTML;
      }
    } else {
      if (newCardElement) {
        newCardElement.id = cardId;
        grid.appendChild(newCardElement);
      }
    }
  });

  Array.from(grid.children).forEach((child) => {
    if (
      child.id &&
      child.id.startsWith("room-card-") &&
      !existingCardIds.has(child.id)
    ) {
      grid.removeChild(child);
    }
  });
}
