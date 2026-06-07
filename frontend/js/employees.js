let employees = [];
let employeePage = 1;

document.addEventListener("DOMContentLoaded", () => {
  qs("#addEmployeeBtn")?.addEventListener("click", () => editEmployee());
  qs("#employeeSearch")?.addEventListener("input", () => { employeePage = 1; renderEmployees(); });
  qs("#prevPage")?.addEventListener("click", () => { employeePage--; renderEmployees(); });
  qs("#nextPage")?.addEventListener("click", () => { employeePage++; renderEmployees(); });
  qs("#employeeForm")?.addEventListener("submit", saveEmployee);
  loadEmployees();
});

async function loadEmployees() {
  const tbody = qs("#employeeRows");
  tbody.innerHTML = emptyRow("Loading employees...", 6);
  try {
    employees = await Api.get("/api/employees");
    renderEmployees();
  } catch (error) {
    tbody.innerHTML = emptyRow(error.message, 6);
  }
}

function renderEmployees() {
  const query = qs("#employeeSearch").value.toLowerCase();
  const filtered = employees.filter((e) => [e.id, e.name, e.email, e.role, e.status].join(" ").toLowerCase().includes(query));
  const page = paginate(filtered, employeePage);
  employeePage = page.current;
  qs("#employeeCount").textContent = `${filtered.length} employees`;
  qs("#pageInfo").textContent = `${page.current} / ${page.totalPages}`;
  qs("#prevPage").disabled = page.current <= 1;
  qs("#nextPage").disabled = page.current >= page.totalPages;
  qs("#employeeRows").innerHTML = page.rows.length ? page.rows.map((e) => `
    <tr><td><code>${escapeHtml(e.id)}</code></td><td><strong>${escapeHtml(e.name)}</strong></td><td>${escapeHtml(e.email)}</td><td>${escapeHtml(roleLabels[e.role] || e.role)}</td><td>${badge(e.status)}</td><td class="actions"><button class="btn btn-secondary" onclick="editEmployee('${e.id}')">Edit</button><button class="btn btn-danger" data-role="admin" onclick="deleteEmployee('${e.id}')">Delete</button></td></tr>
  `).join("") : emptyRow("No employees found.", 6);
  renderShell(Auth.user() || { role: Api.role() });
}

function editEmployee(id) {
  const form = qs("#employeeForm");
  form.reset();
  const fields = form.elements;
  const employee = employees.find((e) => e.id === id);
  qs("#employeeModalTitle").textContent = employee ? "Edit Employee" : "Add Employee";
  fields.id.value = employee?.id || "";
  fields.name.value = employee?.name || "";
  fields.email.value = employee?.email || "";
  fields.email.disabled = Boolean(employee);
  fields.password.disabled = Boolean(employee);
  fields.password.required = !employee;
  fields.phone_no.value = employee?.phone_no || "";
  fields.role.value = employee?.role || "employee";
  fields.status.value = employee?.status || "active";
  openModal("employeeModal");
}

async function saveEmployee(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = formData(form);
  const id = data.id;
  delete data.id;
  if (id) {
    delete data.email;
    delete data.password;
  }
  try {
    id ? await Api.put(`/api/employees/${id}`, data) : await Api.post("/api/employees", data);
    toast("Employee saved.", "success");
    closeModal("employeeModal");
    await loadEmployees();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteEmployee(id) {
  if (!confirm("Deactivate this employee?")) return;
  try {
    await Api.delete(`/api/employees/${id}`);
    toast("Employee deactivated.", "success");
    await loadEmployees();
  } catch (error) {
    toast(error.message, "error");
  }
}
