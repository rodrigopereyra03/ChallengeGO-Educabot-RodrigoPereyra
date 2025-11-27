# Bookshop Metrics API – Documentación tipo Swagger

## Información general

- **Nombre**: Bookshop Metrics API  
- **Base URL**: `http://localhost:8080` (ajustar al puerto que uses)  
- **Versión**: `1.0.0`  
- **Formato**: JSON  
- **Autenticación**: No requiere (demo)

---

## Endpoints

### GET /metrics

Obtiene métricas agregadas de ventas de libros a partir de un servicio externo.

**Descripción**  
Devuelve:
- `mean_units_sold`: promedio de unidades vendidas
- `cheapest_book`: nombre del libro más barato
- `books_written_by_author`: cantidad de libros escritos por un autor

**Query parameters**

| Nombre | Tipo   | Obligatorio | Descripción                                                 |
|--------|--------|------------|-------------------------------------------------------------|
| author | string | No         | Nombre del autor a normalizar y buscar (case insensitive). |

**Ejemplo de request**

```http
GET /metrics?author=Alan%20Donovan HTTP/1.1
Host: localhost:8080
Accept: application/json
```

**Responses**

#### 200 OK

```json
{
  "mean_units_sold": 11000,
  "cheapest_book": "The Go Programming Language",
  "books_written_by_author": 1
}
```

Campos:

- `mean_units_sold` (number, uint): promedio de `units_sold` de todos los libros retornados por el servicio externo.  
- `cheapest_book` (string): nombre del libro con menor `price`.  
- `books_written_by_author` (number, uint): cantidad de libros cuyo `author` coincide (normalizado a lowercase/trim) con el parámetro `author`.

#### 502 Bad Gateway

Error en el servicio externo (fallas repetidas, circuit breaker abierto, etc.).

```json
{
  "message": "external service error"
}
```

Se produce cuando:

- El provider devuelve `ErrExternalService` (errores de red, 5xx del externo, circuit breaker, etc.).

#### 504 Gateway Timeout

Timeout de comunicación con la API externa.

```json
{
  "message": "external service timeout"
}
```

Se produce cuando:

- El `context` de la petición expira durante el llamado al servicio externo.  
- El provider devuelve `ErrTimeout`.

#### 422 Unprocessable Entity

Respuesta inválida desde la API externa (payload inesperado o no parseable).

```json
{
  "message": "invalid response from external API"
}
```

Se produce cuando:

- El provider devuelve `ErrInvalidResponse` (por ejemplo, JSON inválido o status codes de negocio que tratan la respuesta como inválida).

#### 500 Internal Server Error

Error interno no controlado.

```json
{
  "message": "internal server error"
}
```

Cualquier error que no matchee con `ErrTimeout`, `ErrExternalService` ni `ErrInvalidResponse`.

---

## Modelos de datos

### Libro (Book)

Ejemplo de objeto `Book` retornado por el servicio externo (no se expone directamente por tu API, se usa para las métricas):

```json
{
  "id": 1,
  "name": "The Go Programming Language",
  "author": "Alan Donovan",
  "units_sold": 5000,
  "price": 40
}
```

Campos:

- `id` (number, uint): identificador del libro.  
- `name` (string): nombre del libro.  
- `author` (string): autor.  
- `units_sold` (number, uint): unidades vendidas.  
- `price` (number, uint): precio.

> Nota: este modelo se usa internamente y es la respuesta "cruda" del servicio externo. El endpoint público `/metrics` expone solo métricas agregadas.

---

## Comportamiento de resiliencia

Aunque no se expone directamente vía HTTP, es importante para consumidores de la API entender las posibles fallas.

### Retry + Backoff

El provider `HttpBooksProvider`:

- Reintenta hasta **3 intentos** ante errores transitorios (errores de red, status 5xx, `429 Too Many Requests`).  
- Usa **backoff exponencial** (aprox. 150ms, 300ms, 600ms).  
- Respeta el `context.Context` de la petición: si se cancela o expira, no sigue reintentando y devuelve `ErrTimeout`.

### Circuit Breaker

Implementado en `internal/platform/resilience.CircuitBreaker` y usado dentro del provider externo.

- Cuando se alcanzan N fallos consecutivos (configurado en el constructor), el circuito:
  - Se abre durante un tiempo (`openDuration`).  
  - Mientras está abierto, las peticiones fallan rápido sin intentar llamar al servicio externo (fail-fast) retornando `ErrCircuitOpen`.
- Los errores se traducen a:
  - `ErrExternalService` (que el handler mapea a `502 Bad Gateway`).  
  - En caso de timeout, `ErrTimeout` (`504 Gateway Timeout`).  
  - Respuestas inválidas → `ErrInvalidResponse` (`422 Unprocessable Entity`).

---

## Ejemplos de uso

### Llamado exitoso

```bash
curl "http://localhost:8080/metrics?author=Alan%20Donovan"
```

Respuesta:

```json
{
  "mean_units_sold": 11000,
  "cheapest_book": "The Go Programming Language",
  "books_written_by_author": 1
}
```

### Servicio externo caído / circuito abierto

```bash
curl "http://localhost:8080/metrics?author=Tolkien"
```

Respuesta típica:

```json
{
  "message": "external service error"
}
```

---
