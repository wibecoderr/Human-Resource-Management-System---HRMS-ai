document.addEventListener("DOMContentLoaded", () => {
  qs("#payrollForm")?.addEventListener("submit", savePayroll);
  qs("#payrollUpdateForm")?.addEventListener("submit", savePayrollUpdate);
  loadPayroll();
});

async function loadPayroll() {
  const rows = qs("#payrollRows");
  rows.innerHTML = emptyRow("Loading payroll...", 6);
  const role = Api.role();
  try {
    const records = role === "employee" ? await Api.get("/api/payroll/me") : await Api.get("/api/payroll");
    rows.innerHTML = records.length ? records.map((p) => `<tr><td>${escapeHtml(p.employee_name || "-")}</td><td><code>${escapeHtml(p.employee_id)}</code></td><td>${p.month}</td><td>${p.year}</td><td>${money(p.basic_salary)}</td><td data-role="admin"><button class="btn btn-secondary" onclick="editPayroll('${p.id}', ${Number(p.basic_salary)})">Edit Salary</button></td></tr>`).join("") : emptyRow("No payroll records.", 6);
    renderShell(Auth.user() || { role });
  } catch (error) {
    rows.innerHTML = emptyRow(error.message, 6);
  }
}

async function savePayroll(event) {
  event.preventDefault();
  const data = formData(event.currentTarget);
  data.month = Number(data.month);
  data.year = Number(data.year);
  data.basic_salary = Number(data.basic_salary);
  try {
    await Api.post("/api/payroll", data);
    event.currentTarget.reset();
    toast("Payroll saved.", "success");
    await loadPayroll();
  } catch (error) {
    toast(error.message, "error");
  }
}

function editPayroll(id, current) {
  const form = qs("#payrollUpdateForm");
  form.elements.id.value = id;
  form.elements.basic_salary.value = current;
  openModal("payrollModal");
}

async function savePayrollUpdate(event) {
  event.preventDefault();
  const data = formData(event.currentTarget);
  const basic = Number(data.basic_salary);
  try {
    await Api.put(`/api/payroll/${data.id}`, { basic_salary: basic });
    toast("Payroll updated.", "success");
    closeModal("payrollModal");
    await loadPayroll();
  } catch (error) {
    toast(error.message, "error");
  }
}
