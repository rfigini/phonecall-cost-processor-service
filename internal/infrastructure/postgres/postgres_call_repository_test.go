package postgres

import (
	"database/sql"
	"log"
	"os"
	"phonecall-cost-processor-service/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	dsn := "postgres://testuser:testpass@localhost:5444/testdb?sslmode=disable"

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil && db.Ping() == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a Postgres: %v", err)
	}

	createTable()
	code := m.Run()
	_ = db.Close()
	os.Exit(code)
}

func createTable() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS calls (
		call_id UUID PRIMARY KEY,
		caller TEXT,
		receiver TEXT,
		duration_in_seconds INT,
		start_timestamp TIMESTAMP,
		cost FLOAT,
		currency TEXT,
		status TEXT,
		refunded BOOLEAN,
		refund_reason TEXT,
		processed_at TIMESTAMP
	);`)
	if err != nil {
		log.Fatalf("❌ Error creando tabla: %v", err)
	}
}

func clearCallsTable() {
	_, _ = db.Exec("DELETE FROM calls;")
}

func setupTest(t *testing.T) *PostgresCallRepository {
	clearCallsTable()
	return NewPostgresCallRepository(db)
}

func TestSaveIncomingCall(t *testing.T) {
	repo := setupTest(t)
	call := model.NewIncomingCall{
		CallID:         uuid.New().String(),
		Caller:         "juan",
		Receiver:       "pedro",
		DurationInSec:  60,
		StartTimestamp: time.Now().Format(time.RFC3339),
	}
	if err := repo.SaveIncomingCall(call); err != nil {
		t.Fatalf("error guardando call: %v", err)
	}
}

func TestSaveAndGetStatus(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()
	call := model.NewIncomingCall{CallID: callID, Caller: "Juan", Receiver: "Maria", DurationInSec: 120, StartTimestamp: time.Now().Format(time.RFC3339)}
	_ = repo.SaveIncomingCall(call)
	status, err := repo.GetCallStatus(callID)
	if err != nil || status != "PENDING" {
		t.Fatalf("expected status PENDING, got %s (err: %v)", status, err)
	}
}

func TestUpdateCallCost(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()
	call := model.NewIncomingCall{CallID: callID, Caller: "Ana", Receiver: "Luis", DurationInSec: 80, StartTimestamp: time.Now().Format(time.RFC3339)}
	_ = repo.SaveIncomingCall(call)
	_ = repo.UpdateCallCost(callID, 19.99, "USD")
	status, err := repo.GetCallStatus(callID)
	if err != nil || status != "OK" {
		t.Fatalf("expected status OK, got %s (err: %v)", status, err)
	}
}

func TestApplyRefund_NewCall(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()
	refund := model.RefundCall{CallID: callID, Reason: "Cobro duplicado"}
	_ = repo.ApplyRefund(refund)
	status, err := repo.GetCallStatus(callID)
	if err != nil || status != "REFUND_PARTIALLY" {
		t.Fatalf("expected status REFUND_PARTIALLY, got %s (err: %v)", status, err)
	}
}

func TestMarkCostAsFailed(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()
	call := model.NewIncomingCall{CallID: callID, Caller: "Pepe", Receiver: "Lalo", DurationInSec: 45, StartTimestamp: time.Now().Format(time.RFC3339)}
	_ = repo.SaveIncomingCall(call)
	_ = repo.MarkCostAsFailed(callID)
	status, err := repo.GetCallStatus(callID)
	if err != nil || status != "ERROR" {
		t.Fatalf("expected status ERROR, got %s (err: %v)", status, err)
	}
}

func TestMarkCallAsInvalid(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()
	call := model.NewIncomingCall{CallID: callID, Caller: "Tom", Receiver: "Jerry", DurationInSec: 20, StartTimestamp: time.Now().Format(time.RFC3339)}
	_ = repo.SaveIncomingCall(call)
	_ = repo.MarkCallAsInvalid(callID)
	status, err := repo.GetCallStatus(callID)
	if err != nil || status != "INVALID" {
		t.Fatalf("expected status INVALID, got %s (err: %v)", status, err)
	}
}

func TestFillMissingCallData(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()
	refund := model.RefundCall{CallID: callID, Reason: "error"}
	_ = repo.ApplyRefund(refund)
	fill := model.NewIncomingCall{CallID: callID, Caller: "Carlos", Receiver: "Daniela", DurationInSec: 100, StartTimestamp: time.Now().Format(time.RFC3339)}
	if err := repo.FillMissingCallData(fill); err != nil {
		t.Fatalf("expected no error filling data, got %v", err)
	}
}

func TestGetCallStatus_Empty(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()

	status, err := repo.GetCallStatus(callID)
	if err != nil {
		t.Fatalf("GetCallStatus failed on empty: %v", err)
	}
	if status != "" {
		t.Errorf("Expected empty status, got %s", status)
	}
}

func TestRefundPartiallyThenFill_ShouldBecomeRefunded(t *testing.T) {
	repo := setupTest(t)
	callID := uuid.New().String()

	refund := model.RefundCall{CallID: callID, Reason: "Cobro anticipado"}
	if err := repo.ApplyRefund(refund); err != nil {
		t.Fatalf("error aplicando refund: %v", err)
	}
	status, _ := repo.GetCallStatus(callID)
	if status != "REFUND_PARTIALLY" {
		t.Fatalf("expected REFUND_PARTIALLY, got %s", status)
	}

	call := model.NewIncomingCall{
		CallID:         callID,
		Caller:         "Leo",
		Receiver:       "Max",
		DurationInSec:  50,
		StartTimestamp: time.Now().Format(time.RFC3339),
	}
	if err := repo.FillMissingCallData(call); err != nil {
		t.Fatalf("error llenando datos faltantes: %v", err)
	}
	status, _ = repo.GetCallStatus(callID)
	if status != "REFUNDED" {
		t.Fatalf("expected status REFUNDED after filling, got %s", status)
	}
}
