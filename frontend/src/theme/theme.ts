const THEME_KEY = "smart_home_theme";

export function initTheme(): string {
  const savedTheme: string = localStorage.getItem(THEME_KEY) || "dark";
  applyTheme(savedTheme);
  return savedTheme;
}

export function toggleTheme(): void {
  const currentTheme: string = document.documentElement.classList.contains(
    "dark",
  )
    ? "dark"
    : "light";
  const newTheme: string = currentTheme === "dark" ? "light" : "dark";
  applyTheme(newTheme);
  localStorage.setItem(THEME_KEY, newTheme);
}

function applyTheme(theme: string): void {
  const themeBtn = document.getElementById(
    "theme-btn",
  ) as HTMLButtonElement | null;
  if (theme === "dark") {
    document.documentElement.classList.add("dark");
    document.body.className =
      "bg-slate-950 text-slate-100 font-sans min-h-screen flex flex-col transition-colors duration-500";
    if (themeBtn) themeBtn.innerHTML = "🌙";
  } else {
    document.documentElement.classList.remove("dark");
    document.body.className =
      "bg-slate-100 text-slate-900 font-sans min-h-screen flex flex-col transition-colors duration-500";
    if (themeBtn) themeBtn.innerHTML = "☀️";
  }
}

/**
 * Creates a beautiful cascading entry effect for room cards inside the active view bounds.
 */
export function animateCards(): void {
  const cards = document.querySelectorAll(".room-card");
  cards.forEach((card, index) => {
    const htmlCard = card as HTMLElement;
    htmlCard.style.opacity = "0";
    htmlCard.style.transform = "translateY(20px)";
    htmlCard.style.transition = "all 0.6s cubic-bezier(0.16, 1, 0.3, 1)";

    setTimeout(() => {
      htmlCard.style.opacity = "1";
      htmlCard.style.transform = "translateY(0)";
    }, index * 100);
  });
}
