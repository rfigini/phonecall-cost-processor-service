# 📞 Phonecall Cost Processor Service

**Autor**: Ricardo Figini\
**Ejercicio Técnico – Backend Brubank**

---

## 🧠 ¿Qué hace este servicio?

Este servicio consume mensajes desde una cola con eventos de llamadas telefónicas, los procesa, consulta una API externa para calcular el costo y persiste los resultados en una base de datos. Está preparado para:

- Manejar mensajes duplicados y sin orden.
- Soportar fallas intermitentes o caídas prolongadas de la API externa.
- Facilitar reintentos y diagnósticos.
- Extender el consumo de nuevos mensajes fácilmente.
- A futuro reprocesar llamadas que hayan quedado sin costo (no implementado).
- A futuro generar reportes mensuales de facturación (no implementado).

---

## 🛠️ Decisiones Técnicas

### ✔️ Tolerancia a duplicados y desorden

- Se garantiza **idempotencia** mediante el uso de `call_id` como clave primaria.
- La lógica actual **ignora llamadas ya procesadas** (con estado `OK`, `ERROR`, `REFUNDED`, `REFUND_PARTIALLY`, `INVALID`), evitando reprocesamientos innecesarios.

### ✔️ Resiliencia ante fallos en la API

- Se utiliza un cliente HTTP con **reintentos automáticos y backoff exponencial** ante errores 5xx o timeouts.
- Si la API falla luego de reintentos, se marca la llamada como `ERROR`, permitiendo **reprocesos posteriores**.

### ✔️ Diagnóstico y trazabilidad

- Se registra el estado final de cada llamada (`OK`, `ERROR`, `REFUNDED`, `INVALID`, `REFUND_PARTIALLY`) junto con timestamps y razón de fallo si aplica.
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

### 1. Levantar dependencias

```bash
docker-compose up -d
```

Esto inicia:

- PostgreSQL (localhost:5433)
- RabbitMQ (localhost:5672 + UI en [http://localhost:15672](http://localhost:15672))
- Mock API de costos en localhost:8081

### 2. Ejecutar el servicio

```bash
go run cmd/main.go
```

El servicio:

- Escucha mensajes desde `calls_queue`.
- Procesa mensajes tipo `new_incoming_call` y `refund_call`.
- Guarda resultados en la base de datos.

---

## 🔮 Test E2E con RabbitMQ

Para probar el sistema de forma completa:

1. Levantar el entorno como indica el README.
2. Ingresar a la UI de RabbitMQ: [http://localhost:15672](http://localhost:15672)
   - Usuario: `guest`, Contraseña: `guest`
3. Ir a la cola `calls_queue` y usar la sección **Publish message**.
   - En routing key: `calls_queue`
   - En payload, usar el JSON estructurado del tipo:

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

> Para ver todos los casos de prueba disponibles, consultar el archivo `E2E_rabbit_mq_test_casess.md` incluido en el proyecto.

---

## 💪 Tests

Tests de integración con PostgreSQL real:

```bash
docker-compose -f docker-compose-postgres-test.yml up -d
go test ./internal/infrastructure/postgres
```

Para correr todos los tests:

```bash
go test ./...
```
---

## 🗃️ Estado de llamadas en la base de datos

Cada llamada persiste su estado:

- `OK`: procesada exitosamente.
- `ERROR`: falló la consulta de costos (reintentos agotados o error técnico).
- `REFUNDED`: fue reembolsada por reclamo.
- `REFUND_PARTIALLY`: se recibió un reembolso antes que la llamada.
- `INVALID`: error de negocio (ej: llamada no encontrada en la API).

Esto permite en el futuro:

- Implementar un **reprocesador automático de llamadas con estado **``.
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

## 📁 Estructura del código

```
cmd/                    # Entry point
internal/
  application/          # Casos de uso (lógica de negocio)
  domain/               # Modelos del negocio
  infrastructure/
    handler/            # RabbitMQ handlers (punto de entrada de la app)
    client/             # API externa de costos
    postgres/           # Repositorio de llamadas
    rabbitmq/           # Consumo de mensajes
mock/                   # Mock de API de costos
```

La arquitectura elegida es **hexagonal** para desacoplar el dominio de la infraestructura. Los handlers funcionan como punto de entrada a la aplicación y se relacionan 1 a 1 con sus respectivos casos de uso.

> ⚠️ El repositorio actual concentra múltiples responsabilidades. Si bien se reconoce este **code smell (SRP)**, se decidió mantenerlo por pragmatismo siendo un ejercicio técnico. Es un área marcada para refactor futuro.

---

## 📝 Consideraciones finales

- Se priorizó un diseño simple, legible y con foco en resiliencia sin sobreingeniería.
- Está diseñado para agregar nuevas funcionalidades sin modificar la lógica existente.

