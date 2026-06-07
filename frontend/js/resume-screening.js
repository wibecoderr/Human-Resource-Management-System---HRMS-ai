let screeningCandidates = [];

document.addEventListener("DOMContentLoaded", () => {
  qs("#uploadResumeBtn")?.addEventListener("click", uploadResume);
  qs("#screeningForm")?.addEventListener("submit", screenCandidate);
  qs("#candidateSelect")?.addEventListener("change", loadAnalysis);
  loadScreeningCandidates();
});

async function loadScreeningCandidates() {
  const select = qs("#candidateSelect");
  try {
    const result = await Api.get("/api/candidates");
    screeningCandidates = Array.isArray(result) ? result : (result.data || result.candidates || []);
    select.innerHTML = '<option value="">Select candidate</option>' + screeningCandidates
      .map((c) => `<option value="${c.id}">${escapeHtml(c.name)} - ${escapeHtml(c.job_title || c.job_id)}${c.ai_score ? ` (${Math.round(c.ai_score)}%)` : ""}</option>`)
      .join("");
    const selectedCandidateID = new URLSearchParams(location.search).get("candidate_id") || "";
    if (selectedCandidateID) {
      select.value = selectedCandidateID;
      await loadAnalysis();
    }
  } catch (error) {
    select.innerHTML = `<option value="">${escapeHtml(error.message)}</option>`;
  }
}

async function uploadResume() {
  const form = qs("#screeningForm");
  const candidateId = form.candidate_id.value;
  const file = form.resume.files[0];
  if (!candidateId) return toast("Choose a candidate first.", "error");
  if (!file) return toast("Choose a PDF resume to upload.", "error");
  if (file.type && file.type !== "application/pdf") return toast("Only PDF resumes are supported.", "error");

  setProgress(35, "Uploading resume...");
  const data = new FormData();
  data.append("resume", file);
  try {
    await Api.request(`/api/candidates/${candidateId}/resume`, { method: "POST", body: data });
    setProgress(100, "Resume uploaded and text extracted.");
    toast("Resume uploaded.", "success");
    await loadScreeningCandidates();
  } catch (error) {
    setProgress(0, "Upload failed.");
    toast(error.message, "error");
  }
}

async function screenCandidate(event) {
  event.preventDefault();
  const candidateId = event.currentTarget.candidate_id.value;
  if (!candidateId) return toast("Choose a candidate first.", "error");

  setProgress(50, "Gemini is analyzing the resume...");
  try {
    const result = await Api.post(`/api/candidates/${candidateId}/screen`);
    renderAnalysis(result.analysis);
    setProgress(100, "Screening complete.");
    toast("Screening complete.", "success");
    await loadScreeningCandidates();
  } catch (error) {
    setProgress(0, "Screening failed.");
    toast(error.message, "error");
  }
}

async function loadAnalysis() {
  const candidateId = qs("#candidateSelect").value;
  if (!candidateId) return clearAnalysis();
  try {
    const analysis = await Api.get(`/api/candidates/${candidateId}/analysis`);
    renderAnalysis(analysis.report || {});
  } catch {
    clearAnalysis();
  }
}

function renderAnalysis(report) {
  const score = Math.round(Number(report.score || 0));
  qs("#aiScore").textContent = `${score}%`;
  qs("#scoreBar").style.width = `${score}%`;
  qs("#analysisStatus").outerHTML = badge(report.recommendation || "Not screened").replace("<span", '<span id="analysisStatus"');
  qs("#matchingSkills").innerHTML = renderList(report.matching_skills, "success", "No matching skills listed.");
  qs("#missingSkills").innerHTML = renderList(report.missing_skills, "warning", "No missing skills listed.");
  qs("#strengths").innerHTML = renderPlainList(report.strengths, "No strengths listed.");
  qs("#weaknesses").innerHTML = renderPlainList(report.weaknesses, "No weaknesses listed.");
  qs("#recommendation").textContent = report.recommendation || "No recommendation available.";
}

function clearAnalysis() {
  qs("#aiScore").textContent = "--";
  qs("#scoreBar").style.width = "0";
  qs("#analysisStatus").outerHTML = '<span id="analysisStatus" class="badge neutral">Not screened</span>';
  qs("#matchingSkills").textContent = "No analysis yet.";
  qs("#missingSkills").textContent = "No analysis yet.";
  qs("#strengths").textContent = "No analysis yet.";
  qs("#weaknesses").textContent = "No analysis yet.";
  qs("#recommendation").textContent = "Upload and screen a resume to generate a recommendation.";
}

function renderList(items, type, empty) {
  return Array.isArray(items) && items.length ? items.map((s) => `<span class="badge ${type}">${escapeHtml(s)}</span>`).join(" ") : empty;
}

function renderPlainList(items, empty) {
  return Array.isArray(items) && items.length ? `<ul>${items.map((s) => `<li>${escapeHtml(s)}</li>`).join("")}</ul>` : empty;
}

function setProgress(percent, label) {
  qs("#screenProgress").style.width = `${percent}%`;
  qs("#screeningState").textContent = label;
}
