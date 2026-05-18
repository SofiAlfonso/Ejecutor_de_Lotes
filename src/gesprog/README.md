¡Perfecto! Ahora que me has mostrado los archivos actualizados de `gesprog` (ya usan `common.GenerarIDPrograma()`, `common.InitIDs` y `common.AbrirPipes`), el siguiente paso es actualizar su `README.md` para reflejar la integración real.  

A continuación te proporciono el contenido **actualizado** para `src/gesprog/README.md`. Puedes copiarlo y reemplazar el archivo existente.

---

```markdown
# gesprog — Gestor de programas

## Descripción

`gesprog` es el servicio encargado de registrar, consultar, actualizar, borrar y almacenar programas ejecutables dentro del sistema de lotes.  
Guarda copias de los binarios en `aralmac/programas/` junto con sus metadatos (argumentos y variables de entorno).  
Asigna identificadores únicos con formato `p-XXXX` mediante el paquete `common`.

El servicio se comunica exclusivamente a través de **tuberías nombradas** (named pipes / FIFOs).  
- En **Linux** se usan dos FIFOs (half‑duplex): uno para peticiones (`-p`) y otro para respuestas (`-c`).  
- En **Windows** se usa un único named pipe full‑duplex (el flag `-c` es opcional).

La implementación utiliza las funciones `common.AbrirPipes` y `common.GenerarIDPrograma` (junto con `common.InitIDs`).

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
- Otras operaciones (`Guardar`, `Actualizar`, `Borrar`, `Suspender`, `Reasumir`, `Terminar`) → error `"servicio suspendido"`.

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

## Generación de IDs

Los identificadores `p-XXXX` se generan automáticamente mediante `common.GenerarIDPrograma()`, que escanea el directorio `aralmac/programas/` y asigna el siguiente número disponible. No se utiliza ningún placeholder; la integración con `common` ya está completa.

## Dependencias

- Solo utiliza la biblioteca estándar de Go, excepto `golang.org/x/sys/windows` (para Windows). Esta dependencia se descarga automáticamente con `go mod tidy`.
- El paquete `common` (ya implementado) proporciona `AbrirPipes` para la comunicación multiplataforma y `GenerarIDPrograma` para la generación atómica de IDs. El servicio llama a `common.InitIDs` al arrancar.

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
$writer.WriteLine('{"servicio":"gesprog","operacion":"Guardar","ejecutable":"C:\\temp\\test_progs\\p1.exe"}')
$writer.Flush()
$reader = new-object System.IO.StreamReader($pipe)
$reader.ReadLine()
```

La respuesta será algo como: `{"estado":"ok","id-programa":"p-0001"}`.

## Código fuente

Los archivos que componen el servicio son:

- `main.go`         – punto de entrada, parsing de flags, inicialización de `common`.
- `estado.go`       – máquina de estados (Corriendo, Suspendido, Terminado).
- `almacenamiento.go` – operaciones de disco (copiar, guardar, leer, listar, actualizar, borrar). Usa `common.GenerarIDPrograma`.
- `operaciones.go`  – despachador de comandos JSON, enlace con `estado` y `almacenamiento`.
- `servidor.go`     – bucle de escucha de pipes mediante `common.AbrirPipes`.

