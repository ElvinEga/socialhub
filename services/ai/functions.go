package ai

import "github.com/ElvinEga/go-openai"

var FunctionSchemas = []openai.FunctionDefinition{
	{
		Name:        "create_project_plan",
		Description: "Creates a structured project plan with features, tech stack, and build strategy",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"features": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Feature name",
							},
							"overview": map[string]interface{}{
								"type":        "string",
								"description": "Brief feature overview",
							},
							"details": map[string]interface{}{
								"type":        "string",
								"description": "Detailed feature description in markdown",
							},
						},
						"required": []string{"name", "overview", "details"},
					},
				},
				"tech_stack": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Overall tech stack description",
						},
						"items": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type":        "string",
										"description": "Technology name",
									},
									"overview": map[string]interface{}{
										"type":        "string",
										"description": "Brief technology overview",
									},
									"details": map[string]interface{}{
										"type":        "string",
										"description": "Detailed technology description in markdown",
									},
								},
								"required": []string{"name", "overview", "details"},
							},
						},
					},
					"required": []string{"description", "items"},
				},
				"build_strategy": map[string]interface{}{
					"type":        "string",
					"description": "Detailed build strategy description",
				},
			},
			"required": []string{"features", "tech_stack", "build_strategy"},
		},
	},
	{
		Name:        "create_prd",
		Description: "Creates a Product Requirements Document for a feature",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "PRD content in markdown format",
				},
			},
			"required": []string{"content"},
		},
	},
}
