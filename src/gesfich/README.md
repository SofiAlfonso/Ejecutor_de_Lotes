# gesfich — Gestor de ficheros

## Descripción

`gesfich` administra ficheros de datos dentro del almacenamiento local (`aralmac/ficheros/`). Asigna identificadores únicos `f-XXXX`, controla el contador de referencias y serializa el acceso concurrente mediante bloqueo de archivos.

## Sinopsis

```
gesfich -f <pipe-req> -b <pipe-res> -x <ruta_aralmac>
```

| Flag | Significado                                    |
|------|------------------------------------------------|
| `-f` | Pipe de peticiones entrantes (desde ctrllt)    |
| `-b` | Pipe de respuestas salientes (hacia ctrllt)    |
| `-x` | Ruta raíz del almacenamiento (`aralmac/`)      |

## Operaciones

| Acción      | Descripción                                              | Retorna       |
|-------------|----------------------------------------------------------|---------------|
| `crear`     | Copia archivo local al almacén                           | `f-XXXX`      |
| `leer`      | Devuelve metadatos de un fichero o lista completa        | JSON del fichero / lista |
| `actualizar`| Reemplaza el contenido con una nueva copia del archivo local | `ok`       |
| `borrar`    | Elimina el fichero solo si `refcount == 0`               | `ok` / error  |
| `suspender` | Detiene la aceptación de nuevas peticiones               | `ok`          |
| `reasumir`  | Reanuda la aceptación de peticiones                      | `ok`          |
| `terminar`  | Cierra pipes y termina el proceso                        | —             |

## Estados del proceso

```
Inicio → Corriendo ⇄ Suspendido → Terminado
```

- **Corriendo:** atiende todas las operaciones.
- **Suspendido:** solo responde con error de servicio no disponible; no modifica el almacén.
- La transición a **Terminado** solo puede hacerse desde **Corriendo**.
