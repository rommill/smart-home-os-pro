import {
  loginRequest,
  fetchTelemetry,
  updateTargetTemperatureRequest,
} from "./api.js";
import { getToken, saveToken, removeToken, isAuthorized } from "./auth.js";
import { updateStatus, renderRooms, toggleAuthView } from "./render.js";
import {
  getCurrentLang,
  setLang,
  translatePage,
  registerLangChangeListener,
} from "./i18n.js";
import { initTheme, toggleTheme } from "./theme.js";

let telemetryInterval = null;

/**
 * Global climate orchestration hook bound to window context for dynamic slider interactions.
 */
window.updateTargetTemperature = async function (roomId, value) {
  const token = getToken();
  if (!token) {
    updateStatus("statusNeedAuth", "warning");
    return;
  }

  try {
    // Go backend and MongoDB
    await updateTargetTemperatureRequest(roomId, value, token);
    console.log(
      `[Climate Control] Room ${roomId} target temperature updated to ${value}°C`,
    );
  } catch (err) {
    console.error(
      `[Climate Control] Synchronization failure on Room ${roomId}:`,
      err,
    );
    if (err.message === "Unauthorized") {
      updateStatus("statusAuthErr", "error");
      removeToken();
      toggleAuthView(true);
      stopPolling();
    }
  }
};

/**
 * Checks authorization and fetches fresh telemetry data from the Go backend
 */
async function checkAndRefreshTelemetry() {
  const token = getToken();
  if (!token) {
    updateStatus("statusNeedAuth", "warning");
    toggleAuthView(true);
    stopPolling();
    return;
  }

  try {
    const data = await fetchTelemetry(token);
    renderRooms(data); // Render will now smart-update values instead of destroying inputs
    updateStatus("statusOnline", "success");
    toggleAuthView(false);
  } catch (err) {
    console.error("Telemetry fetch error:", err);
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

function startPolling() {
  if (!telemetryInterval) {
    telemetryInterval = setInterval(checkAndRefreshTelemetry, 3000);
  }
}

function stopPolling() {
  if (telemetryInterval) {
    clearInterval(telemetryInterval);
    telemetryInterval = null;
  }
}

registerLangChangeListener(() => {
  translatePage();
  checkAndRefreshTelemetry();
});

/**
 * Main application orchestration lifecycle
 */
document.addEventListener("DOMContentLoaded", () => {
  const loginForm = document.getElementById("login-form");
  const logoutBtn = document.getElementById("logout-btn");
  const langSelect = document.getElementById("lang-select");
  const themeBtn = document.getElementById("theme-btn");

  initTheme();
  if (themeBtn) {
    themeBtn.addEventListener("click", () => toggleTheme());
  }

  const savedLang = getCurrentLang();
  if (langSelect) {
    langSelect.value = savedLang;
    langSelect.addEventListener("change", (e) => setLang(e.target.value));
  }

  translatePage();

  if (loginForm) {
    loginForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const usernameInput = document.getElementById("username").value;
      const passwordInput = document.getElementById("password").value;
      updateStatus("statusAuthenticating", "info");

      try {
        const res = await loginRequest(usernameInput, passwordInput);
        if (res && res.token) {
          saveToken(res.token);
          updateStatus("statusSuccessAuth", "success");
          toggleAuthView(false);
          checkAndRefreshTelemetry();
          startPolling();
        }
      } catch (err) {
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
