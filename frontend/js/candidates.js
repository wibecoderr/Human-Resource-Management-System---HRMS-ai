let candidates = [];

document.addEventListener("DOMContentLoaded", () => {
  qs("#candidateSearch")?.addEventListener("input", renderCandidates);
  qs("#jobFilter")?.addEventListener("change", loadCandidates);
  qs("#jobFilter")?.addEventListener("keydown", (event) => { if (event.key === "Enter") loadCandidates(); });
  loadCandidates();
});

async function loadCandidates() {
  const rows = qs("#candidateRows");
  rows.innerHTML = emptyRow("Loading candidates...", 7);
  const jobId = qs("#jobFilter").value.trim();
  try {
    const result = await Api.get(jobId ? `/api/candidates?job_id=${encodeURIComponent(jobId)}` : "/api/candidates");
    candidates = Array.isArray(result) ? result : (result.data || result.candidates || []);
    renderCandidates();
  } catch (error) {
    rows.innerHTML = emptyRow(error.message, 7);
  }
}

function renderCandidates() {
  const q = qs("#candidateSearch").value.toLowerCase();
  const list = candidates.filter((c) => [c.name, c.email, c.phone, c.job_title, c.job_id, c.status, scoreLabel(c)].join(" ").toLowerCase().includes(q));
  qs("#candidateRows").innerHTML = list.length ? list.map((c) => `<tr><td><strong>${escapeHtml(c.name)}</strong><br><small class="muted">ID ${escapeHtml(c.id)}</small></td><td>${escapeHtml(c.email)}</td><td>${escapeHtml(c.phone)}</td><td>${escapeHtml(c.job_title || c.job_id)}<br><small class="muted">Job ID ${escapeHtml(c.job_id)}</small></td><td>${scoreBadge(c)}</td><td>${badge(c.status)}</td><td class="actions"><button class="btn btn-secondary" onclick="showCandidate('${c.id}')">Details</button><a class="btn btn-primary" href="resume-screening.html?candidate_id=${encodeURIComponent(c.id)}">Screen</a><select onchange="setCandidateStatus('${c.id}', this.value)"><option value="">Status</option><option value="applied">Applied</option><option value="shortlisted">Shortlisted</option><option value="interviewing">Interviewing</option><option value="interviewed">Interviewed</option><option value="offered">Offered</option><option value="rejected">Rejected</option></select><button class="btn btn-danger" onclick="deleteCandidate('${c.id}')">Delete</button></td></tr>`).join("") : emptyRow("No candidates found.", 7);
}

async function showCandidate(id) {
  const panel = qs("#candidateDetails");
  setLoading(panel, "Loading candidate...");
  try {
    const c = await Api.get(`/api/candidates/${id}`);
    panel.innerHTML = `<div class="panel-header"><div><h2>${escapeHtml(c.name)}</h2><p>Candidate ID ${escapeHtml(c.id)} | Job ID ${escapeHtml(c.job_id)}</p></div>${badge(c.status)}</div><div class="grid grid-2"><p><strong>Email</strong><br>${escapeHtml(c.email)}</p><p><strong>Phone</strong><br>${escapeHtml(c.phone)}</p><p><strong>Job</strong><br>${escapeHtml(c.job_title || c.job_id)}</p><p><strong>Applied</strong><br>${fmtDate(c.created_at)}</p><p><strong>Resume</strong><br>${c.resume_url ? `<a href="${escapeHtml(c.resume_url)}" target="_blank">Open resume</a>` : "-"}</p><p><strong>AI Score</strong><br>${scoreBadge(c)}</p></div><p><strong>Cover Letter</strong><br>${escapeHtml(c.cover_letter || "-")}</p><div class="actions"><a class="btn btn-primary" href="resume-screening.html?candidate_id=${encodeURIComponent(c.id)}">Open AI Screening</a><button class="btn btn-danger" onclick="deleteCandidate('${c.id}')">Delete Candidate</button></div>`;
  } catch (error) {
    setError(panel, error);
  }
}

function scoreLabel(c) {
  return c.ai_score == null ? "Pending" : `${Math.round(Number(c.ai_score))}%`;
}

function scoreBadge(c) {
  if (c.ai_score == null) return '<span class="badge neutral">Pending</span>';
  const score = Math.round(Number(c.ai_score));
  const cls = score >= 80 ? "success" : score >= 60 ? "warning" : "danger";
  return `<span class="badge ${cls}">${score}%</span>`;
}

async function deleteCandidate(id) {
  if (!confirm("Delete this candidate?")) return;
  try {
    await Api.delete(`/api/candidates/${id}`);
    toast("Candidate deleted.", "success");
    qs("#candidateDetails").innerHTML = '<div class="empty">Select a candidate to view details.</div>';
    await loadCandidates();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function setCandidateStatus(id, status) {
  if (!status) return;
  try {
    await Api.put(`/api/candidates/${id}`, { status });
    toast("Candidate status updated.", "success");
    await loadCandidates();
    await showCandidate(id);
  } catch (error) {
    toast(error.message, "error");
  }
}
