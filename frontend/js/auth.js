const roleLabels = {
  admin: "Admin",
  hr_recruiter: "HR Recruiter",
  senior_manager: "Senior Manager",
  employee: "Employee"
};

const navItems = [
  { href: "dashboard.html", label: "Dashboard", icon: "layout-dashboard", roles: ["admin", "hr_recruiter", "senior_manager", "employee"] },
  { href: "employees.html", label: "Employees", icon: "users", roles: ["admin", "senior_manager"] },
  { href: "attendance.html", label: "Attendance", icon: "clock", roles: ["admin", "employee"] },
  { href: "leave.html", label: "Leave", icon: "calendar-days", roles: ["admin", "senior_manager", "employee"] },
  { href: "payroll.html", label: "Payroll", icon: "wallet", roles: ["admin", "employee"] },
  { href: "jobs.html", label: "Jobs", icon: "briefcase", roles: ["admin", "hr_recruiter"] },
  { href: "candidates.html", label: "Candidates", icon: "user-check", roles: ["admin", "hr_recruiter"] },
  { href: "performance.html", label: "Performance", icon: "trending-up", roles: ["admin", "hr_recruiter", "senior_manager"] },
  { href: "resume-screening.html", label: "AI Screening", icon: "scan-search", roles: ["admin", "hr_recruiter"] },
  { href: "interviews.html", label: "AI Interview", icon: "messages-square", roles: ["admin", "hr_recruiter"] },
  { href: "interview-history.html", label: "Interview History", icon: "history", roles: ["admin", "hr_recruiter", "senior_manager"] },
  { href: "profile.html", label: "Profile", icon: "circle-user", roles: ["employee", "admin", "hr_recruiter", "senior_manager"] },
  { href: "#", label: "Settings", icon: "settings", roles: ["admin"], disabled: true }
];

const Auth = {
  clear() {
    localStorage.removeItem("hrms_token");
    localStorage.removeItem("hrms_user_id");
    localStorage.removeItem("hrms_role");
    localStorage.removeItem("hrms_user");
  },
  save(auth) {
    localStorage.setItem("hrms_token", auth.token);
    localStorage.setItem("hrms_user_id", auth.user_id);
    localStorage.setItem("hrms_role", auth.role);
  },
  user() {
    try { return JSON.parse(localStorage.getItem("hrms_user") || "null"); }
    catch { return null; }
  },
  async require() {
    if (!Api.token()) {
      window.location.href = "login.html";
      return null;
    }
    try {
      const user = await Api.get("/api/auth/me");
      localStorage.setItem("hrms_user", JSON.stringify(user));
      localStorage.setItem("hrms_role", user.role || Api.role());
      renderShell(user);
      return user;
    } catch (error) {
      toast(error.message, "error");
      return null;
    }
  },
  async logout() {
    try { await Api.post("/api/auth/logout"); } catch {}
    Auth.clear();
    window.location.href = "login.html";
  }
};

function icon(name) {
  const icons = {
    "layout-dashboard": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="9"/><rect x="14" y="3" width="7" height="5"/><rect x="14" y="12" width="7" height="9"/><rect x="3" y="16" width="7" height="5"/></svg>',
    users: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>',
    clock: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>',
    "calendar-days": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 2v4M16 2v4M3 10h18"/><rect x="3" y="4" width="18" height="18" rx="2"/></svg>',
    wallet: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 7V5a2 2 0 0 0-2-2H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3v4a1 1 0 0 1-1 1H5a2 2 0 0 1-2-2V5"/><path d="M18 12h.01"/></svg>',
    briefcase: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2M2 13h20"/></svg>',
    "user-check": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="m16 11 2 2 4-4"/></svg>',
    "trending-up": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m22 7-8.5 8.5-5-5L2 17"/><path d="M16 7h6v6"/></svg>',
    "scan-search": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 3H5a2 2 0 0 0-2 2v2M17 3h2a2 2 0 0 1 2 2v2M7 21H5a2 2 0 0 1-2-2v-2M13 13l5 5M15 10a5 5 0 1 1-10 0 5 5 0 0 1 10 0Z"/></svg>',
    "messages-square": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 9a2 2 0 0 1-2 2H6l-4 4V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2z"/><path d="M18 9h2a2 2 0 0 1 2 2v10l-4-4h-6a2 2 0 0 1-2-2v-1"/></svg>',
    history: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5M12 7v5l4 2"/></svg>',
    "circle-user": '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="10" r="3"/><path d="M7 20.7a7 7 0 0 1 10 0"/></svg>',
    settings: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.88-.34 1.7 1.7 0 0 0-1 1.55V21a2 2 0 1 1-4 0v-.09a1.7 1.7 0 0 0-1-1.55 1.7 1.7 0 0 0-1.88.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.55-1H3a2 2 0 1 1 0-4h.09a1.7 1.7 0 0 0 1.55-1 1.7 1.7 0 0 0-.34-1.88l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.55V3a2 2 0 1 1 4 0v.09a1.7 1.7 0 0 0 1 1.55 1.7 1.7 0 0 0 1.88-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.7 1.7 0 0 0 19.4 9c.25.6.84 1 1.55 1H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.55 1Z"/></svg>',
    menu: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M3 12h18M3 18h18"/></svg>',
    bell: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 21h4"/><path d="M18 8a6 6 0 1 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"/></svg>',
    logout: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5M21 12H9"/></svg>',
    plus: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5v14"/></svg>',
    edit: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>',
    trash: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/></svg>'
  };
  return icons[name] || "";
}

function renderShell(user) {
  const role = user.role || Api.role();
  const current = location.pathname.split("/").pop() || "dashboard.html";
  const nav = qs("#navList");
  if (nav) {
    nav.innerHTML = navItems
      .filter((item) => item.roles.includes(role))
      .map((item) => `<a class="nav-link ${item.href === current ? "active" : ""} ${item.disabled ? "muted" : ""}" href="${item.disabled ? "#" : item.href}">${icon(item.icon)}<span>${item.label}</span></a>`)
      .join("");
  }
  qs("#userName") && (qs("#userName").textContent = user.name || "HRMS User");
  qs("#userRole") && (qs("#userRole").textContent = roleLabels[role] || role);
  qs("#userAvatar") && (qs("#userAvatar").textContent = (user.name || "U").slice(0, 1).toUpperCase());
  qsa("[data-role]").forEach((node) => {
    const roles = node.dataset.role.split(",");
    node.classList.toggle("hide", !roles.includes(role));
  });
}

function setupShell() {
  qs("#menuBtn") && (qs("#menuBtn").innerHTML = icon("menu"));
  qs("#notificationBtn") && (qs("#notificationBtn").innerHTML = icon("bell"));
  qs("#logoutBtn") && (qs("#logoutBtn").innerHTML = icon("logout"));
  qs("#menuBtn")?.addEventListener("click", () => document.body.classList.toggle("sidebar-open"));
  qs("#logoutBtn")?.addEventListener("click", Auth.logout);
  qs("#notificationBtn")?.addEventListener("click", async () => {
    qs("#notificationMenu")?.classList.toggle("show");
    await loadNotifications();
  });
  qs("#markReadBtn")?.addEventListener("click", markNotificationsRead);
  attachModalClose();
}

async function loadNotifications() {
  const menu = qs("#notificationMenu");
  if (!menu || !Api.token()) return;
  try {
    const data = await Api.get("/api/notifications");
    const notifications = data.notifications || [];
    const items = notifications.length
      ? notifications.map((item) => `<div class="notification-item ${item.is_read ? "" : "unread"}" data-id="${item.id}"><strong>${escapeHtml(item.title)}</strong><br>${escapeHtml(item.message)}</div>`).join("")
      : '<div class="notification-item">No notifications.</div>';
    menu.innerHTML = `${items}<button class="btn btn-ghost" id="markReadBtn" style="width:100%">Mark as read (${data.unread_count || 0})</button>`;
    qs("#markReadBtn")?.addEventListener("click", markNotificationsRead);
  } catch (error) {
    menu.innerHTML = `<div class="notification-item">${escapeHtml(error.message)}</div>`;
  }
}

async function markNotificationsRead() {
  const unread = qsa(".notification-item.unread[data-id]");
  await Promise.all(unread.map((item) => Api.put(`/api/notifications/${item.dataset.id}/read`)));
  unread.forEach((item) => item.classList.remove("unread"));
  await loadNotifications();
}

document.addEventListener("DOMContentLoaded", () => {
  setupShell();
  if (!document.body.classList.contains("auth-body")) Auth.require();
});
