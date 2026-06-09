const state = {
  cards: [],
  settings: { theme: "system" },
  filter: "active",
  editingId: null,
  revealCache: new Map(),
  searchSeen: new Set(),
  searchSource: null,
  cardsEl: $("#cards"),
  emptyEl: $("#empty"),
  cardDialog: $("#cardDialog"),
  settingsDialog: $("#settingsDialog"),
  cardForm: $("#cardForm"),
  settingsForm: $("#settingsForm"),
  searchInput: $("#searchInput"),
  searchButton: $("#searchButton"),
};

init();

async function init() {
  bindEvents();
  await loadSettings();
  await refresh();
}

function bindEvents() {
  $("#newButton").addEventListener("click", () => openCardDialog());
  $("#settingsButton").addEventListener("click", openSettingsDialog);
  bindDialogCloseButtons();
  bindTabs();
  bindCardForm();
  bindSettingsForm();
  state.searchButton.addEventListener("click", performSearch);
  state.searchInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter") performSearch();
  });
}

function bindTabs() {
  document.querySelectorAll(".tab").forEach((tab) => {
    tab.addEventListener("click", async () => {
      document.querySelectorAll(".tab").forEach((item) => item.classList.remove("active"));
      tab.classList.add("active");
      state.filter = tab.dataset.filter;
      if (state.searchInput.value.trim()) {
        performSearch();
      } else {
        await refresh();
      }
    });
  });
}
