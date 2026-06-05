// Модуль бизнес-логики для обработки климатических данных
export function getTemperatureStyles(tempStr, isOffline) {
  if (isOffline) {
    return {
      percent: 0,
      colorClass: "text-slate-300 dark:text-slate-700",
      barColor: "from-slate-400 to-slate-500",
    };
  }

  const temp = parseFloat(tempStr) || 0;

  // Рассчитываем процент заполнения полоски (диапазон от 10°C до 30°C)
  let percent = ((temp - 10) / (30 - 10)) * 100;
  percent = Math.max(5, Math.min(100, percent));

  // Цветовые схемы по умолчанию: Зеленый / Комфорт (20°C - 24°C)
  let colorClass = "text-emerald-500 dark:text-emerald-400";
  let barColor = "from-emerald-500 to-teal-400";

  if (temp < 20) {
    colorClass = "text-sky-500 dark:text-sky-400";
    barColor = "from-sky-500 to-blue-400";
  } else if (temp > 24) {
    colorClass = "text-amber-500 dark:text-rose-400";
    barColor = "from-amber-500 to-rose-500";
  }

  return { percent, colorClass, barColor };
}
