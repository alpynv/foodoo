package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alpynv/foodoo/services/order/internal/orders"
	"github.com/google/uuid"
)

type Handler struct {
	store orders.Store
}

// NewHandler returns the HTTP routes exposed by the order service.
func NewHandler(store orders.Store) http.Handler {
	h := Handler{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /readyz", h.readyz)
	mux.HandleFunc("POST /orders", h.createOrder)
	mux.HandleFunc("GET /orders/{id}", h.getOrder)
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 2*time.Second)
	defer cancel()
	if h.store == nil || h.store.Ping(ctx) != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type createOrderRequest struct {
	CustomerID       string `json:"customer_id"`
	TotalAmountCents int64  `json:"total_amount_cents"`
	Currency         string `json:"currency"`
}

func (h Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if len(key) == 0 || len(key) > 255 {
		writeError(w, http.StatusBadRequest, "Idempotency-Key must be between 1 and 255 characters")
		return
	}

	var request createOrderRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}
	customerID, err := uuid.Parse(request.CustomerID)
	if err != nil || request.TotalAmountCents < 0 || len(request.Currency) != 3 {
		writeError(w, http.StatusBadRequest, "customer_id must be a UUID, total_amount_cents must be non-negative, and currency must be a 3-letter code")
		return
	}
	currency := strings.ToUpper(request.Currency)
	if currency[0] < 'A' || currency[0] > 'Z' || currency[1] < 'A' || currency[1] > 'Z' || currency[2] < 'A' || currency[2] > 'Z' {
		writeError(w, http.StatusBadRequest, "currency must be a 3-letter code")
		return
	}

	correlationID := r.Header.Get("X-Request-ID")
	if correlationID == "" {
		correlationID = uuid.NewString()
	}
	order, err := h.store.Create(r.Context(), orders.CreateInput{
		CustomerID: customerID, TotalAmountCents: request.TotalAmountCents, Currency: currency,
		IdempotencyKey: key, CorrelationID: correlationID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create order")
		return
	}
	w.Header().Set("Location", "/orders/"+order.ID.String())
	writeJSON(w, http.StatusCreated, order)
}

func (h Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "order ID must be a UUID")
		return
	}
	order, err := h.store.Get(r.Context(), id)
	if errors.Is(err, orders.ErrNotFound) {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not get order")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func contextWithTimeout(r *http.Request, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), timeout)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
