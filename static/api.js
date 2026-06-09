async function request(url, options = {}) {
  const init = { method: options.method || "GET", headers: {} };
  if (options.body !== undefined) {
    init.headers["Content-Type"] = "application/json";
    init.body = JSON.stringify(options.body);
  }
  const response = await fetch(url, init);
  const payload = await response.json();
  if (!response.ok) throw new Error(payload.error || "请求失败");
  return payload;
}

async function loadSettings() {
  state.settings = await request("/api/settings");
  applyTheme();
}

async function refresh() {
  closeSearchStream();
  const includeArchive = state.filter === "archive";
  state.cards = await request(`/api/cards?archive=${includeArchive ? "1" : "0"}`);
  renderCards();
}

function reloadCurrentView() {
  if (state.searchInput.value.trim()) {
    performSearch();
    return;
  }
  refresh();
}

function performSearch() {
  const query = state.searchInput.value.trim();
  if (!query) {
    refresh();
    return;
  }

  closeSearchStream();
  state.cards = [];
  state.searchSeen = new Set();
  renderCards();

  const params = new URLSearchParams({
    q: query,
    archive: state.filter === "archive" ? "1" : "0",
  });
  const source = new EventSource(`/api/search-stream?${params.toString()}`);
  state.searchSource = source;

  source.addEventListener("card", (event) => {
    const card = JSON.parse(event.data);
    if (state.searchSeen.has(card.id)) return;
    state.searchSeen.add(card.id);
    state.cards.push(card);
    renderCards();
  });
  source.addEventListener("search-error", (event) => {
    const payload = JSON.parse(event.data);
    notify(payload.error || "搜索失败");
    closeSearchStream();
  });
  source.addEventListener("done", () => {
    closeSearchStream();
  });
  source.onerror = () => {
    if (state.searchSource === source) {
      notify("搜索连接已中断");
      closeSearchStream();
    }
  };
}

function closeSearchStream() {
  if (!state.searchSource) return;
  state.searchSource.close();
  state.searchSource = null;
}

async function reveal(id) {
  if (state.revealCache.has(id)) return state.revealCache.get(id);
  const result = await request(`/api/cards/${id}/reveal`, { method: "POST", body: {} });
  return result.apiKey;
}
