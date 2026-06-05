const THEME_KEY = "smart_home_theme";

export function initTheme() {
  const savedTheme = localStorage.getItem(THEME_KEY) || "dark";
  applyTheme(savedTheme);
  return savedTheme;
}

export function toggleTheme() {
  const currentTheme = document.documentElement.classList.contains("dark")
    ? "dark"
    : "light";
  const newTheme = currentTheme === "dark" ? "light" : "dark";
  applyTheme(newTheme);
  localStorage.setItem(THEME_KEY, newTheme);
}

function applyTheme(theme) {
  const themeBtn = document.getElementById("theme-btn");
  if (theme === "dark") {
    document.documentElement.classList.add("dark");
    // Глубокий темный фон
    document.body.className =
      "bg-slate-950 text-slate-100 font-sans min-h-screen flex flex-col transition-colors duration-500";
    if (themeBtn) themeBtn.innerHTML = "🌙";
  } else {
    document.documentElement.classList.remove("dark");
    // Контрастный светлый фон (Slate-100) и темный текст по умолчанию
    document.body.className =
      "bg-slate-100 text-slate-900 font-sans min-h-screen flex flex-col transition-colors duration-500";
    if (themeBtn) themeBtn.innerHTML = "☀️";
  }
}

// Функция для создания эффекта плавного каскадного появления карточек
export function animateCards() {
  const cards = document.querySelectorAll(".room-card");
  cards.forEach((card, index) => {
    // Сбрасываем начальное состояние
    card.style.opacity = "0";
    card.style.transform = "translateY(20px)";
    card.style.transition = "all 0.6s cubic-bezier(0.16, 1, 0.3, 1)";

    // Включаем анимацию с задержкой (каждая следующая карточка появляется чуть позже)
    setTimeout(() => {
      card.style.opacity = "1";
      card.style.transform = "translateY(0)";
    }, index * 100);
  });
}
