// Модуль для переключения экранов авторизации и дашборда
export function toggleAuthView(showLogin) {
  const loginSection = document.getElementById("login-section");
  const mainDashboard = document.getElementById("main-dashboard");
  const logoutBtn = document.getElementById("logout-btn");

  if (showLogin) {
    if (loginSection) loginSection.classList.remove("hidden");
    if (mainDashboard) mainDashboard.classList.add("hidden");
    if (logoutBtn) logoutBtn.classList.add("hidden");
  } else {
    if (loginSection) loginSection.classList.add("hidden");
    if (mainDashboard) mainDashboard.classList.remove("hidden");
    if (logoutBtn) logoutBtn.classList.remove("hidden");
  }
}
