function openCardDialog(card = null) {
  state.editingId = card?.id || null;
  $("#cardDialogTitle").textContent = card ? "编辑卡片" : "新增卡片";
  state.cardForm.reset();
  state.cardForm.name.value = card?.name || "";
  state.cardForm.baseUrl.value = card?.baseUrl || "";
  state.cardForm.apiKey.value = "";
  state.cardForm.apiKey.placeholder = card ? "留空则不修改 APIKEY" : "";
  state.cardForm.apiKey.required = !card;
  state.cardForm.description.value = card?.description || "";
  state.cardForm.docUrl.value = card?.docUrl || "";
  state.cardDialog.showModal();
}

function openSettingsDialog() {
  state.settingsForm.reset();
  state.settingsForm.embeddingUrl.value = state.settings.embeddingUrl || "";
  state.settingsForm.embeddingModel.value = state.settings.embeddingModel || "";
  state.settingsForm.embeddingBatchSize.value = state.settings.embeddingBatchSize || 10;
  state.settingsForm.embeddingMaxTokens.value = state.settings.embeddingMaxTokens || 8192;
  state.settingsForm.embeddingApiKey.value = "";
  state.settingsForm.embeddingApiKey.placeholder = state.settings.hasEmbeddingApiKey
    ? "已配置，留空则保留"
    : "";
  state.settingsForm.theme.value = state.settings.theme || "system";
  state.settingsForm.autostart.checked = !!state.settings.autostart;
  state.settingsForm.clearEmbeddingKey.checked = false;
  state.settingsDialog.showModal();
}

function bindDialogCloseButtons() {
  document.querySelectorAll("[data-close]").forEach((button) => {
    button.addEventListener("click", () => {
      const dialog = document.getElementById(button.dataset.close);
      dialog.close();
    });
  });
}

function bindCardForm() {
  state.cardForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(state.cardForm);
    const payload = Object.fromEntries(form.entries());
    try {
      if (state.editingId) {
        await request(`/api/cards/${state.editingId}`, { method: "PUT", body: payload });
        state.revealCache.delete(state.editingId);
        notify("卡片已更新");
      } else {
        await request("/api/cards", { method: "POST", body: payload });
        notify("卡片已创建");
      }
      state.cardDialog.close();
      reloadCurrentView();
    } catch (error) {
      notify(error.message);
    }
  });
}

function bindSettingsForm() {
  state.settingsForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(state.settingsForm);
    const payload = Object.fromEntries(form.entries());
    payload.autostart = form.get("autostart") === "on";
    payload.clearEmbeddingKey = form.get("clearEmbeddingKey") === "on";
    payload.embeddingBatchSize = Number(payload.embeddingBatchSize || 10);
    payload.embeddingMaxTokens = Number(payload.embeddingMaxTokens || 8192);
    try {
      state.settings = await request("/api/settings", { method: "PUT", body: payload });
      applyTheme();
      state.settingsDialog.close();
      notify("设置已保存");
    } catch (error) {
      notify(error.message);
    }
  });
}

function applyTheme() {
  document.documentElement.dataset.theme = state.settings.theme || "system";
}
