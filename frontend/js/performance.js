let reviews = [];
let reviewEmployees = [];

document.addEventListener("DOMContentLoaded", () => {
  qs("#performanceForm")?.addEventListener("submit", saveReview);
  qs("#resetReviewBtn")?.addEventListener("click", resetReviewForm);
  loadReviewEmployees();
  loadReviews();
});

async function loadReviewEmployees() {
  const select = qs("#reviewEmployeeSelect");
  if (!select) return;
  try {
    reviewEmployees = await Api.get("/api/employees");
    select.innerHTML = '<option value="">Select employee by ID or name</option>' + reviewEmployees
      .map((e) => `<option value="${escapeHtml(e.id)}">${escapeHtml(e.id)} - ${escapeHtml(e.name)} (${escapeHtml(roleLabels[e.role] || e.role)})</option>`)
      .join("");
  } catch (error) {
    select.innerHTML = `<option value="">${escapeHtml(error.message)}</option>`;
  }
}

async function loadReviews() {
  const rows = qs("#performanceRows");
  rows.innerHTML = emptyRow("Loading reviews...", 7);
  try {
    reviews = await Api.get(Api.role() === "employee" ? "/api/performance/me" : "/api/performance");
    rows.innerHTML = reviews.length ? reviews.map((r) => `
      <tr>
        <td><code>${escapeHtml(r.employee_id)}</code></td>
        <td>${escapeHtml(r.employee_name || "-")}</td>
        <td>${escapeHtml(r.reviewer_name || r.reviewer_id)}</td>
        <td>${escapeHtml(r.review_period)}</td>
        <td>${"*".repeat(Number(r.rating || 0))}</td>
        <td>${badge(r.status)}</td>
        <td class="actions"><button class="btn btn-secondary" onclick="editReview('${r.id}')">Edit</button><button class="btn btn-danger" data-role="admin" onclick="deleteReview('${r.id}')">Delete</button></td>
      </tr>
    `).join("") : emptyRow("No performance reviews.", 7);
    renderShell(Auth.user() || { role: Api.role() });
  } catch (error) {
    rows.innerHTML = emptyRow(error.message, 7);
  }
}

function editReview(id) {
  const review = reviews.find((r) => r.id === id);
  if (!review) return;
  const form = qs("#performanceForm");
  const fields = form.elements;
  fields.id.value = review.id;
  fields.employee_id.value = review.employee_id;
  fields.employee_id.disabled = true;
  fields.review_period.value = review.review_period || "";
  fields.rating.value = review.rating || 3;
  fields.status.value = review.status || "draft";
  fields.feedback.value = review.feedback || "";
  fields.goals.value = review.goals || "";
  qs("#reviewFormTitle").textContent = "Update Review";
  scrollTo({ top: 0, behavior: "smooth" });
}

function resetReviewForm() {
  const form = qs("#performanceForm");
  const fields = form.elements;
  form.reset();
  fields.id.value = "";
  fields.employee_id.disabled = false;
  qs("#reviewFormTitle").textContent = "Create Review";
}

async function saveReview(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = formData(form);
  const id = data.id;
  delete data.id;
  data.rating = Number(data.rating);
  if (id) delete data.employee_id;
  try {
    id ? await Api.put(`/api/performance/${id}`, data) : await Api.post("/api/performance", data);
    toast("Review saved.", "success");
    resetReviewForm();
    await loadReviews();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteReview(id) {
  if (!confirm("Delete this performance review?")) return;
  try {
    await Api.delete(`/api/performance/${id}`);
    toast("Review deleted.", "success");
    await loadReviews();
  } catch (error) {
    toast(error.message, "error");
  }
}
