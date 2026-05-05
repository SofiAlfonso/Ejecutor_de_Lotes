# gesprog — Gestor de programas

## Descripción

`gesprog` administra los binarios ejecutables dentro del almacenamiento local (`aralmac/programas/`). Asigna identificadores únicos `p-XXXX` y controla el acceso concurrente mediante bloqueo de archivos.

## Sinopsis

```
gesprog -p <pipe-req> -g <pipe-res> -x <ruta_aralmac>
```

| Flag | Significado                                    |
|------|------------------------------------------------|
| `-p` | Pipe de peticiones entrantes (desde ctrllt)    |
| `-g` | Pipe de respuestas salientes (hacia ctrllt)    |
| `-x` | Ruta raíz del almacenamiento (`aralmac/`)      |

## Operaciones

| Acción      | Descripción                                              | Retorna       |
|-------------|----------------------------------------------------------|---------------|
| `guardar`   | Copia el binario local al almacén                        | `p-XXXX`      |
| `leer`      | Devuelve metadatos de un programa o lista completa       | JSON del programa / lista |
| `actualizar`| Reemplaza el binario con una nueva versión               | `ok`          |
| `borrar`    | Elimina el programa del almacén                          | `ok` / error  |
| `suspender` | Detiene la aceptación de nuevas peticiones de escritura  | `ok`          |
| `reasumir`  | Reanuda la aceptación de todas las peticiones            | `ok`          |
| `terminar`  | Cierra pipes y termina el proceso                        | —             |

## Estados del proceso

```
Inicio → Corriendo ⇄ Suspendido → Terminado
```

- **Corriendo:** atiende todas las operaciones.
- **Suspendido:** la operación `leer` sí está permitida; las operaciones de escritura (`guardar`, `actualizar`, `borrar`) responden con error de servicio suspendido.
- La transición a **Terminado** solo puede hacerse desde **Corriendo**.
