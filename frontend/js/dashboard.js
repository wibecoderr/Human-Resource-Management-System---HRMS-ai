document.addEventListener("DOMContentLoaded", async () => {
  const metrics = qs("#metricGrid");
  setLoading(metrics, "Loading dashboard...");
  try {
    const role = Api.role();
    let dashboard = { employees: 0, openJobs: 0, candidates: 0, pendingLeaves: 0 };
    let employees = [], attendance = [], leaves = [];
    if (role === "admin") dashboard = await Api.get("/api/dashboard/admin");
    else if (role === "hr_recruiter") dashboard = await Api.get("/api/dashboard/hr");
    else if (role === "senior_manager") dashboard = await Api.get("/api/dashboard/manager");
    else dashboard = await Api.get("/api/dashboard/employee");
    try { employees = await Api.get("/api/employees"); } catch {}
    try { attendance = await Api.get("/api/attendance/me"); } catch {}
    try { leaves = role === "employee" ? await Api.get("/api/leave/me") : await Api.get("/api/leave"); } catch {}
    metrics.innerHTML = [
      ["Total Employees", dashboard.employees || dashboard.teamSize || employees.length],
      ["Open Jobs", dashboard.openJobs || 0],
      ["Candidates", dashboard.candidates || 0],
      ["Pending Leaves", dashboard.pendingLeaves || leaves.filter((x) => x.status === "pending").length]
    ].map(([label, value]) => `<div class="card metric"><span>${label}</span><strong>${value}</strong></div>`).join("");
    drawBars("#employeeChart", countBy(employees, "role"));
    drawBars("#opsChart", { Attendance: attendance.length, Leaves: leaves.length, Pending: leaves.filter((x) => x.status === "pending").length });
  } catch (error) {
    setError(metrics, error);
  }
});

function countBy(items, key) {
  return items.reduce((acc, item) => {
    const value = item[key] || "unknown";
    acc[value] = (acc[value] || 0) + 1;
    return acc;
  }, {});
}

function drawBars(selector, data) {
  const chart = qs(selector);
  const entries = Object.entries(data);
  if (!entries.length) {
    chart.innerHTML = '<div class="empty">No chart data available.</div>';
    return;
  }
  const max = Math.max(...entries.map(([, value]) => value), 1);
  chart.innerHTML = entries.map(([label, value]) => `<div class="bar"><div class="bar-fill" style="height:${Math.max(8, (value / max) * 100)}%"></div><span>${escapeHtml(label)} (${value})</span></div>`).join("");
}
