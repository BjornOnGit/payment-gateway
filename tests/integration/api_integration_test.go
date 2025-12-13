package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/BjornOnGit/payment-gateway/internal/bus"
	"github.com/BjornOnGit/payment-gateway/internal/db/repo"
	"github.com/BjornOnGit/payment-gateway/internal/util"
	"github.com/BjornOnGit/payment-gateway/internal/service"
	"github.com/BjornOnGit/payment-gateway/internal/api"
)

func StartTestServer(t *testing.T, db *sql.DB, mem bus.Bus) (addr string, shutdown func()) {
	txRepo := repo.NewPostgresTransactionRepository(db)
	svc := service.NewTransactionService(txRepo, mem)

	mux := api.NewRouter(svc)
	// wrap with metrics middleware
	handler := util.MetricsMiddleware(mux)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	srv := &http.Server{Handler: handler}
	go srv.Serve(ln)

	return ln.Addr().String(), func() {
		_ = srv.Shutdown(context.Background())
	}
}

func TestCreateTransactionIntegration(t *testing.T) {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set; skipping integration test")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	mem := bus.NewInMemoryBus()
	addr, shutdown := StartTestServer(t, db, mem)
	defer shutdown()

	// subscribe to transaction.created
	ch := make(chan []byte, 1)
	_ = mem.Subscribe(context.Background(), "transaction.created", func(ctx context.Context, topic, key string, payload []byte) error {
		ch <- append([]byte(nil), payload...)
		return nil
	})

	// prepare request
	user := uuid.New()
	merchant := uuid.New()
	body := map[string]any{
		"amount":   1500,
		"currency": "NGN",
		"user_id": user,
		"merchant_id": merchant,
	}
	bb, _ := json.Marshal(body)
	resp, err := http.Post("http://"+addr+"/v1/transactions", "application/json", bytes.NewReader(bb))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	idStr := respBody["id"]

	// assert DB row exists
	var count int
	err = db.QueryRow(`SELECT COUNT(1) FROM transactions WHERE id = $1`, idStr).Scan(&count)
	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected transaction row, got %d", count)
	}

	// assert event was published
	select {
	case p := <-ch:
		var got map[string]any
		_ = json.Unmarshal(p, &got)
		if got["id"] != idStr {
			t.Fatalf("event id mismatch: %v != %v", got["id"], idStr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for event")
	}
}
