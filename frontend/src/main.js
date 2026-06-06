import "./index.css";
import { loginRequest, fetchTelemetry, updateTargetTemperatureRequest, } from "./api/api.js";
import { getToken, saveToken, removeToken, isAuthorized } from "./auth/auth.js";
import { updateStatus, renderRooms, toggleAuthView } from "./render/render.js";
import { getCurrentLang, setLang, translatePage, registerLangChangeListener, } from "./i18n/i18n.js";
import { initTheme, toggleTheme } from "./theme/theme.js";
// Enforce strict typing for background polling routines
let telemetryInterval = null;
/**
 * Handles target temperature synchronization with the Go/PostgreSQL backend layer.
 */
async function syncTargetTemperature(roomId, value) {
    const token = getToken();
    if (!token) {
        updateStatus("statusNeedAuth", "warning");
        return;
    }
    try {
        const numericValue = parseFloat(value);
        await updateTargetTemperatureRequest(roomId, numericValue, token);
        console.log(`[Climate Orchestrator] Room ${roomId} mutated successfully to ${numericValue}°C`);
    }
    catch (err) {
        console.error(`[Climate Orchestrator] Microservice execution error on room ${roomId}:`, err);
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
        renderRooms(data);
        updateStatus("statusOnline", "success");
        toggleAuthView(false);
    }
    catch (err) {
        console.error("[Runtime Telemetry] Extraction pipeline error:", err);
        if (err.message === "Unauthorized") {
            updateStatus("statusAuthErr", "error");
            removeToken();
            toggleAuthView(true);
            stopPolling();
        }
        else {
            updateStatus("statusConnErr", "error");
        }
    }
}
function startPolling() {
    if (!telemetryInterval) {
        telemetryInterval = window.setInterval(checkAndRefreshTelemetry, 3000);
    }
}
function stopPolling() {
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
    const loginForm = document.getElementById("login-form");
    const logoutBtn = document.getElementById("logout-btn");
    const langSelect = document.getElementById("lang-select");
    const themeBtn = document.getElementById("theme-btn");
    const roomsGrid = document.getElementById("rooms-grid");
    initTheme();
    if (themeBtn) {
        themeBtn.addEventListener("click", () => toggleTheme());
    }
    const savedLang = getCurrentLang();
    if (langSelect) {
        langSelect.value = savedLang;
        langSelect.addEventListener("change", (e) => {
            const target = e.target;
            setLang(target.value);
        });
    }
    translatePage();
    // Delegation pattern for interactive inputs to secure cleaner state rendering
    if (roomsGrid) {
        roomsGrid.addEventListener("change", (e) => {
            const target = e.target;
            if (target &&
                target.type === "range" &&
                target.hasAttribute("data-room-id")) {
                e.stopPropagation();
                const roomId = parseInt(target.getAttribute("data-room-id") || "0", 10);
                syncTargetTemperature(roomId, target.value);
            }
        });
    }
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
                    await checkAndRefreshTelemetry();
                    startPolling();
                }
            }
            catch (err) {
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
    }
    else {
        toggleAuthView(true);
        updateStatus("statusRequired", "warning");
    }
});
