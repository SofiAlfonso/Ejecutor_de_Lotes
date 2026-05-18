# common — Utilidades compartidas

Este paquete concentra funcionalidades reutilizables por los demás servicios (`ctrllt`, `gesfich`, `gesprog`, `ejecutor`).

## Archivos reales

| Archivo           | Descripción                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| `ids.go`          | Generación de identificadores únicos `f-XXXX`, `p-XXXX`, `e-XXXX`.          |
| `pipe_linux.go`   | Implementación de `AbrirPipes` para Linux (FIFOs half‑duplex).              |
| `pipe_windows.go` | Implementación de `AbrirPipes` para Windows (named pipe full‑duplex).       |

## Uso

### Inicialización
Cada servicio debe llamar a `common.InitIDs` al arrancar, pasando la ruta raíz del almacenamiento (`aralmac`).  
Ejemplo:
```go
common.InitIDs(*aralmac)
```

### Generación de IDs
```go
idFichero, _ := common.GenerarIDFichero()   // formato f-0001
idPrograma, _ := common.GenerarIDPrograma() // formato p-0001
idEjecucion, _ := common.GenerarIDEjecucion() // formato e-0001
```
- La generación se basa en escanear los subdirectorios `ficheros/`, `programas/` y `ejecuciones/` dentro de `aralmac`.
- El número se incrementa automáticamente, asegurando unicidad en el mismo proceso (el mutex `mu` evita carreras en el mismo servicio).

### Comunicación por pipes
```go
entrada, salida, err := common.AbrirPipes(pipePeticiones, pipeRespuestas)
```
- **Linux**: se esperan dos nombres de FIFO (half‑duplex). La función los crea si no existen y abre `pipePeticiones` en modo lectura, `pipeRespuestas` en modo escritura. La llamada bloquea hasta que el otro extremo (cliente) abra los FIFOs.
- **Windows**: se usa un solo pipe full‑duplex. El parámetro `pipeRespuestas` se ignora. La función crea el pipe y espera a que un cliente se conecte (`ConnectNamedPipe`). Devuelve el mismo `*os.File` para lectura y escritura.

Los descriptores devueltos pueden usarse directamente con `bufio.NewScanner` y `bufio.NewWriter`.

## Dependencias
- Para Windows se requiere `golang.org/x/sys/windows`. Se descarga automáticamente con `go mod tidy`.
- No hay dependencias adicionales en Linux (solo biblioteca estándar).

## Nota de diseño
- La generación de IDs **no utiliza bloqueo de archivo** entre procesos diferentes porque cada tipo de ID es generado por un único servicio. La exclusión mutua dentro del mismo proceso se logra con `sync.Mutex`.
- El servidor (el que llama a `AbrirPipes`) atiende **una sola conexión** y finaliza cuando el cliente cierra el pipe. Esto es suficiente para la arquitectura donde `ctrllt` mantiene la conexión abierta.

