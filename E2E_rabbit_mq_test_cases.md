### 1 - Llamada exitosa - Esperado: `OK`

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

### 2 - Refund de llamada existente - Esperado: `REFUNDED`

```json
{
  "type": "refund_call",
  "body": {
    "call_id": "11111111-1111-1111-1111-111111111111",
    "reason": "Cliente reclamo"
  }
}
```

### 3.0 - Refund parcial (call no existe aún) - Esperado: `REFUND_PARTIALLY`

```json
{
  "type": "refund_call",
  "body": {
    "call_id": "22222222-2222-2222-2222-222222222222",
    "reason": "Cobro indebido"
  }
}
```

### 3.1 - Completar datos refund parcial - Esperado: `REFUNDED` - Se completan los datos faltantes 

```json
{
  "type": "new_incoming_call",
  "body": {
    "call_id": "22222222-2222-2222-2222-222222222222",
    "caller": "+1111111111",
    "receiver": "+2222222222",
    "duration_in_seconds": 60,
    "start_timestamp": "2024-08-29T12:00:00Z"
  }
}
```

### 4 - Mensaje duplicado (se descarta) - Esperado: No cambia estado

```json
{
  "type": "new_incoming_call",
  "body": {
    "call_id": "11111111-1111-1111-1111-111111111111",
    "caller": "+9999999999",
    "receiver": "+8888888888",
    "duration_in_seconds": 200,
    "start_timestamp": "2024-08-29T12:00:00Z"
  }
}
```

### 5 - Error 5xx que no se recupera - Esperado: `ERROR`

```json
{
  "type": "new_incoming_call",
  "body": {
    "call_id": "123e4567-e89b-12d3-a456-426614174997",
    "caller": "+7777777777",
    "receiver": "+6666666666",
    "duration_in_seconds": 140,
    "start_timestamp": "2024-08-29T12:00:00Z"
  }
}
```

### 6 - Error 5xx que se recupera - Esperado: `OK`

```json
{
  "type": "new_incoming_call",
  "body": {
    "call_id": "123e4567-e89b-12d3-a456-426614174999",
    "caller": "+1111222233",
    "receiver": "+3333444455",
    "duration_in_seconds": 180,
    "start_timestamp": "2024-08-29T12:00:00Z"
  }
}
```

### 7 - Error de negocio 4xx - Esperado: `INVALID`

```json
{
  "type": "new_incoming_call",
  "body": {
    "call_id": "123e4567-e89b-12d3-a456-426614174998",
    "caller": "+5555555555",
    "receiver": "+6666666666",
    "duration_in_seconds": 90,
    "start_timestamp": "2024-08-29T12:00:00Z"
  }
}
```