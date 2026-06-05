import { t } from "../i18n.js";

// Модуль управления верхним статус-баром системы
export function updateStatus(textKey, type = "info", isRawText = false) {
  const statusEl = document.getElementById("status");
  if (!statusEl) return;

  statusEl.innerText = isRawText ? textKey : t(textKey);
  statusEl.removeAttribute("data-i18n");

  let colors = "bg-slate-800/80 text-slate-300 border border-slate-700/50";
  if (type === "success")
    colors = "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20";
  if (type === "warning")
    colors = "bg-amber-500/10 text-amber-400 border border-amber-500/20";
  if (type === "error")
    colors = "bg-rose-500/10 text-rose-400 border border-rose-500/20";

  statusEl.className = `text-xs font-medium px-3 py-1.5 rounded-xl backdrop-blur-md shadow-sm transition-all duration-300 ${colors}`;
}
