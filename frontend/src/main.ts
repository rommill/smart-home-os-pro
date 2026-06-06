import "./index.css";
import {
  loginRequest,
  fetchTelemetry,
  updateTargetTemperatureRequest,
} from "./api/api.js";
import { getToken, saveToken, removeToken, isAuthorized } from "./auth/auth.js";
import { updateStatus, renderRooms, toggleAuthView } from "./render/render.js";
import {
  getCurrentLang,
  setLang,
  translatePage,
  registerLangChangeListener,
} from "./i18n/i18n.js";
import { initTheme, toggleTheme } from "./theme/theme.js";
import { RoomData } from "./types/telemetry.js";

// Enforce strict typing for background polling routines
let telemetryInterval: number | null = null;

/**
 * Handles target temperature synchronization with the Go/PostgreSQL backend layer.
 */
async function syncTargetTemperature(
  roomId: number,
  value: string,
): Promise<void> {
  const token: string | null = getToken();
  if (!token) {
    updateStatus("statusNeedAuth", "warning");
    return;
  }

  try {
    const numericValue: number = parseFloat(value);
    await updateTargetTemperatureRequest(roomId, numericValue, token);
    console.log(
      `[Climate Orchestrator] Room ${roomId} mutated successfully to ${numericValue}°C`,
    );
  } catch (err: any) {
    console.error(
      `[Climate Orchestrator] Microservice execution error on room ${roomId}:`,
      err,
    );
    if (err.message === "Unauthorized") {
      updateStatus("statusAuthErr", "error");
      removeToken();
      toggleAuthView(true);
      stopPolling();
    }
  }
}

/**
 * Validates session boundaries and updates interface component states.
 */
async function checkAndRefreshTelemetry(): Promise<void> {
  const token: string | null = getToken();
  if (!token) {
    updateStatus("statusNeedAuth", "warning");
    toggleAuthView(true);
    stopPolling();
    return;
  }

  try {
    const data: RoomData[] = await fetchTelemetry(token);
    renderRooms(data);
    updateStatus("statusOnline", "success");
    toggleAuthView(false);
  } catch (err: any) {
    console.error("[Runtime Telemetry] Extraction pipeline error:", err);
    if (err.message === "Unauthorized") {
      updateStatus("statusAuthErr", "error");
      removeToken();
      toggleAuthView(true);
      stopPolling();
    } else {
      updateStatus("statusConnErr", "error");
    }
  }
}

function startPolling(): void {
  if (!telemetryInterval) {
    telemetryInterval = window.setInterval(checkAndRefreshTelemetry, 3000);
  }
}

function stopPolling(): void {
  if (telemetryInterval) {
    clearInterval(telemetryInterval);
    telemetryInterval = null;
  }
}

// Subscribe runtime event loop handlers to internationalization state bounds
registerLangChangeListener(() => {
  translatePage();
  checkAndRefreshTelemetry();
});

/**
 * Single Page Application execution context initialization.
 */
document.addEventListener("DOMContentLoaded", () => {
  const loginForm = document.getElementById(
    "login-form",
  ) as HTMLFormElement | null;
  const logoutBtn = document.getElementById(
    "logout-btn",
  ) as HTMLButtonElement | null;
  const langSelect = document.getElementById(
    "lang-select",
  ) as HTMLSelectElement | null;
  const themeBtn = document.getElementById(
    "theme-btn",
  ) as HTMLButtonElement | null;
  const roomsGrid = document.getElementById(
    "rooms-grid",
  ) as HTMLDivElement | null;

  initTheme();
  if (themeBtn) {
    themeBtn.addEventListener("click", () => toggleTheme());
  }

  const savedLang: string = getCurrentLang();
  if (langSelect) {
    langSelect.value = savedLang;
    langSelect.addEventListener("change", (e: Event) => {
      const target = e.target as HTMLSelectElement;
      setLang(target.value);
    });
  }

  translatePage();

  // Delegation pattern for interactive inputs to secure cleaner state rendering
  if (roomsGrid) {
    roomsGrid.addEventListener("change", (e: Event) => {
      const target = e.target as HTMLInputElement;
      if (
        target &&
        target.type === "range" &&
        target.hasAttribute("data-room-id")
      ) {
        e.stopPropagation();
        const roomId = parseInt(target.getAttribute("data-room-id") || "0", 10);
        syncTargetTemperature(roomId, target.value);
      }
    });
  }

  if (loginForm) {
    loginForm.addEventListener("submit", async (e: SubmitEvent) => {
      e.preventDefault();
      const usernameInput = (
        document.getElementById("username") as HTMLInputElement
      ).value;
      const passwordInput = (
        document.getElementById("password") as HTMLInputElement
      ).value;
      updateStatus("statusAuthenticating", "info");

      try {
        const res = await loginRequest(usernameInput, passwordInput);
        if (res && res.token) {
          saveToken(res.token);
          updateStatus("statusSuccessAuth", "success");
          toggleAuthView(false);
          await checkAndRefreshTelemetry();
          startPolling();
        }
      } catch (err: any) {
        updateStatus(err.message, "error", true);
      }
    });
  }

  if (logoutBtn) {
    logoutBtn.addEventListener("click", () => {
      removeToken();
      updateStatus("statusLogout", "info");
      toggleAuthView(true);
      stopPolling();
    });
  }

  if (isAuthorized()) {
    toggleAuthView(false);
    checkAndRefreshTelemetry();
    startPolling();
  } else {
    toggleAuthView(true);
    updateStatus("statusRequired", "warning");
  }
});
