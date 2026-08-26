import "./index.css";
import { loginRequest } from "./api/api";
import { saveToken, removeToken, isAuthorized } from "./auth/auth";
import { updateStatus, toggleAuthView } from "./render/render";
import {
  getCurrentLang,
  setLang,
  translatePage,
  registerLangChangeListener,
} from "./i18n/i18n";
import { initTheme, toggleTheme } from "./theme/theme";
import {
  startPolling,
  stopPolling,
  checkAndRefreshTelemetry,
  syncTargetTemperature,
} from "./orchestrator/telemetryOrchestrator";

// Subscribe internationalization change listener
registerLangChangeListener(() => {
  translatePage();
  checkAndRefreshTelemetry();
});

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

  // Delegation pattern for climate range sliders
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

  // App entry state validation
  if (isAuthorized()) {
    toggleAuthView(false);
    checkAndRefreshTelemetry();
    startPolling();
  } else {
    toggleAuthView(true);
    updateStatus("statusRequired", "warning");
  }
});
