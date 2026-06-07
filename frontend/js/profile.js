document.addEventListener("DOMContentLoaded", async () => {
  qs("#profileForm")?.addEventListener("submit", saveProfile);
  setTimeout(loadProfile, 150);
});

async function loadProfile() {
  try {
    const user = await Api.get("/api/auth/me");
    localStorage.setItem("hrms_user", JSON.stringify(user));
    renderShell(user);
    qs("#profileCard").innerHTML = `<div class="panel-header"><div><h2>${escapeHtml(user.name)}</h2><p>${escapeHtml(roleLabels[user.role] || user.role)}</p></div>${badge(user.status)}</div><div class="grid grid-2"><p><strong>Email</strong><br>${escapeHtml(user.email)}</p><p><strong>Phone</strong><br>${escapeHtml(user.phone_no)}</p><p><strong>User ID</strong><br>${escapeHtml(user.id)}</p><p><strong>Created</strong><br>${fmtDate(user.created_at)}</p></div>`;
    qs("#profileForm").elements.name.value = user.name || "";
    qs("#profileForm").elements.phone_no.value = user.phone_no || "";
  } catch (error) {
    setError(qs("#profileCard"), error);
  }
}

async function saveProfile(event) {
  event.preventDefault();
  try {
    await Api.put(`/api/employees/${Api.userId()}`, formData(event.currentTarget));
    toast("Profile updated.", "success");
    await loadProfile();
  } catch (error) {
    toast(error.message, "error");
  }
}
