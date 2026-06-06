/**
 * Internationalization (i18n) Module
 * Handles multi-language application state and DOM string updates.
 */

const translations = {
  ru: {
    title: "SmartHome OS",
    statusLoading: "Загрузка...",
    statusOnline: "Система онлайн",
    statusNeedAuth: "Нужен вход (токен не найден)",
    statusAuthErr: "Сессия истекла. Войдите заново",
    statusConnErr: "Ошибка связи с сервером",
    statusRequired: "Требуется авторизация",
    statusAuthenticating: "Авторизация...",
    statusSuccessAuth: "Успешный вход!",
    statusLogout: "Вышли из системы",
    loginTitle: "Вход в систему",
    usernameLabel: "Имя пользователя",
    passwordLabel: "Пароль",
    loginBtn: "Войти",
    logoutBtn: "Выйти",
    noData: "Нет данных",
    noRooms: "Нет доступных комнат или датчиков",
    updated: "Обновлено",
    targetTemp: "Целевая температура:", // Added missing key
  },
  et: {
    title: "SmartHome OS",
    statusLoading: "Laadimine...",
    statusOnline: "Süsteem on võrgus",
    statusNeedAuth: "Vajalik sisselogimine (lubatõendit ei leitud)",
    statusAuthErr: "Sessioon on aegunud. Logige uuesti sisse",
    statusConnErr: "Ühenduse viga serveriga",
    statusRequired: "Autoriseerimine on vajalik",
    statusAuthenticating: "Autoriseerimine...",
    statusSuccessAuth: "Edukalt sisse logitud!",
    statusLogout: "Süsteemist välja logitud",
    loginTitle: "Sisselogimine",
    usernameLabel: "Kasutajatunnus",
    passwordLabel: "Parool",
    loginBtn: "Logi sisse",
    logoutBtn: "Logi välja",
    noData: "Andmed puuduvad",
    noRooms: "Saadaolevaid ruume või andureid ei leitud",
    updated: "Uuendatud",
    targetTemp: "Sihttemperatuur:", // Added missing key
  },
  en: {
    title: "SmartHome OS",
    statusLoading: "Loading...",
    statusOnline: "System Online",
    statusNeedAuth: "Login required (token not found)",
    statusAuthErr: "Session expired. Please log in again",
    statusConnErr: "Server connection error",
    statusRequired: "Authorization required",
    statusAuthenticating: "Authenticating...",
    statusSuccessAuth: "Success login!",
    statusLogout: "Logged out",
    loginTitle: "Sign In",
    usernameLabel: "Username",
    passwordLabel: "Password",
    loginBtn: "Login",
    logoutBtn: "Logout",
    noData: "No data",
    noRooms: "No rooms or sensors available",
    updated: "Updated",
    targetTemp: "Target Temperature:", // Added missing key
  },
};

const LANG_KEY = "smart_home_lang";

// Optional runtime listener callback to hook into state shifts (e.g. re-rendering grids)
let onLanguageChangeCallback = null;

export function getCurrentLang() {
  return localStorage.getItem(LANG_KEY) || "en"; // Defaulting to 'en' is cleaner for global standard profiles
}

export function setLang(lang) {
  if (translations[lang]) {
    localStorage.setItem(LANG_KEY, lang);

    translatePage();

    if (typeof onLanguageChangeCallback === "function") {
      onLanguageChangeCallback(lang);
    }
  }
}

/**
 * Register a listener to handle dynamic component rerenders upon translation shifts
 */
export function registerLangChangeListener(callback) {
  onLanguageChangeCallback = callback;
}

/**
 * Translates a given key based on language fallback architecture
 */
export function t(key, lang = getCurrentLang()) {
  return translations[lang]?.[key] || key;
}

/**
 * Parses DOM layout looking for data-i18n target schemas to apply translations
 */
export function translatePage() {
  // Read localized state configuration once per compilation cycle
  const currentLang = getCurrentLang();
  const elements = document.querySelectorAll("[data-i18n]");

  elements.forEach((el) => {
    const key = el.getAttribute("data-i18n");
    const translation = t(key, currentLang);

    // Advanced type check adjustments for native input nodes
    if (el.tagName === "INPUT") {
      if (el.type === "submit" || el.type === "button") {
        el.value = translation;
      } else {
        el.placeholder = translation;
      }
    } else {
      el.innerText = translation;
    }
  });
}
