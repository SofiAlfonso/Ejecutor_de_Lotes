// Package common provee utilidades compartidas por todos los servicios:
// comunicación por named pipes y generación atómica de identificadores.
package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// mu protege la generación de IDs ante llamadas concurrentes dentro del mismo proceso.
var mu sync.Mutex

// aralmacPath es la ruta base del almacenamiento. Se establece con InitIDs.
var aralmacPath string

// InitIDs establece la ruta base del aralmac.
// Debe llamarse desde main() de cada servicio antes de generar cualquier ID.
//
//	common.InitIDs("/ruta/al/aralmac")
func InitIDs(ruta string) {
	aralmacPath = ruta
}

// GenerarIDFichero genera el siguiente identificador disponible con formato f-XXXX.
// Escanea aralmac/ficheros/ para determinar el número máximo ya usado.
func GenerarIDFichero() (string, error) {
	return generarID(filepath.Join(aralmacPath, "ficheros"), "f")
}

// GenerarIDPrograma genera el siguiente identificador disponible con formato p-XXXX.
// Escanea aralmac/programas/ para determinar el número máximo ya usado.
func GenerarIDPrograma() (string, error) {
	return generarID(filepath.Join(aralmacPath, "programas"), "p")
}

// GenerarIDEjecucion genera el siguiente identificador disponible con formato e-XXXX.
// Escanea aralmac/ejecuciones/ para determinar el número máximo ya usado.
func GenerarIDEjecucion() (string, error) {
	return generarID(filepath.Join(aralmacPath, "ejecuciones"), "e")
}

// generarID es la función interna compartida por los tres generadores públicos.
// Recorre el directorio dir buscando archivos cuyo nombre sigue el patrón
// <prefijo>-NNNN.<ext> y devuelve el siguiente número formateado como
// <prefijo>-NNNN (con ceros a la izquierda hasta 4 dígitos).
//
// El mutex mu serializa el acceso cuando varios goroutines del mismo proceso
// llaman a esta función simultáneamente; es suficiente para el proyecto porque
// cada tipo de ID (f, p, e) es generado por un único servicio.
func generarID(dir, prefijo string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	// Crear directorio si no existe (primera ejecución).
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("generarID %s: no se pudo crear directorio: %w", prefijo, err)
	}

	entradas, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("generarID %s: no se pudo leer directorio: %w", prefijo, err)
	}

	maximo := 0
	patron := prefijo + "-"

	for _, e := range entradas {
		nombre := e.Name()
		// El nombre debe empezar con "f-", "p-" o "e-".
		if !strings.HasPrefix(nombre, patron) {
			continue
		}
		// Eliminar la extensión para quedarnos con, p.ej., "p-0003".
		sinExt := strings.TrimSuffix(nombre, filepath.Ext(nombre))
		// sinExt tiene la forma "<prefijo>-NNNN"; separar por "-".
		partes := strings.SplitN(sinExt, "-", 2)
		if len(partes) != 2 {
			continue
		}
		n, err := strconv.Atoi(partes[1])
		if err != nil {
			continue
		}
		if n > maximo {
			maximo = n
		}
	}

	return fmt.Sprintf("%s-%04d", prefijo, maximo+1), nil
}