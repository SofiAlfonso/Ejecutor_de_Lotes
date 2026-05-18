
# ejecutor — Servicio de ejecución de procesos de lotes

## Descripción

`ejecutor` lanza procesos de forma independiente en background, administra su ciclo de vida y persiste el estado de cada ejecución en `aralmac/ejecuciones/`. Lee los binarios directamente desde `aralmac/programas/` (gestionados por `gesprog`) y redirige stdin/stdout/stderr desde/hacia ficheros en `aralmac/ficheros/` (gestionados por `gesfich`).

## Sinopsis

```
ejecutor -e <pipe-req> [-d <pipe-res>] -x <ruta_aralmac>
```

| Flag | Significado                                                     |
|------|-----------------------------------------------------------------|
| `-e` | Pipe de peticiones entrantes (lo crea `ejecutor` al arrancar)   |
| `-d` | Pipe de respuestas salientes — solo en Linux (half‑duplex)      |
| `-x` | Ruta raíz del almacenamiento (`aralmac/`)                       |

## Operaciones

| Operación   | Descripción                                                                           | Retorna                        |
|-------------|---------------------------------------------------------------------------------------|--------------------------------|
| `Ejecutar`  | Lanza el programa en background; retorna inmediatamente sin esperar que termine       | `{ "id-ejecucion": "e-XXXX" }` |
| `Estado`    | Devuelve el estado de una ejecución específica o la lista de todas las ejecuciones    | JSON de la ejecución / lista   |
| `Matar`     | Termina forzosamente un proceso en ejecución (`Process.Kill`)                         | `{ "estado": "ok" }`           |
| `Suspender` | Suspende el servicio; rechaza nuevos `Ejecutar`, los procesos activos siguen corriendo | `{ "estado": "ok" }`          |
| `Reasumir`  | Reanuda la aceptación de peticiones desde estado `Suspendido`                         | `{ "estado": "ok" }`           |
| `Parar`     | Cierre ordenado: deja de aceptar `Ejecutar`, espera que terminen los procesos activos | `{ "estado": "ok" }`           |

### Payload de `Ejecutar`

```json
{
  "servicio":    "ejecutor",
  "operacion":   "Ejecutar",
  "id-programa": "p-0001",
  "stdin":       "f-0001",
  "stdout":      "f-0002",
  "stderr":      "f-0003"
}
```

- `id-programa`: obligatorio. Debe existir como `p-XXXX.bin` y `p-XXXX.json` en `aralmac/programas/`.
- `stdin`, `stdout`, `stderr`: opcionales. Si se especifican, deben existir como `f-XXXX.dat` en `aralmac/ficheros/`.
- Si el programa o algún fichero no existe → `{ "estado": "error", "mensaje": "..." }`. No se lanza ningún proceso.

### Payload de `Estado`

```json
{ "servicio": "ejecutor", "operacion": "Estado" }
```
```json
{ "servicio": "ejecutor", "operacion": "Estado", "id-ejecucion": "e-0001" }
```

- Sin `id-ejecucion` → lista todas las ejecuciones registradas en la sesión.
- Con `id-ejecucion` → devuelve el estado individual.

### Payload de `Matar`

```json
{ "servicio": "ejecutor", "operacion": "Matar", "id-ejecucion": "e-0001" }
```

- Si el proceso ya terminó → `{ "estado": "error", "mensaje": "proceso no encontrado o ya terminado" }`.

## Estados del servicio

```
[Ejecutar] ──Suspender──> [Suspendido] ──Reasumir──> [Ejecutar]
    │
    └────Parar────> [Parando] ──(procesos_activos==0)──> [Terminado]
```

| Estado       | Acepta `Ejecutar` | Acepta `Estado`/`Matar` | Descripción                                      |
|--------------|:-----------------:|:-----------------------:|--------------------------------------------------|
| `Ejecutar`   | ✅                | ✅                      | Estado inicial. Acepta todas las operaciones.    |
| `Suspendido` | ❌                | ✅                      | Rechaza nuevas ejecuciones. Procesos siguen.     |
| `Parando`    | ❌                | ✅                      | Espera que los procesos activos terminen.        |
| `Terminado`  | ❌                | ❌                      | Servicio finalizado. No acepta ninguna petición. |

## Estados de una ejecución

| Estado       | Descripción                                                        |
|--------------|--------------------------------------------------------------------|
| `Ejecutando` | El proceso hijo sigue en ejecución.                                |
| `Terminado`  | El proceso terminó. El campo `codigo-salida` indica el resultado.  |

## Persistencia

```
aralmac/
├── programas/
│   ├── p-0001.bin       ← binario gestionado por gesprog
│   └── p-0001.json      ← { "id-programa", "nombre", "args", "env" }
├── ficheros/
│   └── f-0001.dat       ← fichero gestionado por gesfich
└── ejecuciones/
    └── e-0001.json      ← { "id-ejecucion", "id-programa", "proceso-estado", "codigo-salida", "terminado" }
```

El archivo `e-XXXX.json` se escribe dos veces: al lanzar el proceso (estado `Ejecutando`) y al terminar (estado `Terminado` con código de salida).

