import { t } from "../i18n/i18n.js";
import { RoomData } from "../types/telemetry.js";
import { createRoomCardHtml } from "./cardTemplate.js";
import { generateSystemBadgesHtml } from "./badges.js";
import { getTemperatureStyles } from "./utils.js";

/**
 * Removes card elements from the grid container if they no longer exist in backend telemetry.
 */
export function removeStaleCards(
  grid: HTMLDivElement,
  validRoomIds: Set<number>,
): void {
  const existingCards = grid.querySelectorAll(".room-card");
  existingCards.forEach((cardNode) => {
    const cardEl = cardNode as HTMLDivElement;
    const roomId = parseInt(cardEl.getAttribute("data-room-id") || "0", 10);
    if (!validRoomIds.has(roomId)) {
      cardEl.remove();
    }
  });
}

/**
 * Inserts a newly discovered room card directly into the DOM tree.
 */
export function appendNewCard(
  grid: HTMLDivElement,
  room: RoomData,
  currentLang: string,
): HTMLDivElement | null {
  const tempWrapper = document.createElement("div");
  tempWrapper.innerHTML = createRoomCardHtml(room, currentLang, true);
  const newCard = tempWrapper.firstElementChild as HTMLDivElement | null;

  if (newCard) {
    grid.appendChild(newCard);
  }
  return newCard;
}

/**
 * Mutates specific elements of an existing room card without re-rendering the HTML layout.
 */
export function updateExistingCard(
  card: HTMLDivElement,
  room: RoomData,
  currentLang: string,
): void {
  const isOffline = !room.temperature || room.temperature === "N/A";
  const tempValue = isOffline ? "N/A" : room.temperature;
  const targetTemp = room.target_temperature ?? 23.0;
  const styles = getTemperatureStyles(room.temperature, isOffline);

  const formattedTime =
    room.last_update && room.last_update !== "0001-01-01T00:00:00Z"
      ? new Date(room.last_update).toLocaleTimeString(
          currentLang === "ru" ? "ru-RU" : "et-EE",
        )
      : t("noData");

  // Perform targeted DOM nodes selection
  const tempDisplay = card.querySelector(
    ".temp-value-display",
  ) as HTMLElement | null;
  const timeDisplay = card.querySelector(".time-display") as HTMLElement | null;
  const progressBar = card.querySelector(".progress-bar") as HTMLElement | null;
  const acContainer = card.querySelector(
    ".ac-badge-container",
  ) as HTMLElement | null;

  if (tempDisplay && tempDisplay.textContent !== tempValue) {
    tempDisplay.textContent = tempValue;
    tempDisplay.className = `temp-value-display text-4xl font-extrabold font-mono tracking-tight ${styles.colorClass}`;
  }

  if (timeDisplay && timeDisplay.textContent !== formattedTime) {
    timeDisplay.textContent = formattedTime;
  }

  if (progressBar) {
    progressBar.style.width = `${styles.percent}%`;
  }

  if (acContainer) {
    acContainer.innerHTML = generateSystemBadgesHtml(
      room.temperature,
      targetTemp,
      isOffline,
      currentLang,
    );
  }
}
