# tests — Plan de pruebas

Las pruebas se organizan por categoría en subcarpetas. Cada subcarpeta contendrá casos de prueba en Go (`_test.go`) o scripts de integración.

## Estructura prevista

```
tests/
├── happy_path/     # flujos normales de extremo a extremo
├── errors/         # entradas inválidas y condiciones de error
├── concurrency/    # múltiples clientes / goroutines simultáneas
└── states/         # transiciones de estado de cada servicio
```

## Casos de prueba mínimos

### happy_path/
- Crear un fichero, leerlo y borrarlo.
- Guardar un programa, ejecutarlo y consultar su estado hasta que termine.
- Listar ficheros y programas (almacén vacío y con elementos).
- Ejecutar varios lotes en paralelo y obtener estados.

### errors/
- Borrar un fichero con `refcount > 0` → debe retornar error.
- Solicitar leer un ID inexistente (`f-9999`, `p-9999`, `l-9999`) → error.
- Enviar JSON malformado al pipe → el servicio no debe caerse.
- Ejecutar un programa que no existe en aralmac → error inmediato.
- Acción desconocida en campo `action` → respuesta de error estándar.

### concurrency/
- 10 clientes crean ficheros simultáneamente → IDs únicos sin colisión.
- Leer y escribir el mismo fichero concurrentemente → sin corrupción.
- Ejecutar 20 lotes al mismo tiempo → todos reciben `l-XXXX` distintos.
- ctrllt enruta peticiones de múltiples clientes sin mezclar respuestas.

### states/
- gesfich: Corriendo → Suspendido → petición rechazada → Corriendo → petición atendida.
- gesprog: Suspendido permite `leer` pero rechaza `guardar`/`actualizar`/`borrar`.
- ejecutor: Suspendido rechaza `ejecutar` pero responde a `estado`.
- ejecutor: `parar` espera lotes activos antes de terminar.
- ejecutor: `terminar` mata lotes activos inmediatamente.
