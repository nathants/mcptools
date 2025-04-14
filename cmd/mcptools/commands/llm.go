package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/f/mcptools/pkg/client"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

const (
	// LLMProviderOpenAI is the identifier for OpenAI
	LLMProviderOpenAI = "openai"
	// LLMProviderAnthropic is the identifier for Anthropic
	LLMProviderAnthropic = "anthropic"
	// LLMProviderMistral is the identifier for Mistral
	LLMProviderMistral = "mistral"
)

// Tool represents a tool available to the LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// LLMCmd creates the llm command.
func LLMCmd() *cobra.Command {
	var providerFlag string
	var modelFlag string
	var apiKeyFlag string

	cmd := &cobra.Command{
		Use:   "llm [command args...]",
		Short: "Start an interactive shell with LLM integration",
		Long: `Start an interactive shell with LLM integration.
This command connects to an LLM provider and provides a chat interface.
The LLM can execute MCP tools on your behalf.

Example usage:
  mcp llm npx -y @modelcontextprotocol/server-filesystem ~
  mcp llm npx -y @modelcontextprotocol/server-filesystem ~ --provider anthropic
  mcp llm npx -y @modelcontextprotocol/server-filesystem ~ --model gpt-4-turbo`,
		DisableFlagParsing: false,
		SilenceUsage:       true,
		RunE: func(thisCmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("command to execute is required when using llm")
			}

			// Set default provider and model
			provider := providerFlag
			if provider == "" {
				provider = LLMProviderOpenAI
			}

			model := modelFlag
			if model == "" {
				// Default models based on provider
				switch provider {
				case LLMProviderOpenAI:
					model = "gpt-4o"
				case LLMProviderAnthropic:
					model = "claude-3-opus-20240229"
				case LLMProviderMistral:
					model = "mistral-large-latest"
				}
			}

			// Get API key
			apiKey := apiKeyFlag
			if apiKey == "" {
				switch provider {
				case LLMProviderOpenAI:
					apiKey = os.Getenv("OPENAI_API_KEY")
				case LLMProviderAnthropic:
					apiKey = os.Getenv("ANTHROPIC_API_KEY")
				case LLMProviderMistral:
					apiKey = os.Getenv("MISTRAL_API_KEY")
				}
			}

			if apiKey == "" {
				return fmt.Errorf("no API key provided for %s. Please set the appropriate environment variable or use --api-key", provider)
			}

			// Create the MCP client
			mcpClient, clientErr := CreateClientFunc(args, client.CloseTransportAfterExecute(false))
			if clientErr != nil {
				return fmt.Errorf("error creating MCP client: %v", clientErr)
			}

			// Verify the connection to the MCP server
			toolsResp, listErr := mcpClient.ListTools()
			if listErr != nil {
				return fmt.Errorf("error connecting to MCP server: %v", listErr)
			}

			// Extract and format tools for LLM function/tool calling
			tools, err := formatToolsForLLM(toolsResp)
			if err != nil {
				return fmt.Errorf("error formatting tools: %v", err)
			}

			// Get the username for the prompt
			username := "user"
			if u := os.Getenv("USER"); u != "" {
				username = u
			} else if u := os.Getenv("USERNAME"); u != "" { // For Windows
				username = u
			}

			fmt.Fprintf(thisCmd.OutOrStdout(), "mcp > MCP LLM Shell (%s)\n", Version)
			fmt.Fprintf(thisCmd.OutOrStdout(), "mcp > Connected to Server: %s\n", strings.Join(args, " "))
			fmt.Fprintf(thisCmd.OutOrStdout(), "mcp > Using provider: %s, model: %s\n", provider, model)
			fmt.Fprintf(thisCmd.OutOrStdout(), "mcp > Registered %d tools with the LLM\n", len(tools))
			fmt.Fprintf(thisCmd.OutOrStdout(), "mcp > Type 'exit' to quit\n\n")

			// Initialize chat session with a special message type for proper handling
			messages := []map[string]interface{}{
				{
					"role": "system",
					"content": `You are a helpful AI assistant with access to tools.
Use these tools when appropriate to perform actions for the user.
Be concise, accurate, and helpful.`,
				},
			}

			// Setup liner for command history and line editing
			line := liner.NewLiner()
			line.SetCtrlCAborts(true)
			defer func() { _ = line.Close() }()

			// Setup history file
			defer setupHistory(line)()

			// Add simple completion for common commands
			line.SetCompleter(func(line string) (c []string) {
				commands := []string{
					"exit",
					"quit",
					"help",
				}
				for _, cmd := range commands {
					if strings.HasPrefix(cmd, line) {
						c = append(c, cmd)
					}
				}
				return
			})

			// Start interactive chat loop
			for {
				// Create prompt with username
				prompt := fmt.Sprintf("%s > ", username)

				// Use liner to get input with history and editing
				input, err := line.Prompt(prompt)
				if err != nil {
					if err == liner.ErrPromptAborted {
						fmt.Fprintf(thisCmd.OutOrStdout(), "\nExiting LLM shell\n")
						break
					}
					fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
					break
				}

				// Skip empty input
				if input == "" {
					continue
				}

				// Add to history
				line.AppendHistory(input)

				// Check for exit commands
				if input == "exit" || input == "quit" {
					fmt.Fprintf(thisCmd.OutOrStdout(), "Exiting LLM shell\n")
					break
				}

				// Add user message to the conversation
				messages = append(messages, map[string]interface{}{
					"role":    "user",
					"content": input,
				})

				// Call the LLM API and stream the response
				fmt.Print("agent > ")
				responseMessages, err := callLLMWithTools(provider, model, apiKey, messages, tools, mcpClient, thisCmd.OutOrStdout())
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nError calling LLM: %v\n", err)
					continue
				}

				// Add all messages returned by the function to the conversation
				messages = append(messages, responseMessages...)

				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providerFlag, "provider", "", "LLM provider (openai, anthropic, mistral)")
	cmd.Flags().StringVar(&modelFlag, "model", "", "The model to use")
	cmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key for the LLM provider")

	return cmd
}

// setupHistory sets up command history for the liner
func setupHistory(line *liner.State) func() {
	// Create history directory if it doesn't exist
	historyDir := filepath.Join(os.Getenv("HOME"), ".config", "mcp")
	_ = os.MkdirAll(historyDir, 0755)

	// History file path
	historyFile := filepath.Join(historyDir, ".mcp_llm_history")

	// Load history from file if it exists
	if f, err := os.Open(filepath.Clean(historyFile)); err == nil {
		_, _ = line.ReadHistory(f)
		_ = f.Close()
	}

	// Return a function that will save history on exit
	return func() {
		if f, err := os.Create(historyFile); err == nil {
			_, _ = line.WriteHistory(f)
			_ = f.Close()
		}
	}
}

// formatToolsForLLM converts MCP tools to a format suitable for LLM function calling
func formatToolsForLLM(toolsResp map[string]interface{}) ([]Tool, error) {
	toolsData, ok := toolsResp["tools"]
	if !ok {
		return nil, fmt.Errorf("no tools found in response")
	}

	toolsList, ok := toolsData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tools data is not in expected format")
	}

	var tools []Tool
	for _, item := range toolsList {
		toolMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := toolMap["name"].(string)
		description, _ := toolMap["description"].(string)

		// Extract parameters from schema if available
		var parameters map[string]interface{}
		if schema, ok := toolMap["schema"].(map[string]interface{}); ok {
			if props, ok := schema["properties"].(map[string]interface{}); ok {
				parameters = map[string]interface{}{
					"type":       "object",
					"properties": props,
				}

				// Add required fields if available
				if required, ok := schema["required"].([]interface{}); ok {
					var requiredArr []string
					for _, r := range required {
						if s, ok := r.(string); ok {
							requiredArr = append(requiredArr, s)
						}
					}
					parameters["required"] = requiredArr
				}
			}
		}

		// If no parameters defined, use empty object
		if parameters == nil {
			parameters = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		}

		tools = append(tools, Tool{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		})
	}

	return tools, nil
}

// callLLMWithTools calls the LLM API with tools support, processes tool calls, and returns all generated messages
func callLLMWithTools(provider, model, apiKey string, messages []map[string]interface{}, tools []Tool, mcpClient *client.Client, out io.Writer) ([]map[string]interface{}, error) {
	switch provider {
	case LLMProviderOpenAI:
		return callOpenAIWithTools(model, apiKey, messages, tools, mcpClient, out)
	case LLMProviderAnthropic:
		return callAnthropicWithTools(model, apiKey, messages, tools, mcpClient, out)
	case LLMProviderMistral:
		return callMistralWithTools(model, apiKey, messages, tools, mcpClient, out)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// convertMessages converts between different message formats if needed
func convertMessages(messages []map[string]interface{}) []map[string]interface{} {
	convertedMessages := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		convertedMsg := make(map[string]interface{})
		for k, v := range msg {
			convertedMsg[k] = v
		}
		convertedMessages = append(convertedMessages, convertedMsg)
	}
	return convertedMessages
}

// callOpenAIWithTools calls OpenAI API with tools support
func callOpenAIWithTools(model, apiKey string, messages []map[string]interface{}, tools []Tool, mcpClient *client.Client, out io.Writer) ([]map[string]interface{}, error) {
	var responseMessages []map[string]interface{}
	var fullContent strings.Builder

	// Format tools for OpenAI
	openAITools := []map[string]interface{}{}
	for _, tool := range tools {
		openAITools = append(openAITools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}

	// Convert messages to OpenAI format if needed
	requestMessages := convertMessages(messages)

	// Prepare request
	requestData := map[string]interface{}{
		"model":       model,
		"messages":    requestMessages,
		"tools":       openAITools,
		"tool_choice": "auto",
		"stream":      true,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)

	var toolCalls []map[string]interface{}
	var assistantMessage map[string]interface{}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")

			if line == "[DONE]" {
				break
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(line), &data); err != nil {
				continue
			}

			if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						// Initialize assistant message if needed
						if assistantMessage == nil {
							assistantMessage = map[string]interface{}{
								"role":       "assistant",
								"content":    "",
								"tool_calls": []map[string]interface{}{},
							}
						}

						// Handle content
						if content, ok := delta["content"].(string); ok {
							fmt.Fprint(out, content)
							fullContent.WriteString(content)
							currentContent, _ := assistantMessage["content"].(string)
							assistantMessage["content"] = currentContent + content
						}

						// Handle tool calls
						if toolCallDelta, ok := delta["tool_calls"].([]interface{}); ok && len(toolCallDelta) > 0 {
							for _, tc := range toolCallDelta {
								if tcDelta, ok := tc.(map[string]interface{}); ok {
									// Get the index for this tool call
									index, _ := tcDelta["index"].(float64)
									indexInt := int(index)

									// Ensure we have enough elements in toolCalls
									for len(toolCalls) <= indexInt {
										toolCalls = append(toolCalls, map[string]interface{}{
											"id":   "",
											"type": "function",
											"function": map[string]interface{}{
												"name":      "",
												"arguments": "",
											},
										})
									}

									// Ensure we have enough elements in the assistant message tool_calls
									toolCallsArr, _ := assistantMessage["tool_calls"].([]map[string]interface{})
									for len(toolCallsArr) <= indexInt {
										toolCallsArr = append(toolCallsArr, map[string]interface{}{
											"id":   "",
											"type": "function",
											"function": map[string]interface{}{
												"name":      "",
												"arguments": "",
											},
										})
									}
									assistantMessage["tool_calls"] = toolCallsArr

									// Update the tool call with this delta
									if id, ok := tcDelta["id"].(string); ok && id != "" {
										toolCalls[indexInt]["id"] = id
										toolCallsArr[indexInt]["id"] = id
									}

									if function, ok := tcDelta["function"].(map[string]interface{}); ok {
										toolCall := toolCalls[indexInt]["function"].(map[string]interface{})
										toolCallInAssistant := toolCallsArr[indexInt]["function"].(map[string]interface{})

										if name, ok := function["name"].(string); ok && name != "" {
											toolCall["name"] = name
											toolCallInAssistant["name"] = name

											// If this is the first time we're seeing the name, print it
											currentName, _ := toolCall["name"].(string)
											if currentName == name {
												fmt.Fprintf(out, "[Calling %s", name)
											}
										}

										if arguments, ok := function["arguments"].(string); ok && arguments != "" {
											current, _ := toolCall["arguments"].(string)
											toolCall["arguments"] = current + arguments

											currentInAssistant, _ := toolCallInAssistant["arguments"].(string)
											toolCallInAssistant["arguments"] = currentInAssistant + arguments
										}
									}
								}
							}
						}

						// Check if done with tool calls and need to execute them
						if finish, ok := choice["finish_reason"].(string); ok && finish == "tool_calls" {
							// Process and execute all tool calls
							fmt.Fprint(out, "]")

							// First add the assistant message to our responses with proper tool_calls
							responseMessages = append(responseMessages, assistantMessage)

							// Now handle each tool call and execute it
							var toolResponses []map[string]interface{}
							for _, toolCall := range toolCalls {
								functionData := toolCall["function"].(map[string]interface{})
								toolName := functionData["name"].(string)
								toolArgsStr := functionData["arguments"].(string)

								var toolArgs map[string]interface{}
								if err := json.Unmarshal([]byte(toolArgsStr), &toolArgs); err != nil {
									fmt.Fprintf(out, "\nmcp > Error parsing tool arguments: %v\n", err)
									continue
								}

								// Execute tool
								fmt.Fprintf(out, "\nmcp > [running %s with params %s]\n", toolName, toolArgsStr)

								result, err := mcpClient.CallTool(toolName, toolArgs)
								if err != nil {
									fmt.Fprintf(out, "mcp > Error executing tool: %v\n", err)
									toolResponses = append(toolResponses, map[string]interface{}{
										"role":         "tool",
										"tool_call_id": toolCall["id"].(string),
										"content":      fmt.Sprintf("Error: %v", err),
									})
								} else {
									// Format result as JSON
									resultBytes, _ := json.MarshalIndent(result, "", "  ")
									resultStr := string(resultBytes)

									fmt.Fprint(out, resultStr+"\n")

									toolResponses = append(toolResponses, map[string]interface{}{
										"role":         "tool",
										"tool_call_id": toolCall["id"].(string),
										"content":      resultStr,
									})
								}
							}

							// Add all tool responses
							responseMessages = append(responseMessages, toolResponses...)

							// Make another call to get the AI's final response after tools
							finalMessages, err := callOpenAIFinalResponse(model, apiKey, append(messages, responseMessages...), out)
							if err != nil {
								return nil, err
							}
							return append(responseMessages, finalMessages...), nil
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Add the assistant message if we didn't execute any tools
	if len(responseMessages) == 0 && fullContent.Len() > 0 {
		responseMessages = append(responseMessages, map[string]interface{}{
			"role":    "assistant",
			"content": fullContent.String(),
		})
	}

	return responseMessages, nil
}

// callOpenAIFinalResponse calls OpenAI to get a response after tools have been executed
func callOpenAIFinalResponse(model, apiKey string, messages []map[string]interface{}, out io.Writer) ([]map[string]interface{}, error) {
	requestData := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")

			if line == "[DONE]" {
				break
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(line), &data); err != nil {
				continue
			}

			if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fmt.Fprint(out, content)
							fullContent.WriteString(content)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if fullContent.Len() > 0 {
		return []map[string]interface{}{
			{
				"role":    "assistant",
				"content": fullContent.String(),
			},
		}, nil
	}

	return nil, nil
}

// callAnthropicWithTools calls Anthropic API with tools support
func callAnthropicWithTools(model, apiKey string, messages []map[string]interface{}, tools []Tool, mcpClient *client.Client, out io.Writer) ([]map[string]interface{}, error) {
	var responseMessages []map[string]interface{}
	var fullContent strings.Builder

	// Extract system message
	systemPrompt := ""
	var anthropicMessages []map[string]interface{}

	for _, msg := range messages {
		role, _ := msg["role"].(string)
		if role == "system" {
			content, _ := msg["content"].(string)
			systemPrompt = content
		} else {
			// Convert tool messages to appropriate format for Anthropic
			if role == "tool" {
				name, _ := msg["name"].(string)
				content, _ := msg["content"].(string)
				anthropicMessages = append(anthropicMessages, map[string]interface{}{
					"role":    "assistant",
					"content": fmt.Sprintf("Tool result from %s: %s", name, content),
				})
			} else {
				anthropicMessages = append(anthropicMessages, msg)
			}
		}
	}

	// Format tools for Anthropic
	anthropicTools := []map[string]interface{}{}
	for _, tool := range tools {
		anthropicTools = append(anthropicTools, map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.Parameters,
		})
	}

	// Prepare request
	requestData := map[string]interface{}{
		"model":    model,
		"messages": anthropicMessages,
		"system":   systemPrompt,
		"tools":    anthropicTools,
		"stream":   true,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	var toolUse map[string]interface{}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(line), &data); err != nil {
				continue
			}

			// Handle content blocks
			if contentBlocks, ok := data["delta"].(map[string]interface{})["content_blocks"].([]interface{}); ok && len(contentBlocks) > 0 {
				for _, block := range contentBlocks {
					if contentBlock, ok := block.(map[string]interface{}); ok {
						if text, ok := contentBlock["text"].(string); ok {
							fmt.Fprint(out, text)
							fullContent.WriteString(text)
						}
					}
				}
			}

			// Handle tool use
			if delta, ok := data["delta"].(map[string]interface{}); ok {
				if usage, ok := delta["tool_use"].(map[string]interface{}); ok {
					// Initialize tool use if needed
					if toolUse == nil {
						toolUse = map[string]interface{}{
							"id":    usage["id"],
							"type":  "tool_use",
							"name":  usage["name"],
							"input": "",
						}

						// Print tool name when we first encounter it
						fmt.Fprintf(out, "[Calling %s", toolUse["name"])
					}

					// Append to input if present
					if input, ok := usage["input"].(string); ok {
						current := toolUse["input"].(string)
						toolUse["input"] = current + input
					}
				}
			}

			// Handle message stop
			if typ, ok := data["type"].(string); ok && typ == "message_stop" {
				if toolUse != nil {
					// We have a tool to execute
					fmt.Fprint(out, "]")

					toolName := toolUse["name"].(string)
					toolInputStr := toolUse["input"].(string)

					var toolArgs map[string]interface{}
					if err := json.Unmarshal([]byte(toolInputStr), &toolArgs); err != nil {
						fmt.Fprintf(out, "\nmcp > Error parsing tool arguments: %v\n", err)
					} else {
						// Execute tool
						fmt.Fprintf(out, "\nmcp > [running %s with params %s]\n", toolName, toolInputStr)

						result, err := mcpClient.CallTool(toolName, toolArgs)
						if err != nil {
							fmt.Fprintf(out, "mcp > Error executing tool: %v\n", err)
						} else {
							// Format result as JSON
							resultBytes, _ := json.MarshalIndent(result, "", "  ")
							resultStr := string(resultBytes)

							fmt.Fprint(out, resultStr+"\n")

							// Add assistant message
							responseMessages = append(responseMessages, map[string]interface{}{
								"role":    "assistant",
								"content": fullContent.String(),
							})

							// Add tool result as a message
							responseMessages = append(responseMessages, map[string]interface{}{
								"role":    "user",
								"content": fmt.Sprintf("Tool result from %s: %s", toolName, resultStr),
							})

							// Get the final response after tool execution
							finalMessages, err := callAnthropicFinalResponse(model, apiKey, append(messages, responseMessages...), out)
							if err != nil {
								return nil, err
							}

							return append(responseMessages, finalMessages...), nil
						}
					}
				}

				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// If we got here without executing tools, return the full content
	if fullContent.Len() > 0 {
		responseMessages = append(responseMessages, map[string]interface{}{
			"role":    "assistant",
			"content": fullContent.String(),
		})
	}

	return responseMessages, nil
}

// callAnthropicFinalResponse calls Anthropic to get final response after tool execution
func callAnthropicFinalResponse(model, apiKey string, messages []map[string]interface{}, out io.Writer) ([]map[string]interface{}, error) {
	// Extract system message and format messages
	systemPrompt := ""
	var anthropicMessages []map[string]interface{}

	for _, msg := range messages {
		role, _ := msg["role"].(string)
		if role == "system" {
			content, _ := msg["content"].(string)
			systemPrompt = content
		} else {
			anthropicMessages = append(anthropicMessages, msg)
		}
	}

	// Prepare request
	requestData := map[string]interface{}{
		"model":    model,
		"messages": anthropicMessages,
		"system":   systemPrompt,
		"stream":   true,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(line), &data); err != nil {
				continue
			}

			// Handle content blocks
			if contentBlocks, ok := data["delta"].(map[string]interface{})["content_blocks"].([]interface{}); ok && len(contentBlocks) > 0 {
				for _, block := range contentBlocks {
					if contentBlock, ok := block.(map[string]interface{}); ok {
						if text, ok := contentBlock["text"].(string); ok {
							fmt.Fprint(out, text)
							fullContent.WriteString(text)
						}
					}
				}
			}

			// Handle message stop
			if typ, ok := data["type"].(string); ok && typ == "message_stop" {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if fullContent.Len() > 0 {
		return []map[string]interface{}{
			{
				"role":    "assistant",
				"content": fullContent.String(),
			},
		}, nil
	}

	return nil, nil
}

// callMistralWithTools calls Mistral API with tools support
func callMistralWithTools(model, apiKey string, messages []map[string]interface{}, tools []Tool, mcpClient *client.Client, out io.Writer) ([]map[string]interface{}, error) {
	var responseMessages []map[string]interface{}
	var fullContent strings.Builder

	// Format tools for Mistral
	mistralTools := []map[string]interface{}{}
	for _, tool := range tools {
		mistralTools = append(mistralTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}

	// Prepare request
	requestData := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"tools":       mistralTools,
		"tool_choice": "auto",
		"stream":      true,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)

	var toolCalls []map[string]interface{}
	var assistantMessage map[string]interface{}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")

			if line == "[DONE]" {
				break
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(line), &data); err != nil {
				continue
			}

			if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						// Initialize assistant message if needed
						if assistantMessage == nil {
							assistantMessage = map[string]interface{}{
								"role":       "assistant",
								"content":    "",
								"tool_calls": []map[string]interface{}{},
							}
						}

						// Handle content
						if content, ok := delta["content"].(string); ok && content != "" {
							fmt.Fprint(out, content)
							fullContent.WriteString(content)
							currentContent, _ := assistantMessage["content"].(string)
							assistantMessage["content"] = currentContent + content
						}

						// Handle tool calls
						if toolCallDelta, ok := delta["tool_calls"].([]interface{}); ok && len(toolCallDelta) > 0 {
							for _, tc := range toolCallDelta {
								if tcDelta, ok := tc.(map[string]interface{}); ok {
									// Get the index for this tool call
									index, _ := tcDelta["index"].(float64)
									indexInt := int(index)

									// Ensure we have enough elements in toolCalls
									for len(toolCalls) <= indexInt {
										toolCalls = append(toolCalls, map[string]interface{}{
											"id":   "",
											"type": "function",
											"function": map[string]interface{}{
												"name":      "",
												"arguments": "",
											},
										})
									}

									// Ensure we have enough elements in the assistant message tool_calls
									toolCallsArr, _ := assistantMessage["tool_calls"].([]map[string]interface{})
									for len(toolCallsArr) <= indexInt {
										toolCallsArr = append(toolCallsArr, map[string]interface{}{
											"id":   "",
											"type": "function",
											"function": map[string]interface{}{
												"name":      "",
												"arguments": "",
											},
										})
									}
									assistantMessage["tool_calls"] = toolCallsArr

									// Update the tool call with this delta
									if id, ok := tcDelta["id"].(string); ok && id != "" {
										toolCalls[indexInt]["id"] = id
										toolCallsArr[indexInt]["id"] = id
									}

									if function, ok := tcDelta["function"].(map[string]interface{}); ok {
										toolCall := toolCalls[indexInt]["function"].(map[string]interface{})
										toolCallInAssistant := toolCallsArr[indexInt]["function"].(map[string]interface{})

										if name, ok := function["name"].(string); ok && name != "" {
											toolCall["name"] = name
											toolCallInAssistant["name"] = name

											// Print tool name when first encountered
											currentName, _ := toolCall["name"].(string)
											if currentName == name {
												fmt.Fprintf(out, "[Calling %s", name)
											}
										}

										if arguments, ok := function["arguments"].(string); ok && arguments != "" {
											current, _ := toolCall["arguments"].(string)
											toolCall["arguments"] = current + arguments

											currentInAssistant, _ := toolCallInAssistant["arguments"].(string)
											toolCallInAssistant["arguments"] = currentInAssistant + arguments
										}
									}
								}
							}
						}

						// Process tool calls when done
						if finish, ok := choice["finish_reason"].(string); ok && finish == "tool_calls" {
							// We have tool calls to execute
							fmt.Fprint(out, "]")

							// First add the assistant message to the conversation
							responseMessages = append(responseMessages, assistantMessage)

							var toolResponses []map[string]interface{}
							for _, toolCall := range toolCalls {
								functionData := toolCall["function"].(map[string]interface{})
								toolName := functionData["name"].(string)
								toolArgsStr := functionData["arguments"].(string)

								var toolArgs map[string]interface{}
								if err := json.Unmarshal([]byte(toolArgsStr), &toolArgs); err != nil {
									fmt.Fprintf(out, "\nmcp > Error parsing tool arguments: %v\n", err)
									continue
								}

								// Execute tool
								fmt.Fprintf(out, "\nmcp > [running %s with params %s]\n", toolName, toolArgsStr)

								result, err := mcpClient.CallTool(toolName, toolArgs)
								if err != nil {
									fmt.Fprintf(out, "mcp > Error executing tool: %v\n", err)
									toolResponses = append(toolResponses, map[string]interface{}{
										"role":         "tool",
										"tool_call_id": toolCall["id"].(string),
										"content":      fmt.Sprintf("Error: %v", err),
									})
								} else {
									// Format result as JSON
									resultBytes, _ := json.MarshalIndent(result, "", "  ")
									resultStr := string(resultBytes)

									fmt.Fprint(out, resultStr+"\n")

									toolResponses = append(toolResponses, map[string]interface{}{
										"role":         "tool",
										"tool_call_id": toolCall["id"].(string),
										"content":      resultStr,
									})
								}
							}

							// Add all tool responses
							responseMessages = append(responseMessages, toolResponses...)

							// Get final response
							finalMessages, err := callMistralFinalResponse(model, apiKey, append(messages, responseMessages...), out)
							if err != nil {
								return nil, err
							}

							return append(responseMessages, finalMessages...), nil
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// If we got here without executing tools, return the full content
	if fullContent.Len() > 0 {
		responseMessages = append(responseMessages, map[string]interface{}{
			"role":    "assistant",
			"content": fullContent.String(),
		})
	}

	return responseMessages, nil
}

// callMistralFinalResponse calls Mistral to get a final response after tools have been executed
func callMistralFinalResponse(model, apiKey string, messages []map[string]interface{}, out io.Writer) ([]map[string]interface{}, error) {
	requestData := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")

			if line == "[DONE]" {
				break
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(line), &data); err != nil {
				continue
			}

			if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fmt.Fprint(out, content)
							fullContent.WriteString(content)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if fullContent.Len() > 0 {
		return []map[string]interface{}{
			{
				"role":    "assistant",
				"content": fullContent.String(),
			},
		}, nil
	}

	return nil, nil
}
