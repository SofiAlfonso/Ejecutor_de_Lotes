# ejecutor — Servicio de ejecución de procesos de lotes

## Descripción

`ejecutor` lanza procesos de forma independiente en background, administra su ciclo de vida y persiste el estado de cada ejecución en `aralmac/ejecuciones/`. Lee los binarios directamente desde `aralmac/programas/` (gestionados por `gesprog`) y redirige stdin/stdout/stderr desde/hacia ficheros en `aralmac/ficheros/` (gestionados por `gesfich`).

La implementación utiliza el paquete `common` para:
- Generación de identificadores de ejecución (`e-XXXX`) mediante `common.GenerarIDEjecucion`.
- Comunicación por pipes nombrados (`common.AbrirPipes`): half‑duplex en Linux (dos FIFOs) y full‑duplex en Windows (un solo pipe).

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

## Generación de IDs

Los identificadores `e-XXXX` se generan mediante `common.GenerarIDEjecucion()`, que escanea el directorio `aralmac/ejecuciones/` y asigna el siguiente número disponible. La integración con `common` ya está completa (el servicio llama a `common.InitIDs` al arrancar).

## Dependencias

- Solo utiliza la biblioteca estándar de Go, excepto `golang.org/x/sys/windows` (para Windows). Esta dependencia se descarga automáticamente con `go mod tidy`.
- El paquete `common` (ya implementado) proporciona `AbrirPipes` y `GenerarIDEjecucion`. El servicio llama a `common.InitIDs` al arrancar.

## Compilación y ejecución

### En Linux (WSL)

```bash
# Compilar
cd src/ejecutor
go build -o ejecutor .

# Crear directorio de almacenamiento y FIFOs
mkdir -p aralmac/ejecuciones aralmac/programas aralmac/ficheros
mkfifo /tmp/ejecutor_in /tmp/ejecutor_out

# Ejecutar
./ejecutor -e /tmp/ejecutor_in -d /tmp/ejecutor_out -x ./aralmac
```

### En Windows (PowerShell como administrador)

```powershell
# Compilar
cd src\ejecutor
go build -o ejecutor.exe

# Crear directorios
mkdir aralmac\ejecuciones, aralmac\programas, aralmac\ficheros -Force

# Ejecutar (full‑duplex, flag -d opcional)
.\ejecutor.exe -e \\.\pipe\ejecutor_pipe -x .\aralmac
```

## Prueba rápida

Una vez el servidor está corriendo, puedes enviar una petición desde otra terminal:

**Linux:**
```bash
echo '{"servicio":"ejecutor","operacion":"Ejecutar","id-programa":"p-0001"}' > /tmp/ejecutor_in
cat /tmp/ejecutor_out
```

**Windows (PowerShell):**
```powershell
$pipe = new-object System.IO.Pipes.NamedPipeClientStream("\\.\pipe\ejecutor_pipe")
$pipe.Connect()
$writer = new-object System.IO.StreamWriter($pipe)
$writer.WriteLine('{"servicio":"ejecutor","operacion":"Ejecutar","id-programa":"p-0001"}')
$writer.Flush()
$reader = new-object System.IO.StreamReader($pipe)
$reader.ReadLine()
```

La respuesta será algo como: `{"estado":"ok","id-ejecucion":"e-0001"}`.

## Código fuente

Los archivos que componen el servicio son:

- `main.go`         – entrada, parsing de flags, inicialización de `common`.
- `estado.go`       – máquinas de estados del servicio y de los procesos.
- `almacenamiento.go` – verificación de programas/ficheros, persistencia de ejecuciones.
- `operaciones.go`  – despachador de comandos JSON.
- `proceso.go`      – lanzamiento de pipeline (tuberías anónimas) y `MatarProceso`.
- `servidor.go`     – bucle de escucha de pipes mediante `common.AbrirPipes`.

