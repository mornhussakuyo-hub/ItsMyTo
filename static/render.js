function renderCards() {
  state.cardsEl.innerHTML = "";
  state.emptyEl.hidden = state.cards.length > 0;
  for (const card of state.cards) {
    const article = document.createElement("article");
    article.className = `card${card.archived ? " archived" : ""}`;
    article.innerHTML = cardTemplate(card);
    article.addEventListener("click", async (event) => {
      const button = event.target.closest("button[data-action]");
      if (!button) return;
      await handleCardAction(button.dataset.action, card);
    });
    state.cardsEl.append(article);
  }
}

function cardTemplate(card) {
  const key = state.revealCache.get(card.id) || card.apiKeyMask;
  return `
    <div class="name">
      <h2>${escapeHTML(card.name)}</h2>
      <div class="meta">${escapeHTML(card.baseUrl)}</div>
    </div>
    <div>
      <div class="key" data-key>${escapeHTML(key)}</div>
      <div class="meta">${formatDate(card.updatedAt)}</div>
    </div>
    <div>
      <div class="desc">${escapeHTML(card.description || "无详细介绍")}</div>
      ${docButton(card)}
    </div>
    <div class="actions">
      <button class="icon-button" data-action="reveal" title="显示" aria-label="显示"><svg><use href="#i-eye"/></svg></button>
      <button class="icon-button" data-action="copy" title="复制" aria-label="复制"><svg><use href="#i-copy"/></svg></button>
      <button class="icon-button" data-action="edit" title="编辑" aria-label="编辑"><svg><use href="#i-edit"/></svg></button>
      <button class="icon-button" data-action="archive" title="${card.archived ? "取消归档" : "归档"}" aria-label="${card.archived ? "取消归档" : "归档"}"><svg><use href="#i-archive"/></svg></button>
      <button class="icon-button danger" data-action="delete" title="删除" aria-label="删除"><svg><use href="#i-trash"/></svg></button>
    </div>
  `;
}

function docButton(card) {
  if (!card.docUrl) return "";
  return `
    <button class="doc doc-button" data-action="doc" type="button" title="打开文档">
      <svg><use href="#i-link"/></svg>${escapeHTML(card.docUrl)}
    </button>
  `;
}

async function handleCardAction(action, card) {
  try {
    if (action === "reveal") await revealCard(card);
    if (action === "copy") await copyCard(card);
    if (action === "edit") openCardDialog(card);
    if (action === "archive") await archiveCard(card);
    if (action === "delete") await deleteCard(card);
    if (action === "doc") await openDoc(card);
  } catch (error) {
    notify(error.message);
  }
}

async function revealCard(card) {
  const key = await reveal(card.id);
  state.revealCache.set(card.id, key);
  renderCards();
}

async function copyCard(card) {
  const key = await reveal(card.id);
  await navigator.clipboard.writeText(key);
  notify("API Key 已复制");
}

async function archiveCard(card) {
  await request(`/api/cards/${card.id}/archive`, {
    method: "POST",
    body: { archived: !card.archived },
  });
  notify(card.archived ? "已取消归档" : "已归档");
  reloadCurrentView();
}

async function deleteCard(card) {
  if (!confirm(`删除 ${card.name}？`)) return;
  await request(`/api/cards/${card.id}`, { method: "DELETE" });
  state.revealCache.delete(card.id);
  notify("卡片已删除");
  reloadCurrentView();
}

async function openDoc(card) {
  await request("/api/open-url", { method: "POST", body: { url: card.docUrl } });
}
