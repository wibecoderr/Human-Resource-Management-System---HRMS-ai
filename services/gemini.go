package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hrms/model"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ResumeAnalysis struct {
	Score          float64  `json:"score"`
	MatchingSkills []string `json:"matching_skills"`
	MissingSkills  []string `json:"missing_skills"`
	Strengths      []string `json:"strengths"`
	Weaknesses     []string `json:"weaknesses"`
	Recommendation string   `json:"recommendation"`
}

func AnalyzeResume(jobTitle string, requiredSkills string, jobDescription string, resumeText string) (*ResumeAnalysis, json.RawMessage, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return nil, nil, errors.New("GEMINI_API_KEY is not configured")
	}
	if strings.TrimSpace(resumeText) == "" {
		return nil, nil, errors.New("resume text is empty")
	}

	prompt := fmt.Sprintf(`Analyze this candidate resume against the job.

Return only valid JSON with exactly these fields:
{
  "score": 0,
  "matching_skills": [],
  "missing_skills": [],
  "strengths": [],
  "weaknesses": [],
  "recommendation": "shortlist|hold|reject"
}

Rules:
- score must be a number from 0 to 100.
- recommendation must be shortlist, hold, or reject.
- Compare against required skills and job description.
- Do not include markdown, commentary, or extra fields.

Job title:
%s

Required skills:
%s

Job description:
%s

Resume:
%s`, jobTitle, requiredSkills, jobDescription, resumeText)

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      0.2,
			"responseMimeType": "application/json",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("gemini request failed: %s", geminiErrorMessage(resp.StatusCode, respBody))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, nil, err
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, nil, errors.New("gemini returned no analysis")
	}

	rawJSON := cleanGeminiJSON(geminiResp.Candidates[0].Content.Parts[0].Text)
	var analysis ResumeAnalysis
	if err := json.Unmarshal([]byte(rawJSON), &analysis); err != nil {
		return nil, nil, fmt.Errorf("failed to parse gemini JSON: %w", err)
	}
	normalizeAnalysis(&analysis)
	stored, err := json.Marshal(analysis)
	if err != nil {
		return nil, nil, err
	}
	return &analysis, stored, nil
}

func GenerateInterviewQuestions(jobTitle, jobDescription string, resumeReport json.RawMessage) ([]model.InterviewQuestion, json.RawMessage, error) {
	prompt := fmt.Sprintf(`Generate a structured AI interview for this candidate.

Return only valid JSON with exactly this shape:
{
  "questions": [
    { "id": 1, "question": "..." }
  ]
}

Rules:
- Create 6 questions.
- Include technical, communication, and problem-solving coverage.
- Questions must be specific to the job description and candidate resume analysis.
- Do not include markdown, commentary, or extra fields.

Job title:
%s

Job description:
%s

Candidate resume analysis:
%s`, jobTitle, jobDescription, string(resumeReport))

	raw, err := callGeminiJSON(prompt, 0.35)
	if err != nil {
		questions := fallbackInterviewQuestions(jobTitle, jobDescription)
		stored, marshalErr := json.Marshal(questions)
		if marshalErr != nil {
			return nil, nil, err
		}
		return questions, stored, nil
	}

	var parsed struct {
		Questions []model.InterviewQuestion `json:"questions"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, nil, fmt.Errorf("failed to parse interview questions: %w", err)
	}
	if len(parsed.Questions) == 0 {
		return nil, nil, errors.New("gemini returned no interview questions")
	}
	for i := range parsed.Questions {
		if parsed.Questions[i].ID == 0 {
			parsed.Questions[i].ID = i + 1
		}
		parsed.Questions[i].Question = strings.TrimSpace(parsed.Questions[i].Question)
		if parsed.Questions[i].Question == "" {
			return nil, nil, errors.New("gemini returned an empty interview question")
		}
	}
	stored, err := json.Marshal(parsed.Questions)
	if err != nil {
		return nil, nil, err
	}
	return parsed.Questions, stored, nil
}

func EvaluateInterviewAnswers(jobTitle, jobDescription string, resumeReport json.RawMessage, answers []model.InterviewAnswer) (*model.InterviewEvaluation, json.RawMessage, error) {
	if len(answers) == 0 {
		return nil, nil, errors.New("interview answers are required")
	}
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return nil, nil, err
	}

	prompt := fmt.Sprintf(`Evaluate this AI interview.

Return only valid JSON with exactly these fields:
{
  "technical_score": 0,
  "communication_score": 0,
  "problem_solving_score": 0,
  "overall_score": 0,
  "recommendation": "Hire|Maybe|Reject",
  "strengths": [],
  "risks": [],
  "summary": ""
}

Rules:
- All scores must be numbers from 0 to 100.
- Recommendation must be Hire, Maybe, or Reject.
- Evaluate only the answers provided.
- Be fair, concise, and job-specific.
- Do not include markdown, commentary, or extra fields.

Job title:
%s

Job description:
%s

Candidate resume analysis:
%s

Interview answers:
%s`, jobTitle, jobDescription, string(resumeReport), string(answersJSON))

	raw, err := callGeminiJSON(prompt, 0.2)
	if err != nil {
		evaluation := fallbackInterviewEvaluation(answers)
		stored, marshalErr := json.Marshal(evaluation)
		if marshalErr != nil {
			return nil, nil, err
		}
		return evaluation, stored, nil
	}

	var evaluation model.InterviewEvaluation
	if err := json.Unmarshal(raw, &evaluation); err != nil {
		return nil, nil, fmt.Errorf("failed to parse interview evaluation: %w", err)
	}
	normalizeInterviewEvaluation(&evaluation)
	stored, err := json.Marshal(evaluation)
	if err != nil {
		return nil, nil, err
	}
	return &evaluation, stored, nil
}

func fallbackInterviewQuestions(jobTitle, jobDescription string) []model.InterviewQuestion {
	title := strings.TrimSpace(jobTitle)
	if title == "" {
		title = "this role"
	}
	description := strings.TrimSpace(jobDescription)
	if description == "" {
		description = "the job requirements"
	}
	return []model.InterviewQuestion{
		{ID: 1, Question: fmt.Sprintf("Walk me through your most relevant experience for %s.", title)},
		{ID: 2, Question: fmt.Sprintf("Which skills from this role are your strongest, and where have you used them? Role context: %s", description)},
		{ID: 3, Question: "Describe a technical problem you solved end to end. What tradeoffs did you consider?"},
		{ID: 4, Question: "How do you explain a complex technical issue to a non-technical stakeholder?"},
		{ID: 5, Question: "Tell me about a time you disagreed with a teammate. How did you handle it?"},
		{ID: 6, Question: fmt.Sprintf("If hired for %s, what would you focus on in your first 30 days?", title)},
	}
}

func fallbackInterviewEvaluation(answers []model.InterviewAnswer) *model.InterviewEvaluation {
	answered := 0
	totalWords := 0
	for _, answer := range answers {
		words := len(strings.Fields(answer.Answer))
		if words > 0 {
			answered++
			totalWords += words
		}
	}
	completeness := 0.0
	if len(answers) > 0 {
		completeness = float64(answered) / float64(len(answers))
	}
	communication := clampScore(45 + float64(totalWords)/4)
	technical := clampScore(50 + completeness*35)
	problemSolving := clampScore(48 + completeness*37)
	overall := clampScore((technical + communication + problemSolving) / 3)
	recommendation := "Reject"
	if overall >= 80 {
		recommendation = "Hire"
	} else if overall >= 60 {
		recommendation = "Maybe"
	}
	return &model.InterviewEvaluation{
		TechnicalScore:      technical,
		CommunicationScore:  communication,
		ProblemSolvingScore: problemSolving,
		OverallScore:        overall,
		Recommendation:      recommendation,
		Strengths:           []string{"Completed interview responses", "Provided role-relevant context"},
		Risks:               []string{"Gemini was unavailable, so this is a fallback assessment"},
		Summary:             "Generated fallback assessment because Gemini was unavailable. Review manually before making a hiring decision.",
	}
}

func callGeminiJSON(prompt string, temperature float64) (json.RawMessage, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not configured")
	}

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      temperature,
			"responseMimeType": "application/json",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini request failed: %s", geminiErrorMessage(resp.StatusCode, respBody))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, err
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("gemini returned no content")
	}
	return json.RawMessage(cleanGeminiJSON(geminiResp.Candidates[0].Content.Parts[0].Text)), nil
}

func geminiErrorMessage(statusCode int, body []byte) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		if parsed.Error.Status != "" {
			return fmt.Sprintf("%s (%s)", parsed.Error.Status, parsed.Error.Message)
		}
		return parsed.Error.Message
	}
	return fmt.Sprintf("status %d", statusCode)
}

func cleanGeminiJSON(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end >= start {
		return text[start : end+1]
	}
	return text
}

func normalizeAnalysis(a *ResumeAnalysis) {
	if a.Score < 0 {
		a.Score = 0
	}
	if a.Score > 100 {
		a.Score = 100
	}
	switch a.Recommendation {
	case "shortlist", "hold", "reject":
	default:
		if a.Score >= 80 {
			a.Recommendation = "shortlist"
		} else if a.Score >= 60 {
			a.Recommendation = "hold"
		} else {
			a.Recommendation = "reject"
		}
	}
}

func normalizeInterviewEvaluation(e *model.InterviewEvaluation) {
	e.TechnicalScore = clampScore(e.TechnicalScore)
	e.CommunicationScore = clampScore(e.CommunicationScore)
	e.ProblemSolvingScore = clampScore(e.ProblemSolvingScore)
	e.OverallScore = clampScore(e.OverallScore)
	switch e.Recommendation {
	case "Hire", "Maybe", "Reject":
	default:
		if e.OverallScore >= 80 {
			e.Recommendation = "Hire"
		} else if e.OverallScore >= 60 {
			e.Recommendation = "Maybe"
		} else {
			e.Recommendation = "Reject"
		}
	}
}

func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
