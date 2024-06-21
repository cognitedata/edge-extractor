package llm

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

type GcpLlmProvider struct {
	projectID          string
	regionLocation     string
	serviceAccountJson []byte
	modelName          string
	client             *genai.Client
}

// NewGcpLlmProvider creates a new GcpLlmProvider with the provided projectID and regionLocation.
func NewGcpLlmProvider(projectID, regionLocation, modelName string) *GcpLlmProvider {
	return &GcpLlmProvider{projectID: projectID, regionLocation: regionLocation, modelName: modelName}
}

func (gcp *GcpLlmProvider) Init() error {
	ctx := context.Background()
	var err error
	gcp.client, err = genai.NewClient(ctx, gcp.projectID, gcp.regionLocation)

	return err
}

func (gcp *GcpLlmProvider) ImageToText(imageData []byte, prompt string) (string, error) {
	ctx := context.Background()
	model := gcp.client.GenerativeModel(gcp.modelName)
	model.SetTemperature(0)
	imgBlob := genai.Blob{
		MIMEType: "image/jpeg",
		Data:     imageData,
	}
	promptPart := genai.Text(prompt)

	res, err := model.GenerateContent(ctx, imgBlob, promptPart)
	if err != nil {
		return "", fmt.Errorf("unable to generate contents: %v", err)
	}

	if len(res.Candidates) == 0 ||
		len(res.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}
	response := ""
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			responseText := part.(genai.Text)
			response += string(responseText)
		}
	}
	return response, nil
}

func (gcp *GcpLlmProvider) Close() {
	gcp.client.Close()
}
