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
  const query = state.searchInput.value.trim();
  const includeArchive = state.filter === "archive";
  if (query) {
    state.cards = await request("/api/search", {
      method: "POST",
      body: { query, includeArchive },
    });
  } else {
    state.cards = await request(`/api/cards?archive=${includeArchive ? "1" : "0"}`);
  }
  renderCards();
}

async function reveal(id) {
  if (state.revealCache.has(id)) return state.revealCache.get(id);
  const result = await request(`/api/cards/${id}/reveal`, { method: "POST", body: {} });
  return result.apiKey;
}
