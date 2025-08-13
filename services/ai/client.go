package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ElvinEga/go-openai"
)

type AIService struct {
	client *openai.Client
	model  string
}

func NewAIService(apiKey, model string) *AIService {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"
	return &AIService{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}
}

func (s *AIService) GenerateProjectPlan(projectName, projectDescription string) (*ProjectPlan, error) {
	prompt := fmt.Sprintf(`Create a comprehensive project plan for:
Project Name: %s
Project Description: %s

Please provide:
1. A list of key features with overviews and details
2. A technology stack with description and individual technologies
3. A detailed build strategy

Use the create_project_plan function to return this information in a structured format.`,
		projectName, projectDescription)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Tools:      FunctionSchemasToTools(FunctionSchemas),
			ToolChoice: "auto",
		},
	)

	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI didn't return any choices")
	}

	message := resp.Choices[0].Message
	if len(message.ToolCalls) == 0 {
		return nil, fmt.Errorf("AI didn't call any tools")
	}

	toolCall := message.ToolCalls[0]
	if toolCall.Function.Name != "create_project_plan" {
		return nil, fmt.Errorf("unexpected tool called: %s", toolCall.Function.Name)
	}

	var plan ProjectPlan
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	return &plan, nil
}

func (s *AIService) GeneratePRD(featureName, projectDescription string) (*PRDContent, error) {
	prompt := fmt.Sprintf(`Generate a comprehensive Product Requirements Document (PRD) for the feature "%s" in a project described as: "%s".

The PRD should include:
- User stories
- Acceptance criteria
- Technical specifications
- Success metrics

Use the create_prd function to return this information in markdown format.`,
		featureName, projectDescription)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Tools:      FunctionSchemasToTools(FunctionSchemas),
			ToolChoice: "auto",
		},
	)

	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI didn't return any choices")
	}

	message := resp.Choices[0].Message
	if len(message.ToolCalls) == 0 {
		return nil, fmt.Errorf("AI didn't call any tools")
	}

	toolCall := message.ToolCalls[0]
	if toolCall.Function.Name != "create_prd" {
		return nil, fmt.Errorf("unexpected tool called: %s", toolCall.Function.Name)
	}

	var prd PRDContent
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &prd); err != nil {
		return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	return &prd, nil
}

// Helper function to convert function schemas to tools
func FunctionSchemasToTools(functionSchemas []openai.FunctionDefinition) []openai.Tool {
	tools := make([]openai.Tool, len(functionSchemas))
	for i, functionSchema := range functionSchemas {
		tools[i] = openai.Tool{
			Type:     "function",
			Function: &functionSchema,
		}
	}
	return tools
}

// Structs to match function schemas
type ProjectPlan struct {
	Features      []FeatureData `json:"features"`
	TechStack     TechStackData `json:"tech_stack"`
	BuildStrategy string        `json:"build_strategy"`
}

type FeatureData struct {
	Name     string `json:"name"`
	Overview string `json:"overview"`
	Details  string `json:"details"`
}

type TechStackData struct {
	Description string          `json:"description"`
	Items       []StackItemData `json:"items"`
}

type StackItemData struct {
	Name     string `json:"name"`
	Overview string `json:"overview"`
	Details  string `json:"details"`
}

type PRDContent struct {
	Content string `json:"content"`
}
