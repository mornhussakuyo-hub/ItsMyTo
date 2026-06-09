const state = {
  cards: [],
  settings: { theme: "system" },
  filter: "active",
  editingId: null,
  revealCache: new Map(),
  cardsEl: $("#cards"),
  emptyEl: $("#empty"),
  cardDialog: $("#cardDialog"),
  settingsDialog: $("#settingsDialog"),
  cardForm: $("#cardForm"),
  settingsForm: $("#settingsForm"),
  searchInput: $("#searchInput"),
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
  state.searchInput.addEventListener("input", debounce(refresh, 180));
}

function bindTabs() {
  document.querySelectorAll(".tab").forEach((tab) => {
    tab.addEventListener("click", async () => {
      document.querySelectorAll(".tab").forEach((item) => item.classList.remove("active"));
      tab.classList.add("active");
      state.filter = tab.dataset.filter;
      await refresh();
    });
  });
}
