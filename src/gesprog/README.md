
# gesprog — Gestor de programas

## Descripción

`gesprog` es el servicio encargado de registrar, consultar, actualizar, eliminar y ejecutar programas dentro del sistema de lotes.  
Guarda copias de los ejecutables en el almacenamiento persistente (`aralmac/programas/`) junto con sus metadatos (argumentos y variables de entorno).  
Asigna identificadores únicos con formato `p-XXXX`.

El servicio se comunica exclusivamente a través de **tuberías nombradas** (named pipes / FIFOs).  
- En **Linux** se usan dos FIFOs (half‑duplex): uno para peticiones (`-p`) y otro para respuestas (`-c`).  
- En **Windows** se usa un único named pipe full‑duplex (el flag `-c` es opcional).

## Sinopsis

```bash
gesprog -p <pipe-peticiones> [-c <pipe-respuestas>] -x <ruta-aralmac>
```

| Flag | Descripción                                                                 |
|------|-----------------------------------------------------------------------------|
| `-p` | Nombre del pipe por donde se reciben las peticiones JSON. **Obligatorio**.  |
| `-c` | Nombre del pipe por donde se envían las respuestas (solo Linux, half‑duplex). En Windows puede omitirse. |
| `-x` | Ruta raíz del almacenamiento (por ejemplo, `./aralmac`). **Obligatorio**.   |

## Operaciones (protocolo JSON)

Todas las peticiones y respuestas son **líneas JSON terminadas en `\n`** y no superan los 4096 bytes.

### Guardar

Registra un nuevo programa. Copia el ejecutable al almacén y guarda sus metadatos.

**Petición:**
```json
{"servicio":"gesprog","operacion":"Guardar","ejecutable":"/ruta/al/ejecutable","args":["arg1","arg2"],"env":["CLAVE=VALOR"]}
```
`args` y `env` son opcionales.

**Respuesta (éxito):**
```json
{"estado":"ok","id-programa":"p-0001"}
```

### Leer

Recupera los metadatos de un programa específico o lista todos los IDs registrados.

**Por ID:**
```json
{"servicio":"gesprog","operacion":"Leer","id-programa":"p-0001"}
```
**Respuesta:**
```json
{"estado":"ok","programa":{"id-programa":"p-0001","nombre":"ejecutable","args":["..."],"env":["..."]}}
```

**Listar todos:**
```json
{"servicio":"gesprog","operacion":"Leer"}
```
**Respuesta:**
```json
{"estado":"ok","programas":["p-0001","p-0002"]}
```

### Actualizar

Reemplaza el binario de un programa existente. Los metadatos (args, env) se mantienen.

**Petición:**
```json
{"servicio":"gesprog","operacion":"Actualizar","id-programa":"p-0001","ruta":"/nuevo/ejecutable"}
```

**Respuesta (éxito):**
```json
{"estado":"ok"}
```

### Borrar

Elimina el programa del almacenamiento (binario y metadatos).

**Petición:**
```json
{"servicio":"gesprog","operacion":"Borrar","id-programa":"p-0001"}
```

**Respuesta (éxito):**
```json
{"estado":"ok"}
```

### Suspender

Pausa el servicio. En estado `Suspendido` solo se permite la operación `Leer`. Las demás devuelven `servicio suspendido`.

**Petición:**
```json
{"servicio":"gesprog","operacion":"Suspender"}
```

**Respuesta (éxito):**
```json
{"estado":"ok"}
```

### Reasumir

Reanuda la operación normal del servicio.

**Petición:**
```json
{"servicio":"gesprog","operacion":"Reasumir"}
```

**Respuesta (éxito):**
```json
{"estado":"ok"}
```

### Terminar

Finaliza el servicio de forma ordenada (cierra pipes y termina el proceso).

**Petición:**
```json
{"servicio":"gesprog","operacion":"Terminar"}
```

**Respuesta (éxito):**
```json
{"estado":"ok"}
```

### Respuesta de error

Cualquier operación puede devolver un error con el siguiente formato:

```json
{"estado":"error","mensaje":"<descripción del error en español>"}
```

Mensajes típicos:  
`"falta campo: ejecutable"`, `"programa no encontrado"`, `"servicio suspendido"`, `"transicion invalida"`, `"operacion desconocida"`.

## Máquina de estados

El servicio parte en estado `Corriendo`. Las transiciones válidas son:

- `Suspender` → `Suspendido` (solo desde `Corriendo`).
- `Reasumir` → `Corriendo` (solo desde `Suspendido`).
- `Terminar` → `Terminado` (desde `Corriendo` o `Suspendido`).

En estado `Suspendido`:
- `Leer` → permitido.
- Otras operaciones de escritura (`Guardar`, `Actualizar`, `Borrar`, `Suspender`, `Reasumir`, `Terminar`) → error `"servicio suspendido"`.

En estado `Terminado` el proceso finaliza.

## Almacenamiento (aralmac)

La ruta base se especifica con `-x`. Internamente, `gesprog` crea la subcarpeta `programas/` y guarda:

```
aralmac/programas/
├── p-0001.bin       ← copia del ejecutable original
├── p-0001.json      ← metadatos (id, nombre, args, env)
└── ...
```

- `p-XXXX.bin` : binario ejecutable.
- `p-XXXX.json` : metadatos en formato JSON, con indentación.

## Compilación y ejecución

### En Linux (WSL)

```bash
# Compilar
go build -o gesprog ./src/gesprog

# Crear directorio de almacenamiento y FIFOs
mkdir -p aralmac/programas
mkfifo /tmp/gesprog_in /tmp/gesprog_out

# Ejecutar servidor
./gesprog -p /tmp/gesprog_in -c /tmp/gesprog_out -x ./aralmac
```

### En Windows (PowerShell como administrador)

```powershell
# Compilar
go build -o gesprog.exe .\src\gesprog

# Crear directorio
mkdir .\aralmac\programas -Force

# Ejecutar servidor (full‑duplex, flag -c opcional)
.\gesprog.exe -p \\.\pipe\gesprog_pipe -x .\aralmac
```

## Prueba rápida

Una vez el servidor está corriendo, puedes enviar una petición desde otra terminal:

**Linux:**
```bash
echo '{"servicio":"gesprog","operacion":"Guardar","ejecutable":"/bin/echo"}' > /tmp/gesprog_in
cat /tmp/gesprog_out
```

**Windows (con PowerShell):**
```powershell
$pipe = new-object System.IO.Pipes.NamedPipeClientStream("\\.\pipe\gesprog_pipe")
$pipe.Connect()
$writer = new-object System.IO.StreamWriter($pipe)
$writer.WriteLine('{"servicio":"gesprog","operacion":"Guardar","ejecutable":"C:\\Windows\\System32\\calc.exe"}')
$writer.Flush()
$reader = new-object System.IO.StreamReader($pipe)
$reader.ReadLine()
```

La respuesta será algo como: `{"estado":"ok","id-programa":"p-0001"}`.

## Nota sobre la generación de IDs

Actualmente, la función `generarIDPrograma()` es un **placeholder** que siempre devuelve `"p-0001"`.  
En la integración final será reemplazada por `common.GenerarIDPrograma()` (proporcionado por el paquete `common` de Ana Sofia), que garantiza IDs únicos mediante bloqueo de archivo.

## Dependencias

- Solo utiliza la biblioteca estándar de Go, excepto `golang.org/x/sys/windows` (para Windows). Esta dependencia se descarga automáticamente con `go mod tidy`.
- El paquete `common` (en desarrollo por Ana Sofia) provee las funciones `AbrirPipes` para la comunicación multiplataforma. Las implementaciones actuales en `pipe_linux.go` y `pipe_windows.go` son temporales y funcionales.

## Código fuente

Los archivos que componen el servicio son:

- `main.go`         – punto de entrada, parsing de flags.
- `estado.go`       – máquina de estados (Corriendo, Suspendido, Terminado).
- `almacenamiento.go` – operaciones de disco (copiar, guardar, leer, listar, actualizar, borrar).
- `operaciones.go`  – despachador de comandos JSON, enlace con `estado` y `almacenamiento`.
- `servidor.go`     – bucle de escucha de pipes, llamada a `ProcesarPeticion`.

Todas las funciones y estructuras siguen la **Guía de Estilos** del proyecto (nombres camelCase/PascalCase, comentarios, manejo de errores, concurrencia con mutex).
