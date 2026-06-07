document.addEventListener("DOMContentLoaded", () => {
  qs("#leaveForm")?.addEventListener("submit", applyLeave);
  loadLeaves();
});

async function loadLeaves() {
  const rows = qs("#leaveRows");
  rows.innerHTML = emptyRow("Loading leave records...", 6);
  const role = Api.role();
  try {
    const leaves = role === "employee" ? await Api.get("/api/leave/me") : await Api.get("/api/leave");
    const visibleLeaves = role === "employee" ? leaves : leaves.filter((l) => l.status === "pending");
    rows.innerHTML = visibleLeaves.length ? visibleLeaves.map((l) => `<tr><td>${escapeHtml(l.leave_type)}</td><td>${fmtDate(l.start_date)}</td><td>${fmtDate(l.end_date)}</td><td>${escapeHtml(l.reason)}</td><td>${badge(l.status)}</td><td data-role="admin,senior_manager" class="actions">${l.status === "pending" ? `<button class="btn btn-success" onclick="updateLeave('${l.id}','approved')">Approve</button><button class="btn btn-danger" onclick="updateLeave('${l.id}','rejected')">Reject</button>` : ""}</td></tr>`).join("") : emptyRow(role === "employee" ? "No leave records." : "No pending leave requests.", 6);
    renderShell(Auth.user() || { role });
  } catch (error) {
    rows.innerHTML = emptyRow(error.message, 6);
  }
}

async function applyLeave(event) {
  event.preventDefault();
  try {
    await Api.post("/api/leave", formData(event.currentTarget));
    event.currentTarget.reset();
    toast("Leave request submitted.", "success");
    await loadLeaves();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function updateLeave(id, status) {
  try {
    const path = status === "rejected" ? `/api/leave/${id}/reject` : `/api/leave/${id}/approve`;
    const payload = status === "rejected" ? {} : { status };
    await Api.put(path, payload);
    toast(`Leave ${status}.`, "success");
    await loadLeaves();
  } catch (error) {
    toast(error.message, "error");
  }
}
