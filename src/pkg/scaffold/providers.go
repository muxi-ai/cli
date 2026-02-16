package scaffold

// LLMProvider represents an LLM provider configuration
type LLMProvider struct {
	Name         string   // Display name
	Vendor       string   // Vendor key for model prefix
	DefaultModel string   // Default model (without vendor prefix)
	KeyPrefix    string   // Expected API key prefix (for validation hint)
	SecretName   string   // Secret name (e.g., OPENAI_API_KEY)
	AltEnvVars   []string // Alternative env var names to check (e.g., GEMINI_API_KEY for Google)
}

// LLMProviders is the list of all supported providers
var LLMProviders = []LLMProvider{
	{Name: "OpenAI", Vendor: "openai", DefaultModel: "gpt-5-mini", KeyPrefix: "sk-", SecretName: "OPENAI_API_KEY"},
	{Name: "Anthropic", Vendor: "anthropic", DefaultModel: "claude-sonnet-4-5", KeyPrefix: "sk-ant-", SecretName: "ANTHROPIC_API_KEY"},
	{Name: "Google", Vendor: "google", DefaultModel: "gemini-2.0-flash", KeyPrefix: "", SecretName: "GOOGLE_API_KEY", AltEnvVars: []string{"GEMINI_API_KEY"}},
	{Name: "Mistral", Vendor: "mistral", DefaultModel: "mistral-large-latest", KeyPrefix: "", SecretName: "MISTRAL_API_KEY"},
	{Name: "Groq", Vendor: "groq", DefaultModel: "llama-3.3-70b-versatile", KeyPrefix: "gsk_", SecretName: "GROQ_API_KEY"},
	{Name: "xAI", Vendor: "xai", DefaultModel: "grok-4", KeyPrefix: "xai-", SecretName: "XAI_API_KEY"},
	{Name: "DeepSeek", Vendor: "deepseek", DefaultModel: "deepseek-chat", KeyPrefix: "sk-", SecretName: "DEEPSEEK_API_KEY"},
	{Name: "Cohere", Vendor: "cohere", DefaultModel: "command-r-plus-08-2024", KeyPrefix: "", SecretName: "COHERE_API_KEY"},
	{Name: "Together", Vendor: "together", DefaultModel: "meta-llama/Llama-3.3-70b-Instruct", KeyPrefix: "", SecretName: "TOGETHER_API_KEY"},
	{Name: "Fireworks", Vendor: "fireworks", DefaultModel: "accounts/fireworks/models/llama-v3p3-70b-instruct", KeyPrefix: "", SecretName: "FIREWORKS_API_KEY"},
	{Name: "Perplexity", Vendor: "perplexity", DefaultModel: "sonar-pro", KeyPrefix: "pplx-", SecretName: "PERPLEXITY_API_KEY"},
	{Name: "OpenRouter", Vendor: "openrouter", DefaultModel: "openai/gpt-4o", KeyPrefix: "sk-or-", SecretName: "OPENROUTER_API_KEY"},
	{Name: "Moonshot", Vendor: "moonshot", DefaultModel: "kimi-k2-instruct", KeyPrefix: "", SecretName: "MOONSHOT_API_KEY"},
	{Name: "Minimax", Vendor: "minimax", DefaultModel: "abab6.5s-chat", KeyPrefix: "", SecretName: "MINIMAX_API_KEY"},
	{Name: "GLM", Vendor: "glm", DefaultModel: "glm-4-plus", KeyPrefix: "", SecretName: "GLM_API_KEY"},
	{Name: "Vercel AI", Vendor: "vercel", DefaultModel: "openai/gpt-5-mini", KeyPrefix: "", SecretName: "VERCEL_API_KEY"},
	{Name: "Anyscale", Vendor: "anyscale", DefaultModel: "meta-llama/Meta-Llama-3.1-70B-Instruct", KeyPrefix: "", SecretName: "ANYSCALE_API_KEY"},
}

// LocalProvider represents a local LLM provider (Ollama/llama_cpp)
type LocalProvider struct {
	Name       string
	Vendor     string
	DefaultURL string
}

// LocalProviders is the list of local providers
var LocalProviders = []LocalProvider{
	{Name: "Ollama", Vendor: "ollama", DefaultURL: "http://localhost:11434/v1"},
	{Name: "llama_cpp", Vendor: "llama_cpp", DefaultURL: "http://localhost:8080/v1"},
}

// EnterpriseProvider represents an enterprise cloud provider
type EnterpriseProvider struct {
	Name         string
	Vendor       string
	YAMLTemplate string
	NextSteps    []string
}

// EnterpriseProviders is the list of enterprise providers
var EnterpriseProviders = []EnterpriseProvider{
	{
		Name:   "Azure OpenAI",
		Vendor: "azure",
		YAMLTemplate: `# Azure OpenAI - uncomment and configure:
# llm:
#   api_keys:
#     azure: "${{ secrets.AZURE_API_KEY }}"
#   models:
#     - text: "azure/<deployment-name>"
#       api_base: "https://<resource-name>.openai.azure.com"
#       api_version: "2024-02-15-preview"`,
		NextSteps: []string{
			"Edit formation.yaml and uncomment the Azure configuration",
			"Fill in your resource name, deployment name, and API version",
			"Run 'muxi secrets set AZURE_API_KEY' to add your API key",
		},
	},
	{
		Name:   "AWS Bedrock",
		Vendor: "bedrock",
		YAMLTemplate: `# AWS Bedrock - uncomment and configure:
# llm:
#   models:
#     - text: "bedrock/anthropic.claude-3-sonnet"
#       aws_region: "us-east-1"`,
		NextSteps: []string{
			"Edit formation.yaml and uncomment the Bedrock configuration",
			"Set your AWS region and model ID",
			"Ensure AWS credentials are configured (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)",
		},
	},
	{
		Name:   "GCP Vertex AI",
		Vendor: "vertexai",
		YAMLTemplate: `# GCP Vertex AI - uncomment and configure:
# llm:
#   models:
#     - text: "vertexai/gemini-1.5-pro"
#       project_id: "<project-id>"
#       region: "us-central1"`,
		NextSteps: []string{
			"Edit formation.yaml and uncomment the Vertex AI configuration",
			"Fill in your project ID and region",
			"Ensure GOOGLE_APPLICATION_CREDENTIALS is set",
		},
	},
}
