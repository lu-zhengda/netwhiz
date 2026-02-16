package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	err := printJSON(data)
	if err != nil {
		t.Fatalf("printJSON returned error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if parsed["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", parsed["key"])
	}
}

func TestPrintJSON_Struct(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	err := printJSON(sample{Name: "test", Count: 42})
	if err != nil {
		t.Fatalf("printJSON returned error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var parsed sample
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.Name != "test" || parsed.Count != 42 {
		t.Errorf("unexpected output: %+v", parsed)
	}
}

func TestFprintJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"hello": "world"}

	if err := fprintJSON(&buf, data); err != nil {
		t.Fatalf("fprintJSON returned error: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if parsed["hello"] != "world" {
		t.Errorf("expected hello=world, got hello=%s", parsed["hello"])
	}
}

func TestFprintJSON_Indented(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]int{"a": 1}

	if err := fprintJSON(&buf, data); err != nil {
		t.Fatalf("fprintJSON returned error: %v", err)
	}

	output := buf.String()
	// Verify indentation (2-space indent).
	if !bytes.Contains([]byte(output), []byte("  \"a\"")) {
		t.Errorf("expected indented output, got: %s", output)
	}
}
