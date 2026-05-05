# Ejecutor de Lotes

Sistema de ejecución de procesos por lotes con arquitectura de microservicios comunicados mediante named pipes.

**Integrantes:**
- Ana Sofia Alfonso Moncada
- Maria Mercedes Olaya Lopez

**Tecnología:** Go | **SO:** Linux y Windows 11 | **Comunicación:** Named pipes half-duplex | **Protocolo:** JSON con `\n` como delimitador

---

## Componentes

| Componente  | Descripción                                           |
|-------------|-------------------------------------------------------|
| cliente     | Interfaz de usuario (dado por el profesor)            |
| ctrllt      | Pasarela central que recibe y enruta peticiones       |
| gesfich     | CRUD de ficheros de datos en aralmac                  |
| gesprog     | CRUD de programas ejecutables en aralmac              |
| ejecutor    | Lanza y controla procesos de lotes                    |
| aralmac     | Almacenamiento en disco (directorio local, ignorado)  |

---

## Arquitectura

```
cliente
   |
   v
ctrllt  ──────────────────────────────────────┐
   |                                           |
   |──── gesfich ───┐                          |
   |──── gesprog ───┼──── aralmac (disco)      |
   └──── ejecutor ──┘                          |
                                               |
         (respuestas vuelven por ctrllt) ──────┘
```

---

## Estructura del repositorio

```
Ejecutor_de_lotes/
├── docs/
│   └── Diseño.md
├── src/
│   ├── common/       # utilidades compartidas
│   ├── ctrllt/       # pasarela central
│   ├── gesfich/      # CRUD ficheros
│   ├── gesprog/      # CRUD programas
│   └── ejecutor/     # lanzador de lotes
├── tests/
├── scripts/
│   └── start.sh
├── go.mod
└── README.md
```

---

## Protocolo de mensajes

Todos los mensajes se envían como una línea JSON terminada en `\n`.

```json
{
  "version":    "1.0",
  "client_id":  "cli-01",
  "request_id": "req-0042",
  "service":    "gesfich",
  "action":     "crear",
  "payload": {
    "nombre": "datos.txt",
    "ruta_local": "/home/usuario/datos.txt"
  }
}
```

---

## Orden de arranque

1. Levantar `gesfich`, `gesprog` y `ejecutor` (crean sus pipes y esperan).
2. Levantar `ctrllt` (conecta con todos los servicios).
3. Levantar el `cliente` (conecta con ctrllt).

Ver `scripts/start.sh` para arranque automatizado.

---

## Entregas

| Entrega  | Contenido                  | Fecha             |
|----------|----------------------------|-------------------|
| Primera  | `docs/Diseño.md`           | 5 de mayo de 2026 |
| Segunda  | Implementación completa    | Por definir       |

---

## Documentación

- [Diseño del sistema](docs/Diseño.md)
