package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ViniciusToledoNunes/sezzle-calculator/backend/internal/calculator"
)

func TestCalculateEndpoint(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), "http://localhost:5173")
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantBody string
	}{
		{name: "success", body: `{"operation":"multiply","operands":[6,7]}`, wantCode: http.StatusOK, wantBody: `"result":42`},
		{name: "division by zero", body: `{"operation":"divide","operands":[6,0]}`, wantCode: http.StatusUnprocessableEntity, wantBody: `"code":"division_by_zero"`},
		{name: "negative square root", body: `{"operation":"square_root","operands":[-1]}`, wantCode: http.StatusUnprocessableEntity, wantBody: `"code":"invalid_square_root"`},
		{name: "undefined result", body: `{"operation":"power","operands":[-1,0.5]}`, wantCode: http.StatusUnprocessableEntity, wantBody: `"code":"undefined_result"`},
		{name: "wrong operands", body: `{"operation":"add","operands":[6]}`, wantCode: http.StatusUnprocessableEntity, wantBody: `"code":"invalid_operands"`},
		{name: "unknown operation", body: `{"operation":"modulo","operands":[6,2]}`, wantCode: http.StatusBadRequest, wantBody: `"code":"unknown_operation"`},
		{name: "malformed JSON", body: `{"operation":`, wantCode: http.StatusBadRequest, wantBody: `"code":"invalid_request"`},
		{name: "unknown field", body: `{"operation":"add","operands":[1,2],"surprise":true}`, wantCode: http.StatusBadRequest, wantBody: `"code":"invalid_request"`},
		{name: "multiple objects", body: `{}` + `{}`, wantCode: http.StatusBadRequest, wantBody: `"code":"invalid_request"`},
		{name: "trailing garbage", body: `{"operation":"add","operands":[1,2]} trailing`, wantCode: http.StatusBadRequest, wantBody: `"code":"invalid_request"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, test.wantCode, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), test.wantBody) {
				t.Errorf("body = %s, want to contain %s", recorder.Body.String(), test.wantBody)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
				t.Errorf("CORS origin = %q", got)
			}
		})
	}
}

func TestOptionsEndpoint(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), "http://localhost:5173")
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/calculate", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
		t.Errorf("allowed methods = %q", got)
	}
}

func TestSPAHandler(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "index.html"), []byte("calculator app"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "asset.txt"), []byte("asset"), 0o600); err != nil {
		t.Fatal(err)
	}
	handler := spaHandler(directory)

	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/", want: "calculator app"},
		{path: "/asset.txt", want: "asset"},
		{path: "/client-side-route", want: "calculator app"},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), test.want) {
			t.Errorf("GET %s: status=%d body=%q", test.path, recorder.Code, recorder.Body.String())
		}
	}
}

func TestNewHandlerServesConfiguredStaticDirectory(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "index.html"), []byte("configured app"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STATIC_DIR", directory)

	handler := NewHandler(nil, "")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/some-route", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "configured app") {
		t.Fatalf("configured static handler: status=%d body=%q", recorder.Code, recorder.Body.String())
	}
}

func TestHealthEndpoint(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response["status"] != "ok" {
		t.Errorf("status body = %q", response["status"])
	}
}

func TestCalculateEndpointUsesDependency(t *testing.T) {
	called := false
	handler := newHandler(nil, "", func(operation calculator.Operation, operands []float64) (float64, error) {
		called = true
		return 99, nil
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"operation":"add","operands":[1,2]}`))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if !called || !strings.Contains(recorder.Body.String(), `"result":99`) {
		t.Fatalf("injected calculator was not used: %s", recorder.Body.String())
	}
}
