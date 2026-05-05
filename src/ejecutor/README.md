# ejecutor — Lanzador de procesos de lotes

## Descripción

`ejecutor` lanza procesos del sistema operativo a partir de programas almacenados en `aralmac/programas/` y registra su ejecución en `aralmac/lotes/`. Accede directamente al almacenamiento sin pasar por `gesfich` ni `gesprog`.

## Sinopsis

```
ejecutor -e <pipe-req> -d <pipe-res> -x <ruta_aralmac>
```

| Flag | Significado                                    |
|------|------------------------------------------------|
| `-e` | Pipe de peticiones entrantes (desde ctrllt)    |
| `-d` | Pipe de respuestas salientes (hacia ctrllt)    |
| `-x` | Ruta raíz del almacenamiento (`aralmac/`)      |

## Operaciones

| Acción      | Descripción                                                             | Retorna         |
|-------------|-------------------------------------------------------------------------|-----------------|
| `ejecutar`  | Lanza el proceso en background; retorna inmediatamente (no bloqueante)  | `l-XXXX`        |
| `estado`    | Devuelve el estado de un lote o lista de todos los lotes                | JSON del lote / lista |
| `matar`     | Envía señal SIGKILL al proceso (`TerminateProcess` en Windows)          | `ok`            |
| `suspender` | Suspende el servicio (deja de aceptar nuevos lotes)                     | `ok`            |
| `reasumir`  | Reanuda la aceptación de peticiones                                     | `ok`            |
| `parar`     | Cierre elegante: espera a que terminen los lotes activos, luego sale    | `ok`            |
| `terminar`  | Cierre inmediato: mata todos los procesos hijos y sale                  | —               |

## Estados del servicio

```
Corriendo ⇄ Suspendido
Corriendo → Parando → Terminado
Corriendo → Terminado
```

- **Suspendido:** rechaza `ejecutar`; sigue respondiendo a `estado` y `matar`.
- **Parando:** no acepta nuevos lotes; espera que los lotes activos terminen naturalmente.

## Estados de un lote

```
corriendo → terminado
corriendo → fallido
corriendo → matado
```

| Estado      | Descripción                                              |
|-------------|----------------------------------------------------------|
| `corriendo` | Proceso OS en ejecución                                  |
| `terminado` | Proceso terminó con código de salida 0                   |
| `fallido`   | Proceso terminó con código de salida distinto de 0       |
| `matado`    | Proceso terminado por la operación `matar`               |

> **Nota:** `ejecutor` lee los binarios directamente desde `aralmac/programas/` sin pasar por `gesprog`, y escribe registros de lotes en `aralmac/lotes/` sin pasar por `gesfich`.
