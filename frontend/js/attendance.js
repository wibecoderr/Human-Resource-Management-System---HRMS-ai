document.addEventListener("DOMContentLoaded", () => {
  qs("#checkInBtn")?.addEventListener("click", () => attendanceAction("/api/attendance/checkin", "Checked in."));
  qs("#checkOutBtn")?.addEventListener("click", () => attendanceAction("/api/attendance/checkout", "Checked out."));
  qs("#attendanceLookup")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const id = formData(event.currentTarget).employee_id;
    await loadAttendance(`/api/attendance/${encodeURIComponent(id)}`, `Employee ${id}`);
  });
  loadAttendance("/api/attendance/me", "My records");
});

async function attendanceAction(path, message) {
  try {
    await Api.post(path);
    toast(message, "success");
    await loadAttendance("/api/attendance/me", "My records");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function loadAttendance(path, scope) {
  const rows = qs("#attendanceRows");
  rows.innerHTML = emptyRow("Loading attendance...", 4);
  qs("#attendanceScope").textContent = scope;
  try {
    const records = await Api.get(path);
    rows.innerHTML = records.length ? records.map((r) => `<tr><td>${fmtDate(r.date)}</td><td>${fmtDateTime(r.check_in)}</td><td>${fmtDateTime(r.check_out)}</td><td>${escapeHtml(r.employee_id)}</td></tr>`).join("") : emptyRow("No attendance records.", 4);
  } catch (error) {
    rows.innerHTML = emptyRow(error.message, 4);
  }
}
