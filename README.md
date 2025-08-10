# 📞 Phonecall Cost Processor Service

**Author**: Ricardo Figini

---

## 🧠 What does this service do?

This service consumes messages from a queue with phone call events, processes them, queries an external API to calculate the cost, and stores the results in a database.  
It is designed to:

- Handle duplicate and out-of-order messages.  
- Tolerate intermittent failures or prolonged outages of the external API.  
- Support retries and easy diagnostics.  
- Easily extend to consume new types of messages.  
- *(Future)* Reprocess calls that were left without cost (not implemented).  
- *(Future)* Generate monthly billing reports (not implemented).  

---

## 🛠️ Technical Decisions

### ✔️ Duplicate and out-of-order tolerance
- **Idempotency** is guaranteed by using `call_id` as the primary key.  
- Already processed calls (`OK`, `ERROR`, `REFUNDED`, `REFUND_PARTIALLY`, `INVALID`) are ignored to avoid unnecessary reprocessing.  

### ✔️ API failure resilience
- The HTTP client uses **automatic retries with exponential backoff** for 5xx errors or timeouts.  
- If the API still fails after retries, the call is marked as `ERROR` so it can be reprocessed later.  

### ✔️ Diagnostics and traceability
- The final state of each call is recorded (`OK`, `ERROR`, `REFUNDED`, `INVALID`, `REFUND_PARTIALLY`), along with timestamps and failure reason (if applicable).  
- This allows identifying **business errors** (e.g., call not found) separately from technical errors.  

### ✔️ Extensibility
- Adding a new message type (e.g., `call_quality_issue`) only requires:
  1. Adding an entry to the message dispatcher.  
  2. Creating a new `UseCase` with its handler.  
  3. Defining the model and testing the flow.  

This follows the **Open/Closed principle** without modifying existing cases.

---

## ▶️ How to run it

### Requirements
- Go 1.20+  
- Docker + Docker Compose  

### 1. Start dependencies
```bash
docker-compose up -d
```
This starts:
- PostgreSQL (localhost:5433)  
- RabbitMQ (localhost:5672 + UI at [http://localhost:15672](http://localhost:15672))  
- Mock cost API on localhost:8081  

### 2. Run the service
```bash
go run cmd/main.go
```
The service:
- Listens to messages from `calls_queue`.  
- Processes `new_incoming_call` and `refund_call` message types.  
- Stores results in the database.  

---

## 🔮 End-to-end test with RabbitMQ

To test the system end-to-end:  
1. Start the environment as described in the README.  
2. Open the RabbitMQ UI: [http://localhost:15672](http://localhost:15672)  
   - User: `guest`, Password: `guest`  
3. Go to the `calls_queue` queue and use **Publish message**:  
   - Routing key: `calls_queue`  
   - Payload example:  
```json
{
  "type": "new_incoming_call",
  "body": {
    "call_id": "11111111-1111-1111-1111-111111111111",
    "caller": "+1234567890",
    "receiver": "+0987654321",
    "duration_in_seconds": 120,
    "start_timestamp": "2024-08-29T12:00:00Z"
  }
}
```

> For all available test cases, see the `E2E_rabbit_mq_test_casess.md` file included in the project.

---

## 💪 Tests

Integration tests with a real PostgreSQL instance:
```bash
docker-compose -f docker-compose-postgres-test.yml up -d
go test ./internal/infrastructure/postgres
```

Run all tests:
```bash
go test ./...
```

---

## 🗃️ Call status in the database

Each call stores its status:

- `OK`: processed successfully.  
- `ERROR`: cost retrieval failed (retries exhausted or technical error).  
- `REFUNDED`: refunded due to a claim.  
- `REFUND_PARTIALLY`: refund received before the call was processed.  
- `INVALID`: business error (e.g., call not found in the API).  

This enables, in the future:
- Implementing an **automatic reprocessor** for calls in `ERROR`.  
- Excluding `INVALID` calls that failed for unrecoverable reasons.  

The `start_timestamp` also allows generating **monthly billing reports**.

---

## 🌐 Environment variables

```env
RABBITMQ_URL=amqp://guest:guest@localhost:5672
RABBITMQ_QUEUE=calls_queue
DB_URL=postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable
COST_API_URL=http://localhost:8081
```

---

## 📁 Code structure

```
cmd/                    # Entry point
internal/
  application/          # Use cases (business logic)
  domain/               # Business models
  infrastructure/
    handler/            # RabbitMQ handlers (application entry point)
    client/             # External cost API
    postgres/           # Call repository
    rabbitmq/           # Message consumption
mock/                   # Mock cost API
```

The architecture follows the **Hexagonal Architecture** pattern to decouple domain from infrastructure.  
Handlers act as the application’s entry point and map 1:1 to their respective use cases.  

> ⚠️ The repository currently holds multiple responsibilities. While this is recognized as an SRP violation, it was kept for pragmatism in the context of a technical exercise. This is marked for future refactoring.

---

## 📝 Final considerations
- The design is simple, readable, and resilience-focused without overengineering.  
- It’s built to easily add new features without modifying existing logic.  
