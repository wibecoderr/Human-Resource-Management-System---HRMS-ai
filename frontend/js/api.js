const savedApiBase = localStorage.getItem("hrms_api_base") || "";
const isLiveServer = ["127.0.0.1:5500", "localhost:5500"].includes(location.host);
const inferredApiBase = isLiveServer ? "http://localhost:8080" : "";
const API_BASE_URL = savedApiBase && !savedApiBase.includes(":5500") ? savedApiBase : inferredApiBase;

const Api = {
  token() {
    return localStorage.getItem("hrms_token");
  },
  userId() {
    return localStorage.getItem("hrms_user_id");
  },
  role() {
    return localStorage.getItem("hrms_role");
  },
  async request(path, options = {}) {
    const headers = new Headers(options.headers || {});
    const body = options.body;
    if (body && !(body instanceof FormData)) headers.set("Content-Type", "application/json");
    const token = Api.token();
    if (token) headers.set("Authorization", `Bearer ${token}`);

    const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });
    const text = await response.text();
    let payload = {};
    if (text) {
      try { payload = JSON.parse(text); } catch { payload = { message: text }; }
    }

    if (response.status === 401) {
      Auth.clear();
      window.location.href = "login.html";
      throw new Error("Session expired. Please sign in again.");
    }

    if (!response.ok || payload.success === false) {
      throw new Error(formatApiError(payload, response.statusText));
    }
    return payload.data ?? payload;
  },
  get(path) {
    return Api.request(path);
  },
  post(path, data) {
    return Api.request(path, { method: "POST", body: JSON.stringify(data || {}) });
  },
  put(path, data) {
    return Api.request(path, { method: "PUT", body: JSON.stringify(data || {}) });
  },
  delete(path) {
    return Api.request(path, { method: "DELETE" });
  }
};

function formatApiError(payload, fallback) {
  if (Array.isArray(payload.errors)) {
    return payload.errors.map((error) => `${error.field}: ${error.message}`).join(", ");
  }
  return payload.message || payload.errors || fallback || "Request failed";
}

function qs(selector, root = document) {
  return root.querySelector(selector);
}

function qsa(selector, root = document) {
  return Array.from(root.querySelectorAll(selector));
}

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#039;"
  }[char]));
}

function fmtDate(value) {
  if (!value) return "-";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? escapeHtml(value) : date.toLocaleDateString();
}

function fmtDateTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? escapeHtml(value) : date.toLocaleString();
}

function money(value) {
  return Number(value || 0).toLocaleString(undefined, { style: "currency", currency: "USD" });
}

function badge(status) {
  const normalized = String(status || "unknown").toLowerCase();
  const cls = normalized.includes("approved") || normalized.includes("active") || normalized.includes("open") || normalized.includes("offered")
    ? "success"
    : normalized.includes("pending") || normalized.includes("shortlisted") || normalized.includes("interviewing") || normalized.includes("interviewed")
      ? "warning"
      : normalized.includes("reject") || normalized.includes("inactive") || normalized.includes("closed")
        ? "danger"
        : "neutral";
  return `<span class="badge ${cls}">${escapeHtml(status || "Unknown")}</span>`;
}

function toast(message, type = "") {
  let stack = qs(".toast-stack");
  if (!stack) {
    stack = document.createElement("div");
    stack.className = "toast-stack";
    document.body.appendChild(stack);
  }
  const item = document.createElement("div");
  item.className = `toast ${type}`;
  item.textContent = message;
  stack.appendChild(item);
  setTimeout(() => item.remove(), 4200);
}

function setLoading(container, label = "Loading records...") {
  container.innerHTML = `<div class="loading">${label}</div>`;
}

function setError(container, error) {
  container.innerHTML = `<div class="error">${escapeHtml(error.message || error)}</div>`;
}

function emptyRow(message, colspan = 6) {
  return `<tr><td colspan="${colspan}" class="empty">${escapeHtml(message)}</td></tr>`;
}

function formData(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function openModal(id) {
  qs(`#${id}`)?.classList.add("show");
}

function closeModal(id) {
  qs(`#${id}`)?.classList.remove("show");
}

function attachModalClose() {
  qsa("[data-close-modal]").forEach((button) => {
    button.addEventListener("click", () => closeModal(button.dataset.closeModal));
  });
  qsa(".modal-backdrop").forEach((modal) => {
    modal.addEventListener("click", (event) => {
      if (event.target === modal) modal.classList.remove("show");
    });
  });
}

function paginate(items, page, perPage = 8) {
  const totalPages = Math.max(1, Math.ceil(items.length / perPage));
  const current = Math.min(Math.max(1, page), totalPages);
  const start = (current - 1) * perPage;
  return { rows: items.slice(start, start + perPage), current, totalPages };
}
