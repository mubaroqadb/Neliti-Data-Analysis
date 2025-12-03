package vertexai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/research-data-analysis/config"
)

// GeminiRequest untuk request ke Vertex AI
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content untuk konten request
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part untuk bagian konten
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse untuk response dari Vertex AI
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
	Error      *ErrorInfo  `json:"error,omitempty"`
}

// Candidate untuk kandidat response
type Candidate struct {
	Content Content `json:"content"`
}

// ErrorInfo untuk informasi error
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// getAccessToken mendapatkan access token dari metadata server
func getAccessToken() (string, error) {
	// Di Google Cloud Functions, token bisa didapat dari metadata server
	metadataURL := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
	
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback ke environment variable
		return os.Getenv("GOOGLE_ACCESS_TOKEN"), nil
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// GenerateContent memanggil Vertex AI Gemini untuk generate content
func GenerateContent(prompt string) (string, error) {
	projectID := config.GetGCPProjectID()
	region := config.GetVertexAIRegion()
	model := "gemini-2.0-flash-exp"

	url := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		region, projectID, region, model,
	)

	accessToken, err := getAccessToken()
	if err != nil {
		return "", err
	}

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}

// GenerateResearchRecommendations menghasilkan rekomendasi metode penelitian
func GenerateResearchRecommendations(context string) (string, error) {
	prompt := fmt.Sprintf(`Anda adalah ahli metodologi penelitian dan statistik. Berdasarkan konteks penelitian berikut, berikan rekomendasi metode analisis yang sesuai.

Konteks Penelitian:
%s

Berikan rekomendasi dalam format JSON dengan struktur berikut:
{
  "recommendations": [
    {
      "method": "nama metode analisis",
      "category": "descriptive/inferential/correlation/regression",
      "reasoning": "penjelasan mengapa metode ini cocok",
      "priority": 1,
      "assumptions": "asumsi yang perlu dipenuhi"
    }
  ]
}

Berikan minimal 3-5 rekomendasi metode yang relevan, diurutkan berdasarkan prioritas.`, context)

	return GenerateContent(prompt)
}

// GenerateAnalysisInterpretation menghasilkan interpretasi hasil analisis
func GenerateAnalysisInterpretation(method, results string) (string, error) {
	prompt := fmt.Sprintf(`Anda adalah ahli statistik penelitian. Interpretasikan hasil analisis berikut dalam bahasa yang mudah dipahami.

Metode Analisis: %s
Hasil: %s

Berikan interpretasi dalam format JSON:
{
  "interpretation": "penjelasan hasil dalam bahasa sederhana",
  "effect_size": "interpretasi effect size jika ada",
  "practical_implications": "implikasi praktis dari hasil",
  "conclusion": "kesimpulan terkait hipotesis/tujuan penelitian"
}`, method, results)

	return GenerateContent(prompt)
}

// GenerateResearchSummary menghasilkan ringkasan penelitian
func GenerateResearchSummary(analysisContext string) (string, error) {
	prompt := fmt.Sprintf(`Anda adalah penulis akademis berpengalaman. Buat ringkasan komprehensif dari sesi analisis penelitian berikut.

%s

Berikan ringkasan dalam format JSON:
{
  "executive_summary": "ringkasan eksekutif 2-3 paragraf",
  "key_findings": ["temuan utama 1", "temuan utama 2", ...],
  "methodology_notes": "catatan tentang metodologi yang digunakan",
  "limitations": ["keterbatasan 1", "keterbatasan 2", ...],
  "future_recommendations": ["rekomendasi penelitian lanjutan"]
}`, analysisContext)

	return GenerateContent(prompt)
}