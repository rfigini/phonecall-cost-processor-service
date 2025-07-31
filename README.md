
# 📞 Phonecall Cost Processor Service

**Autor**: Ricardo Figini  
**Ejercicio Técnico – Backend Brubank**

---

## 🧠 ¿Qué hace este servicio?

Este servicio consume mensajes desde una cola con eventos de llamadas telefónicas, los procesa, consulta una API externa para calcular el costo y persiste los resultados en una base de datos. Está preparado para:

- Manejar mensajes duplicados y sin orden.
- Soportar fallas intermitentes o caídas prolongadas de la API externa.
- Facilitar reintentos y diagnósticos.
- Extender el consumo de nuevos mensajes facilmente.
- A futuro reprocesar llamadas que hayan quedao sin costo (no implementado)
- A futuro generar reportes mensuales de facturación. (no implementado)

---

## 🛠️ Decisiones Técnicas

### ✔️ Tolerancia a duplicados y desorden
- Se garantiza **idempotencia** mediante el uso de `call_id` como clave primaria.
- La lógica actual **ignora llamadas ya procesadas** (con estado `OK`, `ERROR`, `REFUNDED`, `INVALID`), evitando reprocesamientos innecesarios.

### ✔️ Resiliencia ante fallos en la API
- Se utiliza un cliente HTTP con **reintentos automáticos y backoff exponencial** ante errores 5xx o timeouts.
- Si la API falla luego de reintentos, se marca la llamada como `ERROR`, permitiendo **reprocesos posteriores**.

### ✔️ Diagnóstico y trazabilidad
- Se registra el estado final de cada llamada (`OK`, `ERROR`, `REFUNDED`, `INVALID`) junto con timestamps y razón de fallo si aplica.
- Esto permite **detectar errores de negocio** (por ejemplo, llamadas no encontradas) y diferenciarlos de errores técnicos.

### ✔️ Extensibilidad
- Agregar un nuevo tipo de mensaje (ej: `call_quality_issue`) requiere:
  1. Agregar una entrada al dispatcher de mensajes.
  2. Crear un nuevo `UseCase` con su handler.
  3. Definir el modelo y testear el flujo.

Esto respeta el principio **Open/Closed** y no requiere tocar los casos ya existentes.

---

## ▶️ Cómo ejecutarlo

### Requisitos

- Go 1.20+
- Docker + Docker Compose

---

### 1. Levantar dependencias

```bash
docker-compose up -d
```

Esto inicia:
- PostgreSQL (localhost:5433)
- RabbitMQ (localhost:5672 + UI en http://localhost:15672)
- Mock API de costos en localhost:8081

---

### 2. Ejecutar el servicio

```bash
go run cmd/main.go
```

El servicio:
- Escucha mensajes desde `calls_queue`.
- Procesa mensajes tipo `new_incoming_call` y `refund_call`.
- Guarda resultados en la base.

---

### 3. Enviar mensaje de prueba

```bash
curl -X POST localhost:8080/incoming-call \
  -H "Content-Type: application/json" \
  -d '{
    "call_id": "123e4567-e89b-12d3-a456-426614174000",
    "caller": "+1234567890",
    "receiver": "+0987654321",
    "duration_in_seconds": 120,
    "start_timestamp": "2024-08-29T09:24:28Z"
  }'
```

---

## 🧪 Tests

Para correr todos los tests:

```bash
go test ./...
```

Tests de integración con PostgreSQL real:

```bash
go test ./internal/infrastructure/postgres
```

---

## 🗃️ Estado de llamadas en la base de datos

Cada llamada persiste su estado:

- `OK`: procesada exitosamente.
- `ERROR`: falló la consulta de costos (reintentos agotados o error técnico).
- `REFUNDED`: fue reembolsada por reclamo.
- `INVALID`: error de negocio (ej: llamada no encontrada en la API).

Esto permite en el futuro:
- Implementar un **reprocesador automático de llamadas con estado `ERROR`**.
- Excluir las `INVALID` que fallaron por causas no recuperables.

La fecha (`start_timestamp`) permite generar **reportes mensuales de facturación**.

---

## 🌐 Variables de entorno utilizadas

```env
RABBITMQ_URL=amqp://guest:guest@localhost:5672
RABBITMQ_QUEUE=calls_queue
DB_URL=postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable
COST_API_URL=http://localhost:8081
```

---

## 🧩 Estructura del código

```
cmd/                    # Entry point
internal/
  application/          # Casos de uso (lógica de negocio)
  domain/               # Modelos del negocio
  infrastructure/
    handler/            # HTTP + RabbitMQ handlers
    client/             # API externa de costos
    postgres/           # Repositorio de llamadas
    rabbitmq/           # Consumo de mensajes
mock/                   # Mock de API de costos
```

---

## 📝 Consideraciones finales

- Se priorizó un diseño simple, legible y con foco en resiliencia sin sobreingeniería.
- Está diseñado para agregar nuevas funcionalidades sin modificar la lógica existente.
