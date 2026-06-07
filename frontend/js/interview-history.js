let interviewHistory = [];

document.addEventListener("DOMContentLoaded", () => {
  qs("#interviewSearch")?.addEventListener("input", renderInterviewHistory);
  loadInterviewHistory();
});

async function loadInterviewHistory() {
  const tbody = qs("#interviewRows");
  setLoading(tbody, "Loading interviews...");
  try {
    interviewHistory = await Api.get("/api/interviews");
    renderInterviewHistory();
  } catch (error) {
    setError(tbody, error);
  }
}

function renderInterviewHistory() {
  const search = qs("#interviewSearch")?.value.toLowerCase() || "";
  const rows = interviewHistory.filter((item) => `${item.candidate_name} ${item.job_title} ${item.status} ${item.recommendation}`.toLowerCase().includes(search));
  qs("#interviewRows").innerHTML = rows.length ? rows.map((item) => `<tr>
    <td><strong>${escapeHtml(item.candidate_name || item.candidate_id)}</strong><br><small class="muted">ID ${escapeHtml(item.candidate_id)}</small></td>
    <td>${escapeHtml(item.job_title || item.job_id)}</td>
    <td>${badge(item.status)}</td>
    <td>${item.overall_score === undefined || item.overall_score === null ? "--" : `${Math.round(Number(item.overall_score))}%`}</td>
    <td>${badge(item.recommendation || "Pending")}</td>
    <td>${fmtDate(item.created_at)}</td>
    <td><button class="btn btn-secondary" onclick="showInterview('${item.id}')">Details</button></td>
  </tr>`).join("") : emptyRow("No interviews found.", 7);
}

async function showInterview(id) {
  const panel = qs("#interviewDetails");
  setLoading(panel, "Loading interview...");
  try {
    const item = await Api.get(`/api/interviews/${id}`);
    const report = item.ai_report || {};
    panel.innerHTML = `<div class="panel-header"><div><h2>${escapeHtml(item.candidate_name || "Candidate")}</h2><p>${escapeHtml(item.job_title || item.job_id)}</p></div>${badge(item.recommendation || item.status)}</div>
      <div class="grid grid-2">
        <div><h3>Technical</h3><strong>${scoreText(item.technical_score)}</strong></div>
        <div><h3>Communication</h3><strong>${scoreText(item.communication_score)}</strong></div>
        <div><h3>Problem Solving</h3><strong>${scoreText(item.problem_solving_score)}</strong></div>
        <div><h3>Overall</h3><strong>${scoreText(item.overall_score)}</strong></div>
      </div>
      <div style="margin-top:18px"><h3>Summary</h3><p>${escapeHtml(report.summary || "No summary yet.")}</p></div>
      <div class="grid grid-2" style="margin-top:18px">
        <div><h3>Strengths</h3>${renderPlainList(report.strengths, "No strengths listed.")}</div>
        <div><h3>Risks</h3>${renderPlainList(report.risks, "No risks listed.")}</div>
      </div>
      <div style="margin-top:18px"><h3>Answers</h3>${renderAnswers(item.answers || [])}</div>`;
  } catch (error) {
    setError(panel, error);
  }
}

function renderAnswers(answers) {
  return Array.isArray(answers) && answers.length ? answers.map((item) => `<div class="panel" style="margin-bottom:10px"><strong>${escapeHtml(item.question)}</strong><p>${escapeHtml(item.answer)}</p></div>`).join("") : '<p class="muted">No answers saved.</p>';
}

function renderPlainList(items, empty) {
  return Array.isArray(items) && items.length ? `<ul>${items.map((s) => `<li>${escapeHtml(s)}</li>`).join("")}</ul>` : `<p class="muted">${escapeHtml(empty)}</p>`;
}

function scoreText(value) {
  return value === undefined || value === null || value === "" ? "--" : `${Math.round(Number(value))}%`;
}
