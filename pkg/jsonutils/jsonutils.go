/*
Package jsonutils implements JSON utility functions.
*/
package jsonutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
)

// isTerminal determines if stdout is a terminal (for colorized output)
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// OutputFormat represents the available output format options.
type OutputFormat string

// constants.
const (
	FormatJSON   OutputFormat = "json"
	FormatPretty OutputFormat = "pretty"
	FormatTable  OutputFormat = "table"
)

// ParseFormat converts a string to an OutputFormat.
func ParseFormat(format string) OutputFormat {
	switch strings.ToLower(format) {
	case "json", "j":
		return FormatJSON
	case "pretty", "p":
		return FormatPretty
	case "table", "t":
		return FormatTable
	default:
		return FormatTable
	}
}

// Format formats the given data according to the specified output format.
func Format(data any, format string) (string, error) {
	outputFormat := ParseFormat(format)

	switch outputFormat {
	case FormatJSON:
		return formatJSON(data, false)
	case FormatPretty:
		return formatJSON(data, true)
	case FormatTable:
		return formatTable(data)
	default:
		return formatTable(data)
	}
}

// formatJSON converts data to JSON with optional pretty printing.
func formatJSON(data any, pretty bool) (string, error) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %w", err)
	}

	return string(output), nil
}

// formatTable formats the data as a tabular view based on its structure.
// It tries to detect common MCP response structures and format them appropriately.
func formatTable(data any) (string, error) {
	val := reflect.ValueOf(data)

	if !val.IsValid() {
		return "No data available", nil
	}

	if val.Kind() != reflect.Map {
		return formatJSON(data, true)
	}

	mapVal, ok := val.Interface().(map[string]any)
	if !ok {
		return formatJSON(data, true)
	}

	if tools, ok1 := mapVal["tools"]; ok1 {
		return formatToolsList(tools)
	}

	if resources, ok2 := mapVal["resources"]; ok2 {
		return formatResourcesList(resources)
	}

	if prompts, ok3 := mapVal["prompts"]; ok3 {
		return formatPromptsList(prompts)
	}

	if content, ok4 := mapVal["content"]; ok4 {
		return formatContent(content)
	}

	return formatGenericMap(mapVal)
}

// formatToolsList formats a list of tools as a man-like page.
func formatToolsList(tools any) (string, error) {
	toolsSlice, ok := tools.([]any)
	if !ok {
		return "", fmt.Errorf("tools is not a slice")
	}

	if len(toolsSlice) == 0 {
		return "No tools available", nil
	}

	var buf bytes.Buffer
	termWidth := getTermWidth()
	descIndent := "     " // 5 spaces for description indentation
	descWidth := termWidth - len(descIndent)
	useColors := isTerminal()

	for i, t := range toolsSlice {
		tool, ok1 := t.(map[string]any)
		if !ok1 {
			continue
		}

		name, _ := tool["name"].(string)
		desc, _ := tool["description"].(string)

		// Format name with parameters if available
		displayName := name
		hasParams := false

		// Check for inputSchema (new format)
		if inputSchema, hasSchema := tool["inputSchema"]; hasSchema && inputSchema != nil {
			paramsStr := formatParameters(inputSchema)
			if paramsStr != "" {
				displayName = formatToolNameWithParams(name, paramsStr, useColors)
				hasParams = true
			}
		}

		// Fallback to old format if no params found yet
		if !hasParams {
			if params, hasParamsField := tool["parameters"]; hasParamsField && params != nil {
				paramsStr := formatParameters(params)
				if paramsStr != "" {
					displayName = formatToolNameWithParams(name, paramsStr, useColors)
					hasParams = true
				}
			}
		}

		// If no parameters were found or they were empty, just color the name
		if !hasParams && useColors {
			displayName = fmt.Sprintf("%s%s%s", ColorBold+ColorCyan, name, ColorReset)
		}

		// Write the name with parameters
		fmt.Fprintln(&buf, displayName)

		// Write the indented description
		if desc != "" {
			lines := wrapText(desc, descWidth)
			for _, line := range lines {
				if useColors {
					fmt.Fprintf(&buf, "%s%s%s%s\n", descIndent, ColorGray, line, ColorReset)
				} else {
					fmt.Fprintf(&buf, "%s%s\n", descIndent, line)
				}
			}
		}

		// Add blank line between tools, but not after the last one
		if i < len(toolsSlice)-1 {
			fmt.Fprintln(&buf)
		}
	}

	return buf.String(), nil
}

// formatToolNameWithParams formats a tool name with parameters, adding colors if enabled
func formatToolNameWithParams(name, params string, useColors bool) string {
	if !useColors {
		return fmt.Sprintf("%s(%s)", name, params)
	}

	// Parse parameters to add colors to required and optional params
	coloredParams := params
	coloredParams = strings.ReplaceAll(coloredParams, "[", ColorYellow+"[")
	coloredParams = strings.ReplaceAll(coloredParams, "]", "]"+ColorReset+ColorGreen)

	// Add any final reset if needed
	if strings.HasSuffix(coloredParams, ColorGreen) {
		coloredParams += ColorReset
	}

	// Return the full colored string
	return fmt.Sprintf("%s%s%s(%s%s%s)",
		ColorBold+ColorCyan, name, ColorReset,
		ColorGreen, coloredParams, ColorReset)
}

// Shortens type names for display
func shortenTypeName(typeName string) string {
	switch typeName {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	case "number":
		return "num"
	case "object":
		return "obj"
	default:
		// If it's already 3 letters or less, return as is
		if len(typeName) <= 3 {
			return typeName
		}
		// Otherwise return first 3 letters
		return typeName[:3]
	}
}

// formatObjectProperties formats object properties recursively
func formatObjectProperties(propMap map[string]any, requiredProps []string) string {
	if len(propMap) == 0 {
		return "obj"
	}

	// Get all property names and sort them for consistent output
	var propNames []string
	for name := range propMap {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	var props []string
	for _, name := range propNames {
		propDef, ok := propMap[name].(map[string]any)
		if !ok {
			continue
		}

		propType, _ := propDef["type"].(string)

		// Handle nested objects
		if propType == "object" {
			var nestedRequired []string
			if req, hasReq := propDef["required"]; hasReq && req != nil {
				if reqArray, ok := req.([]any); ok {
					for _, r := range reqArray {
						if reqStr, ok := r.(string); ok {
							nestedRequired = append(nestedRequired, reqStr)
						}
					}
				}
			}

			if properties, hasProps := propDef["properties"]; hasProps && properties != nil {
				if propsMap, ok := properties.(map[string]any); ok {
					propType = formatObjectProperties(propsMap, nestedRequired)
				}
			} else {
				propType = "obj"
			}
		} else if propType == "array" {
			// Handle array types
			if items, hasItems := propDef["items"]; hasItems && items != nil {
				if itemsMap, ok := items.(map[string]any); ok {
					itemType, hasType := itemsMap["type"]
					if hasType {
						if itemTypeStr, ok := itemType.(string); ok {
							if itemTypeStr == "object" {
								// Handle array of objects
								var nestedRequired []string
								if req, hasReq := itemsMap["required"]; hasReq && req != nil {
									if reqArray, ok := req.([]any); ok {
										for _, r := range reqArray {
											if reqStr, ok := r.(string); ok {
												nestedRequired = append(nestedRequired, reqStr)
											}
										}
									}
								}

								if properties, hasProps := itemsMap["properties"]; hasProps && properties != nil {
									if propsMap, ok := properties.(map[string]any); ok {
										propType = formatObjectProperties(propsMap, nestedRequired) + "[]"
									}
								} else {
									propType = "obj[]"
								}
							} else {
								// Simple array
								propType = shortenTypeName(itemTypeStr) + "[]"
							}
						}
					}
				}
			} else {
				propType = "arr"
			}
		} else {
			// Regular types
			propType = shortenTypeName(propType)
		}

		props = append(props, fmt.Sprintf("%s:%s", name, propType))
	}

	return "{" + strings.Join(props, ",") + "}"
}

// formatParameters formats the parameters for display in the tool name
func formatParameters(params any) string {
	// Handle case where we have an inputSchema structure
	if inputSchema, ok := params.(map[string]any); ok {
		// Check if this is the JSON Schema structure
		if properties, hasProps := inputSchema["properties"]; hasProps && properties != nil {
			propsMap, ok := properties.(map[string]any)
			if !ok {
				return ""
			}

			// Get required parameters
			var requiredParams []string
			if required, hasRequired := inputSchema["required"]; hasRequired && required != nil {
				if reqArray, ok := required.([]any); ok {
					for _, r := range reqArray {
						if reqStr, ok := r.(string); ok {
							requiredParams = append(requiredParams, reqStr)
						}
					}
				}
			}

			// Get all parameter names and sort them for consistent output
			var paramNames []string
			for name := range propsMap {
				paramNames = append(paramNames, name)
			}

			// Sort parameter names
			sort.Strings(paramNames)

			// Format parameters, putting required ones first
			var requiredParamStrs []string
			var optionalParamStrs []string

			for _, paramName := range paramNames {
				propDef, _ := propsMap[paramName]
				propDefMap, ok := propDef.(map[string]any)
				if !ok {
					continue
				}

				paramType, _ := propDefMap["type"].(string)

				// Handle object types
				if paramType == "object" {
					// Get nested required fields
					var nestedRequired []string
					if req, hasReq := propDefMap["required"]; hasReq && req != nil {
						if reqArray, ok := req.([]any); ok {
							for _, r := range reqArray {
								if reqStr, ok := r.(string); ok {
									nestedRequired = append(nestedRequired, reqStr)
								}
							}
						}
					}

					// Format object properties
					if properties, hasProps := propDefMap["properties"]; hasProps && properties != nil {
						if propsMap, ok := properties.(map[string]any); ok {
							paramType = formatObjectProperties(propsMap, nestedRequired)
						}
					} else {
						paramType = "obj"
					}
				} else if paramType == "array" {
					// Handle array types
					if items, hasItems := propDefMap["items"]; hasItems && items != nil {
						if itemsMap, ok := items.(map[string]any); ok {
							itemType, hasType := itemsMap["type"]
							if hasType {
								if itemTypeStr, ok := itemType.(string); ok {
									if itemTypeStr == "object" {
										// Handle array of objects
										var nestedRequired []string
										if req, hasReq := itemsMap["required"]; hasReq && req != nil {
											if reqArray, ok := req.([]any); ok {
												for _, r := range reqArray {
													if reqStr, ok := r.(string); ok {
														nestedRequired = append(nestedRequired, reqStr)
													}
												}
											}
										}

										if properties, hasProps := itemsMap["properties"]; hasProps && properties != nil {
											if propsMap, ok := properties.(map[string]any); ok {
												paramType = formatObjectProperties(propsMap, nestedRequired) + "[]"
											}
										} else {
											paramType = "obj[]"
										}
									} else {
										// Simple array
										paramType = shortenTypeName(itemTypeStr) + "[]"
									}
								}
							}
						}
					} else {
						// If no item type is specified, just use "array"
						paramType = "arr"
					}
				} else {
					// Shorten non-array type names
					paramType = shortenTypeName(paramType)
				}

				// Check if this parameter is required
				isRequired := false
				for _, req := range requiredParams {
					if req == paramName {
						isRequired = true
						break
					}
				}

				if isRequired {
					requiredParamStrs = append(requiredParamStrs, fmt.Sprintf("%s:%s", paramName, paramType))
				} else {
					optionalParamStrs = append(optionalParamStrs, fmt.Sprintf("[%s:%s]", paramName, paramType))
				}
			}

			// Join all parameters, required first, then optional
			var allParamStrs []string
			allParamStrs = append(allParamStrs, requiredParamStrs...)
			allParamStrs = append(allParamStrs, optionalParamStrs...)

			return strings.Join(allParamStrs, ", ")
		}
	}

	// Original function for other parameter formats
	switch p := params.(type) {
	case string:
		// If parameters is already a string (e.g., "param1:type1,param2:type2")
		// Add spaces after commas if they don't exist
		if !strings.Contains(p, ", ") && strings.Contains(p, ",") {
			return strings.ReplaceAll(p, ",", ", ")
		}
		return p
	case []any:
		// If parameters is an array of parameter objects
		var paramStrs []string
		for _, param := range p {
			if paramObj, ok := param.(map[string]any); ok {
				name, _ := paramObj["name"].(string)
				paramType, _ := paramObj["type"].(string)
				if name != "" {
					if paramType != "" {
						// Shorten the type name
						paramType = shortenTypeName(paramType)
						paramStrs = append(paramStrs, fmt.Sprintf("%s:%s", name, paramType))
					} else {
						paramStrs = append(paramStrs, name)
					}
				}
			}
		}
		return strings.Join(paramStrs, ", ")
	case map[string]any:
		// If parameters is a map of parameter definitions
		var paramNames []string
		for name := range p {
			paramNames = append(paramNames, name)
		}
		sort.Strings(paramNames)

		var paramStrs []string
		for _, name := range paramNames {
			paramType := p[name]
			if typeStr, ok := paramType.(string); ok {
				// Shorten the type name
				typeStr = shortenTypeName(typeStr)
				paramStrs = append(paramStrs, fmt.Sprintf("%s:%s", name, typeStr))
			} else {
				paramStrs = append(paramStrs, name)
			}
		}
		return strings.Join(paramStrs, ", ")
	default:
		return ""
	}
}

// getTermWidth returns the terminal width or a default value if detection fails.
func getTermWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // Default width if terminal width cannot be determined
	}
	return width
}

// wrapText wraps text to fit within a specified width.
func wrapText(text string, width int) []string {
	if text == "" {
		return []string{}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		switch {
		case len(currentLine) == 0:
			currentLine = word
		case len(currentLine)+len(word)+1 > width:
			// Add current line to lines and start a new line
			lines = append(lines, currentLine)
			currentLine = word
		default:
			currentLine += " " + word
		}
	}

	// Add the last line
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

// formatResourcesList formats a list of resources as a table.
func formatResourcesList(resources any) (string, error) {
	resourcesSlice, ok := resources.([]any)
	if !ok {
		return "", fmt.Errorf("resources is not a slice")
	}

	if len(resourcesSlice) == 0 {
		return "No resources available", nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	useColors := isTerminal()

	if useColors {
		fmt.Fprintf(w, "%s%sNAME%s\t%sTYPE%s\t%sURI%s\n",
			ColorBold, ColorCyan, ColorReset,
			ColorCyan, ColorReset,
			ColorCyan, ColorReset)
	} else {
		fmt.Fprintln(w, "NAME\tTYPE\tURI")
	}

	fmt.Fprintln(w, "----\t----\t---")

	for _, r := range resourcesSlice {
		resource, ok1 := r.(map[string]any)
		if !ok1 {
			continue
		}

		name, _ := resource["name"].(string)
		resType, _ := resource["type"].(string)
		uri, _ := resource["uri"].(string)

		// Use the entire URI instead of truncating
		if useColors {
			fmt.Fprintf(w, "%s%s%s\t%s%s\t%s%s%s\n",
				ColorGreen, name, ColorReset,
				resType, ColorReset,
				ColorYellow, uri, ColorReset)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, resType, uri)
		}
	}

	_ = w.Flush()
	return buf.String(), nil
}

// formatPromptsList formats a list of prompts as a table.
func formatPromptsList(prompts any) (string, error) {
	promptsSlice, ok := prompts.([]any)
	if !ok {
		return "", fmt.Errorf("prompts is not a slice")
	}

	if len(promptsSlice) == 0 {
		return "No prompts available", nil
	}

	var buf bytes.Buffer
	termWidth := getTermWidth()
	descIndent := "     " // 5 spaces for description indentation
	descWidth := termWidth - len(descIndent)
	useColors := isTerminal()

	for i, p := range promptsSlice {
		prompt, ok1 := p.(map[string]any)
		if !ok1 {
			continue
		}

		name, _ := prompt["name"].(string)
		desc, _ := prompt["description"].(string)

		// Write the prompt name
		if useColors {
			fmt.Fprintf(&buf, "%s%s%s\n", ColorBold+ColorCyan, name, ColorReset)
		} else {
			fmt.Fprintln(&buf, name)
		}

		// Write the indented description
		if desc != "" {
			lines := wrapText(desc, descWidth)
			for _, line := range lines {
				if useColors {
					fmt.Fprintf(&buf, "%s%s%s%s\n", descIndent, ColorGray, line, ColorReset)
				} else {
					fmt.Fprintf(&buf, "%s%s\n", descIndent, line)
				}
			}
		}

		// Add blank line between prompts, but not after the last one
		if i < len(promptsSlice)-1 {
			fmt.Fprintln(&buf)
		}
	}

	return buf.String(), nil
}

func formatContent(content any) (string, error) {
	contentSlice, ok := content.([]any)
	if !ok {
		return "", fmt.Errorf("content is not a slice")
	}

	var buf strings.Builder
	useColors := isTerminal()

	for _, c := range contentSlice {
		contentItem, ok1 := c.(map[string]any)
		if !ok1 {
			continue
		}

		contentType, _ := contentItem["type"].(string)

		switch contentType {
		case "text":
			text, _ := contentItem["text"].(string)
			if useColors {
				buf.WriteString(ColorGray + text + ColorReset)
			} else {
				buf.WriteString(text)
			}
		case "image":
			if useColors {
				buf.WriteString(ColorYellow + "[IMAGE CONTENT]" + ColorReset + "\n")
			} else {
				buf.WriteString("[IMAGE CONTENT]\n")
			}
		default:
			if useColors {
				buf.WriteString(fmt.Sprintf("%s[%s CONTENT]%s\n",
					ColorYellow, strings.ToUpper(contentType), ColorReset))
			} else {
				buf.WriteString(fmt.Sprintf("[%s CONTENT]\n", strings.ToUpper(contentType)))
			}
		}
	}

	return buf.String(), nil
}

func formatGenericMap(data map[string]any) (string, error) {
	if len(data) == 0 {
		return "No data available", nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	useColors := isTerminal()

	if useColors {
		fmt.Fprintf(w, "%s%sKEY%s\t%sVALUE%s\n",
			ColorBold, ColorCyan, ColorReset,
			ColorCyan, ColorReset)
	} else {
		fmt.Fprintln(w, "KEY\tVALUE")
	}

	fmt.Fprintln(w, "---\t-----")

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		var valueStr string

		switch val := v.(type) {
		case string:
			valueStr = val
		case nil:
			valueStr = "<nil>"
		default:
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				valueStr = fmt.Sprintf("<%T>", val)
			} else {
				valueStr = string(jsonBytes)
				if len(valueStr) > 50 {
					valueStr = valueStr[:47] + "..."
				}
			}
		}

		if useColors {
			fmt.Fprintf(w, "%s%s%s\t%s%s%s\n",
				ColorGreen, k, ColorReset,
				ColorYellow, valueStr, ColorReset)
		} else {
			fmt.Fprintf(w, "%s\t%s\n", k, valueStr)
		}
	}

	_ = w.Flush()
	return buf.String(), nil
}

// NormalizeParameterType converts common type names to their canonical form.
// This is used to accept alternative type names (like "str" for "string").
func NormalizeParameterType(typeName string) string {
	typeName = strings.ToLower(typeName)

	// Map of alternative type names to canonical forms
	typeMap := map[string]string{
		// String types
		"str":     "string",
		"text":    "string",
		"char":    "string",
		"varchar": "string",

		// Integer types
		"integer":  "int",
		"long":     "int",
		"short":    "int",
		"byte":     "int",
		"bigint":   "int",
		"smallint": "int",

		// Float types
		"double":  "float",
		"decimal": "float",
		"number":  "float",
		"real":    "float",

		// Boolean types
		"boolean": "bool",
		"bit":     "bool",
		"flag":    "bool",
	}

	if canonical, exists := typeMap[typeName]; exists {
		return canonical
	}

	return typeName
}
