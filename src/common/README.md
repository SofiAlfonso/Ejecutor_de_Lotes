# common — Utilidades compartidas

Este paquete concentra todo el código que reutilizan los demás componentes (`ctrllt`, `gesfich`, `gesprog`, `ejecutor`). No contiene lógica de negocio.

## Archivos previstos

| Archivo      | Responsabilidad                                                                 |
|--------------|---------------------------------------------------------------------------------|
| `json.go`    | Serialización/deserialización de mensajes del protocolo (encode/decode JSON+\n) |
| `pipe.go`    | Apertura, lectura y escritura de named pipes en Linux (`mkfifo`) y Windows 11   |
| `lock.go`    | Bloqueo exclusivo/compartido de archivos (`flock` en Linux, `LockFile` en Win)  |
| `ids.go`     | Generación atómica de identificadores: `f-XXXX` (ficheros), `p-XXXX` (programas), `l-XXXX` (lotes) |
| `errors.go`  | Códigos de error estándar del sistema y función de construcción de respuestas de error |

## Notas de portabilidad

- `pipe.go` y `lock.go` usarán build tags (`//go:build linux` / `//go:build windows`) para separar las implementaciones por SO.
- `ids.go` usa `sync/atomic` para garantizar unicidad bajo concurrencia sin mutex.
