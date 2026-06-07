let activeInterview = null;
let interviewCandidates = [];
const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;

document.addEventListener("DOMContentLoaded", () => {
  qs("#startInterviewBtn")?.addEventListener("click", startInterview);
  qs("#interviewForm")?.addEventListener("submit", submitInterview);
  loadInterviewCandidates();
  updateVoiceSupport();
});

async function loadInterviewCandidates() {
  const select = qs("#interviewCandidateSelect");
  try {
    const result = await Api.get("/api/candidates");
    interviewCandidates = Array.isArray(result) ? result : (result.data || result.candidates || []);
    select.innerHTML = '<option value="">Select candidate</option>' + interviewCandidates
      .map((c) => `<option value="${c.id}">${escapeHtml(c.name)} - ${escapeHtml(c.job_title || c.job_id)}${c.ai_score ? ` (${Math.round(c.ai_score)}%)` : ""}</option>`)
      .join("");
  } catch (error) {
    select.innerHTML = `<option value="">${escapeHtml(error.message)}</option>`;
  }
}

async function startInterview() {
  const candidateId = qs("#interviewCandidateSelect").value;
  if (!candidateId) return toast("Choose a candidate first.", "error");
  setInterviewProgress(45, "Gemini is generating interview questions...");
  try {
    activeInterview = await Api.post("/api/interviews/start", { candidate_id: candidateId });
    renderQuestions(activeInterview.questions || []);
    qs("#interviewStatus").outerHTML = badge(activeInterview.status || "in_progress").replace("<span", '<span id="interviewStatus"');
    setInterviewProgress(100, "Interview ready.");
    toast("Interview questions generated.", "success");
  } catch (error) {
    setInterviewProgress(0, "Could not start interview.");
    toast(error.message, "error");
  }
}

async function submitInterview(event) {
  event.preventDefault();
  if (!activeInterview?.id) return toast("Start an interview first.", "error");
  const answers = qsa("[data-question-id]").map((node) => ({
    question_id: Number(node.dataset.questionId),
    question: node.dataset.question,
    answer: node.value.trim()
  }));
  if (answers.some((item) => !item.answer)) return toast("Answer every question before submitting.", "error");

  setInterviewProgress(65, "Gemini is evaluating answers...");
  try {
    const result = await Api.put(`/api/interviews/${activeInterview.id}/submit`, { answers });
    activeInterview = result.interview;
    renderEvaluation(result.evaluation || activeInterview.ai_report || {});
    setInterviewProgress(100, "Assessment saved.");
    toast("Interview assessment saved.", "success");
  } catch (error) {
    setInterviewProgress(0, "Evaluation failed.");
    toast(error.message, "error");
  }
}

function renderQuestions(questions) {
  const container = qs("#questionList");
  if (!Array.isArray(questions) || !questions.length) {
    container.innerHTML = '<div class="empty">No questions returned.</div>';
    return;
  }
  container.innerHTML = questions.map((item, index) => {
    const id = item.id || index + 1;
    const question = item.question || String(item);
    return `<div class="panel" style="margin-bottom:12px">
      <div class="panel-header"><div><h3>Question ${id}</h3><p>${escapeHtml(question)}</p></div><div class="actions"><button class="btn btn-secondary" type="button" onclick="speakQuestion(${index})">Speak</button><button class="btn btn-secondary" type="button" onclick="dictateAnswer(${index})">Voice Answer</button></div></div>
      <textarea rows="4" data-question-id="${id}" data-question="${escapeHtml(question)}" placeholder="Type or dictate the candidate answer"></textarea>
    </div>`;
  }).join("");
}

function renderEvaluation(evaluation) {
  qs("#technicalScore").textContent = scoreText(evaluation.technical_score);
  qs("#communicationScore").textContent = scoreText(evaluation.communication_score);
  qs("#problemScore").textContent = scoreText(evaluation.problem_solving_score);
  qs("#overallScore").textContent = scoreText(evaluation.overall_score);
  qs("#interviewStatus").outerHTML = badge(evaluation.recommendation || "completed").replace("<span", '<span id="interviewStatus"');
  qs("#interviewRecommendation").textContent = evaluation.recommendation || "No recommendation.";
  qs("#interviewSummary").textContent = evaluation.summary || "No summary returned.";
}

function speakQuestion(index) {
  const node = qsa("[data-question-id]")[index];
  if (!node || !window.speechSynthesis) return toast("Speech synthesis is not available in this browser.", "error");
  window.speechSynthesis.cancel();
  window.speechSynthesis.speak(new SpeechSynthesisUtterance(node.dataset.question));
}

function dictateAnswer(index) {
  const node = qsa("[data-question-id]")[index];
  if (!node || !SpeechRecognition) return toast("Speech recognition is not available in this browser.", "error");
  const recognition = new SpeechRecognition();
  recognition.lang = "en-US";
  recognition.interimResults = false;
  recognition.maxAlternatives = 1;
  recognition.onresult = (event) => {
    const transcript = event.results[0][0].transcript;
    node.value = `${node.value ? `${node.value} ` : ""}${transcript}`.trim();
  };
  recognition.onerror = () => toast("Voice capture failed.", "error");
  recognition.start();
}

function updateVoiceSupport() {
  const text = [];
  text.push(window.speechSynthesis ? "Speech output available" : "Speech output unavailable");
  text.push(SpeechRecognition ? "voice answers available" : "voice answers unavailable");
  qs("#voiceSupport").textContent = text.join("; ") + ".";
}

function scoreText(value) {
  return value === undefined || value === null || value === "" ? "--" : `${Math.round(Number(value))}%`;
}

function setInterviewProgress(percent, label) {
  qs("#interviewProgress").style.width = `${percent}%`;
  qs("#interviewState").textContent = label;
}
