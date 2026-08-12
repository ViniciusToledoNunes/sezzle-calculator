package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ViniciusToledoNunes/sezzle-calculator/backend/internal/calculator"
)

const maxRequestBody = 1 << 20

type calculateFunc func(calculator.Operation, []float64) (float64, error)

type Handler struct {
	calculate     calculateFunc
	logger        *slog.Logger
	allowedOrigin string
}

type calculateRequest struct {
	Operation calculator.Operation `json:"operation"`
	Operands  []float64            `json:"operands"`
}

type calculateResponse struct {
	Operation calculator.Operation `json:"operation"`
	Operands  []float64            `json:"operands"`
	Result    float64              `json:"result"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewHandler(logger *slog.Logger, allowedOrigin string) http.Handler {
	return newHandler(logger, allowedOrigin, calculator.Calculate)
}

func newHandler(logger *slog.Logger, allowedOrigin string, calculate calculateFunc) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{calculate: calculate, logger: logger, allowedOrigin: allowedOrigin}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.health)
	mux.HandleFunc("POST /api/v1/calculate", handler.handleCalculate)
	mux.HandleFunc("OPTIONS /api/v1/calculate", handler.options)

	if staticDirectory := strings.TrimSpace(os.Getenv("STATIC_DIR")); staticDirectory != "" {
		mux.Handle("/", spaHandler(staticDirectory))
	}

	return handler.middleware(mux)
}

func (h *Handler) health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) options(response http.ResponseWriter, _ *http.Request) {
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleCalculate(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var input calculateRequest
	if err := decoder.Decode(&input); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON with operation and operands.")
		return
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "Request body must contain a single JSON object.")
		return
	}

	result, err := h.calculate(input.Operation, input.Operands)
	if err != nil {
		status, code := calculationError(err)
		writeError(response, status, code, err.Error()+".")
		return
	}

	writeJSON(response, http.StatusOK, calculateResponse{
		Operation: input.Operation,
		Operands:  input.Operands,
		Result:    result,
	})
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values")
	}
	return err
}

func calculationError(err error) (int, string) {
	switch {
	case errors.Is(err, calculator.ErrUnknownOperation):
		return http.StatusBadRequest, "unknown_operation"
	case errors.Is(err, calculator.ErrInvalidOperandCount), errors.Is(err, calculator.ErrInvalidOperand):
		return http.StatusUnprocessableEntity, "invalid_operands"
	case errors.Is(err, calculator.ErrDivisionByZero):
		return http.StatusUnprocessableEntity, "division_by_zero"
	case errors.Is(err, calculator.ErrNegativeSquareRoot):
		return http.StatusUnprocessableEntity, "invalid_square_root"
	default:
		return http.StatusUnprocessableEntity, "undefined_result"
	}
}

func (h *Handler) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("Referrer-Policy", "no-referrer")
		if h.allowedOrigin != "" {
			response.Header().Set("Access-Control-Allow-Origin", h.allowedOrigin)
			response.Header().Set("Vary", "Origin")
			response.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			response.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		}

		next.ServeHTTP(response, request)
		h.logger.Info("request completed", "method", request.Method, "path", request.URL.Path)
	})
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSONStatus(response, status, errorEnvelope{Error: apiError{Code: code, Message: message}})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	writeJSONStatus(response, status, value)
}

func writeJSONStatus(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func spaHandler(directory string) http.Handler {
	fileServer := http.FileServer(http.Dir(directory))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		cleanPath := filepath.Clean(strings.TrimPrefix(request.URL.Path, "/"))
		if cleanPath == "." {
			cleanPath = "index.html"
		}
		if _, err := os.Stat(filepath.Join(directory, cleanPath)); err != nil {
			request.URL.Path = "/"
		}
		fileServer.ServeHTTP(response, request)
	})
}
