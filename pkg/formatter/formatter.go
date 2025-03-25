package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
	"reflect"
	"sort"
	"text/tabwriter"
	"bytes"
)

// Format formats the given data based on the output format
func Format(data interface{}, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json", "j":
		return formatJSON(data, false)
	case "pretty", "p":
		return formatJSON(data, true)
	case "table", "t":
		return formatTable(data)
	default:
		return formatTable(data)
	}
}

// formatJSON formats the data as JSON with optional pretty printing
func formatJSON(data interface{}, pretty bool) (string, error) {
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

// formatTable formats the data as a table-like view
func formatTable(data interface{}) (string, error) {
	// Handle special cases based on common MCP server responses
	val := reflect.ValueOf(data)
	
	// For nil values
	if !val.IsValid() {
		return "No data available", nil
	}
	
	// If it's not a map, just return the JSON representation
	if val.Kind() != reflect.Map {
		return formatJSON(data, true)
	}
	
	// Try to detect common MCP response structures
	mapVal := val.Interface().(map[string]interface{})
	
	// Handle tool list
	if tools, ok := mapVal["tools"]; ok {
		return formatToolsList(tools)
	}
	
	// Handle resource list
	if resources, ok := mapVal["resources"]; ok {
		return formatResourcesList(resources)
	}
	
	// Handle tool call with content
	if content, ok := mapVal["content"]; ok {
		return formatContent(content)
	}
	
	// Generic table for other map structures
	return formatGenericMap(mapVal)
}

// formatToolsList formats a list of tools as a table
func formatToolsList(tools interface{}) (string, error) {
	toolsSlice, ok := tools.([]interface{})
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
		tool, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		
		name, _ := tool["name"].(string)
		desc, _ := tool["description"].(string)
		
		// Truncate long descriptions
		if len(desc) > 70 {
			desc = desc[:67] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\n", name, desc)
	}
	
	w.Flush()
	return buf.String(), nil
}

// formatResourcesList formats a list of resources as a table
func formatResourcesList(resources interface{}) (string, error) {
	resourcesSlice, ok := resources.([]interface{})
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
		resource, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		
		name, _ := resource["name"].(string)
		resType, _ := resource["type"].(string)
		uri, _ := resource["uri"].(string)
		
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, resType, uri)
	}
	
	w.Flush()
	return buf.String(), nil
}

// formatContent formats content (usually tool call results) in a readable way
func formatContent(content interface{}) (string, error) {
	contentSlice, ok := content.([]interface{})
	if !ok {
		return "", fmt.Errorf("content is not a slice")
	}
	
	var buf strings.Builder
	
	for _, c := range contentSlice {
		contentItem, ok := c.(map[string]interface{})
		if !ok {
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

// formatGenericMap formats a generic map as a table with keys and values
func formatGenericMap(data map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "No data available", nil
	}
	
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, "KEY\tVALUE")
	fmt.Fprintln(w, "---\t-----")
	
	// Sort keys for consistent output
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
			// For complex types, use JSON
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				valueStr = fmt.Sprintf("<%T>", val)
			} else {
				valueStr = string(jsonBytes)
				// Truncate long values
				if len(valueStr) > 50 {
					valueStr = valueStr[:47] + "..."
				}
			}
		}
		
		fmt.Fprintf(w, "%s\t%s\n", k, valueStr)
	}
	
	w.Flush()
	return buf.String(), nil
} 