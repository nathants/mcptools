/*
Package jsonutils implements JSON utility functions.
*/
package jsonutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
)

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

func formatToolsList(tools any) (string, error) {
	toolsSlice, ok := tools.([]any)
	if !ok {
		return "", fmt.Errorf("tools is not a slice")
	}

	if len(toolsSlice) == 0 {
		return "No tools available", nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-----------")

	for _, t := range toolsSlice {
		tool, ok1 := t.(map[string]any)
		if !ok1 {
			continue
		}

		name, _ := tool["name"].(string)
		desc, _ := tool["description"].(string)

		if len(desc) > 70 {
			desc = desc[:67] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\n", name, desc)
	}

	_ = w.Flush()
	return buf.String(), nil
}

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

	fmt.Fprintln(w, "NAME\tTYPE\tURI")
	fmt.Fprintln(w, "----\t----\t---")

	for _, r := range resourcesSlice {
		resource, ok1 := r.(map[string]any)
		if !ok1 {
			continue
		}

		name, _ := resource["name"].(string)
		resType, _ := resource["type"].(string)
		uri, _ := resource["uri"].(string)

		fmt.Fprintf(w, "%s\t%s\t%s\n", name, resType, uri)
	}

	_ = w.Flush()
	return buf.String(), nil
}

func formatPromptsList(prompts any) (string, error) {
	promptsSlice, ok := prompts.([]any)
	if !ok {
		return "", fmt.Errorf("prompts is not a slice")
	}

	if len(promptsSlice) == 0 {
		return "No prompts available", nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-----------")

	for _, p := range promptsSlice {
		prompt, ok1 := p.(map[string]any)
		if !ok1 {
			continue
		}

		name, _ := prompt["name"].(string)
		desc, _ := prompt["description"].(string)

		if len(desc) > 70 {
			desc = desc[:67] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\n", name, desc)
	}

	_ = w.Flush()
	return buf.String(), nil
}

func formatContent(content any) (string, error) {
	contentSlice, ok := content.([]any)
	if !ok {
		return "", fmt.Errorf("content is not a slice")
	}

	var buf strings.Builder

	for _, c := range contentSlice {
		contentItem, ok1 := c.(map[string]any)
		if !ok1 {
			continue
		}

		contentType, _ := contentItem["type"].(string)

		switch contentType {
		case "text":
			text, _ := contentItem["text"].(string)
			buf.WriteString(text)
		case "image":
			buf.WriteString("[IMAGE CONTENT]\n")
		default:
			buf.WriteString(fmt.Sprintf("[%s CONTENT]\n", strings.ToUpper(contentType)))
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

	fmt.Fprintln(w, "KEY\tVALUE")
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

		fmt.Fprintf(w, "%s\t%s\n", k, valueStr)
	}

	_ = w.Flush()
	return buf.String(), nil
}
