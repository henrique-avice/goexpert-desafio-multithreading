package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newBrasilAPIWithBase(baseURL string) *BrasilAPI {
	c := &BrasilAPI{
		client:  &http.Client{Timeout: 2 * time.Second},
		baseURL: baseURL,
		logger:  logger,
	}
	return c
}

func TestBrasilAPI_Query_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(brasilAPIResponse{
			CEP:          "01310100",
			State:        "SP",
			City:         "São Paulo",
			Neighborhood: "Bela Vista",
			Street:       "Avenida Paulista",
		})
	}))
	defer srv.Close()

	api := newBrasilAPIWithBase(srv.URL)
	addr, err := api.Query(context.Background(), "01310100")
	if err != nil {
		t.Fatalf("esperado sucesso, got: %v", err)
	}
	if addr.Cidade != "São Paulo" {
		t.Errorf("Cidade esperada 'São Paulo', got: %s", addr.Cidade)
	}
	if addr.Estado != "SP" {
		t.Errorf("Estado esperado 'SP', got: %s", addr.Estado)
	}
	if addr.Logradouro != "Avenida Paulista" {
		t.Errorf("Logradouro esperado 'Avenida Paulista', got: %s", addr.Logradouro)
	}
	if addr.Bairro != "Bela Vista" {
		t.Errorf("Bairro esperado 'Bela Vista', got: %s", addr.Bairro)
	}
	if addr.CEP != "01310100" {
		t.Errorf("CEP esperado '01310100', got: %s", addr.CEP)
	}
}

func TestBrasilAPI_Query_InvalidCEP(t *testing.T) {
	api := newBrasilAPIWithBase("http://localhost")
	_, err := api.Query(context.Background(), "123")
	if err == nil {
		t.Fatal("esperado erro para CEP inválido")
	}
}

func TestBrasilAPI_Query_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	api := newBrasilAPIWithBase(srv.URL)
	_, err := api.Query(context.Background(), "00000000")
	if err == nil {
		t.Fatal("esperado erro para HTTP 404")
	}
}

func TestBrasilAPI_Query_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("isso nao e json"))
	}))
	defer srv.Close()

	api := newBrasilAPIWithBase(srv.URL)
	_, err := api.Query(context.Background(), "01310100")
	if err == nil {
		t.Fatal("esperado erro para JSON inválido")
	}
}

func TestBrasilAPI_Query_IncompleteResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(brasilAPIResponse{
			CEP:   "01310100",
			State: "SP",
			// City e Street ausentes
		})
	}))
	defer srv.Close()

	api := newBrasilAPIWithBase(srv.URL)
	_, err := api.Query(context.Background(), "01310100")
	if err == nil {
		t.Fatal("esperado erro para resposta incompleta")
	}
}

func TestBrasilAPI_Query_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(brasilAPIResponse{CEP: "01310100", State: "SP", City: "São Paulo", Street: "Av"})
	}))
	defer srv.Close()

	api := newBrasilAPIWithBase(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := api.Query(ctx, "01310100")
	if err == nil {
		t.Fatal("esperado erro para contexto cancelado")
	}
}
