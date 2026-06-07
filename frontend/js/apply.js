document.addEventListener("DOMContentLoaded", () => {
  qs("#applyForm")?.addEventListener("submit", submitApplication);
  loadOpenJobs();
});

async function loadOpenJobs() {
  const select = qs("#applyJobSelect");
  try {
    const jobs = await Api.get("/api/jobs");
    const selectedJobID = new URLSearchParams(location.search).get("job_id") || "";
    select.innerHTML = '<option value="">Select job</option>' + jobs
      .map((job) => `<option value="${job.id}">${escapeHtml(job.title)} - ${escapeHtml(job.location || job.department)}</option>`)
      .join("");
    if (selectedJobID) select.value = selectedJobID;
  } catch (error) {
    select.innerHTML = `<option value="">${escapeHtml(error.message)}</option>`;
  }
}

async function submitApplication(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const file = form.resume.files[0];
  if (!file) return showApplyMessage("Choose a PDF resume.", "error");
  if (file.type && file.type !== "application/pdf") return showApplyMessage("Only PDF resumes are allowed.", "error");

  const data = new FormData(form);
  showApplyMessage("Submitting application...");
  try {
    const candidate = await Api.request("/api/candidates/apply", { method: "POST", body: data });
    form.reset();
    await loadOpenJobs();
    showApplyMessage(`Application submitted successfully. Candidate ID: ${candidate.id}.`, "success");
  } catch (error) {
    showApplyMessage(error.message, "error");
  }
}

function showApplyMessage(message, type = "") {
  const node = qs("#applyMessage");
  node.textContent = message;
  node.style.color = type === "error" ? "#dc2626" : type === "success" ? "#16a34a" : "";
}
