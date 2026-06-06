import { t } from "../i18n/i18n.js";
export function getRoomIcon(roomId) {
    // Простая и лаконичная фабрика иконок под типы комнат
    const icons = {
        1: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/></svg>`, // Living Room
        2: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/></svg>`, // Bedroom
        3: `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"/></svg>`, // Kitchen
    };
    return (icons[roomId] ||
        `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"/></svg>`);
}
export function getTemperatureStyles(temperature, isOffline) {
    if (isOffline || !temperature) {
        return {
            colorClass: "text-slate-400",
            barColor: "from-slate-400 to-slate-500",
            percent: 0,
        };
    }
    const val = parseFloat(temperature);
    let percent = ((val - 15) / (30 - 15)) * 100; // Нормализация от 15°C до 30°C
    percent = Math.max(0, Math.min(100, percent));
    if (val < 19)
        return {
            colorClass: "text-blue-400",
            barColor: "from-blue-400 to-indigo-500",
            percent,
        };
    if (val > 24)
        return {
            colorClass: "text-rose-400",
            barColor: "from-amber-400 to-rose-500",
            percent,
        };
    return {
        colorClass: "text-emerald-400",
        barColor: "from-teal-400 to-emerald-500",
        percent,
    };
}
export function toggleAuthView(showLogin) {
    const loginSection = document.getElementById("login-section");
    const mainDashboard = document.getElementById("main-dashboard");
    const logoutBtn = document.getElementById("logout-btn");
    if (showLogin) {
        loginSection?.classList.remove("hidden");
        mainDashboard?.classList.add("hidden");
        logoutBtn?.classList.add("hidden");
    }
    else {
        loginSection?.classList.add("hidden");
        mainDashboard?.classList.remove("hidden");
        logoutBtn?.classList.remove("hidden");
    }
}
export function updateStatus(i18nKey, severity, raw = false) {
    const statusEl = document.getElementById("status");
    if (!statusEl)
        return;
    statusEl.className =
        "text-xs font-bold px-2.5 py-1 rounded-lg border transition-all ";
    if (severity === "success")
        statusEl.className +=
            "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
    else if (severity === "error")
        statusEl.className += "bg-rose-500/10 text-rose-400 border-rose-500/20";
    else if (severity === "warning")
        statusEl.className += "bg-amber-500/10 text-amber-400 border-amber-500/20";
    else
        statusEl.className += "bg-sky-500/10 text-sky-400 border-sky-500/20";
    if (raw) {
        statusEl.textContent = i18nKey;
    }
    else {
        statusEl.setAttribute("data-i18n", i18nKey);
        statusEl.textContent = t(i18nKey);
    }
}
