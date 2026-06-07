let jobs = [];

document.addEventListener("DOMContentLoaded", () => {
  qs("#addJobBtn")?.addEventListener("click", () => editJob());
  qs("#jobSearch")?.addEventListener("input", renderJobs);
  qs("#jobForm")?.addEventListener("submit", saveJob);
  loadJobs();
});

async function loadJobs() {
  const rows = qs("#jobRows");
  rows.innerHTML = emptyRow("Loading jobs...", 6);
  try {
    jobs = await Api.get("/api/jobs?all=true");
    renderJobs();
  } catch (error) {
    rows.innerHTML = emptyRow(error.message, 6);
  }
}

function renderJobs() {
  const q = qs("#jobSearch").value.toLowerCase();
  const list = jobs.filter((j) => [j.title, j.department, j.location, j.status, j.required_skills].join(" ").toLowerCase().includes(q));
  qs("#jobRows").innerHTML = list.length ? list.map((j) => `<tr><td><strong>${escapeHtml(j.title)}</strong><br><small class="muted">${escapeHtml(j.required_skills || j.description || "")}</small></td><td>${escapeHtml(j.department)}</td><td>${escapeHtml(j.location || "-")}</td><td>${badge(j.status)}</td><td>${fmtDate(j.created_at)}</td><td class="actions">${String(j.status).toLowerCase() === "open" ? `<a class="btn btn-primary" href="apply.html?job_id=${encodeURIComponent(j.id)}">Apply</a>` : ""}<button class="btn btn-secondary" onclick="editJob('${j.id}')">Edit</button><button class="btn btn-danger" onclick="deleteJob('${j.id}')">Close</button></td></tr>`).join("") : emptyRow("No jobs found.", 6);
}

function editJob(id) {
  const form = qs("#jobForm");
  form.reset();
  const fields = form.elements;
  const job = jobs.find((j) => j.id === id);
  qs("#jobModalTitle").textContent = job ? "Edit Job" : "Create Job";
  fields.id.value = job?.id || "";
  fields.title.value = job?.title || "";
  fields.department.value = job?.department || "";
  fields.required_skills.value = job?.required_skills || "";
  fields.location.value = job?.location || "";
  fields.description.value = job?.description || "";
  fields.status.value = job?.status || "open";
  openModal("jobModal");
}

async function saveJob(event) {
  event.preventDefault();
  const data = formData(event.currentTarget);
  const id = data.id;
  delete data.id;
  if (!id) delete data.status;
  try {
    id ? await Api.put(`/api/jobs/${id}`, data) : await Api.post("/api/jobs", data);
    toast("Job saved.", "success");
    closeModal("jobModal");
    await loadJobs();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteJob(id) {
  if (!confirm("Close this job?")) return;
  try {
    await Api.delete(`/api/jobs/${id}`);
    toast("Job closed.", "success");
    await loadJobs();
  } catch (error) {
    toast(error.message, "error");
  }
}
