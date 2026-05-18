//go:build windows

package common

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// AbrirPipes crea un named pipe full-duplex en Windows y espera a que
// un cliente se conecte.
//
// El parámetro pipeRespuestas se ignora porque en Windows un solo pipe
// sirve para leer y escribir.
//
// Uso esperado (servidor):
//
//	entrada, salida, err := common.AbrirPipes(`\\.\pipe\gesprog_pipe`, "")
//
// Los dos *os.File devueltos apuntan al mismo handle subyacente.
func AbrirPipes(pipePeticiones, pipeRespuestas string) (*os.File, *os.File, error) {
	// Convertir el nombre del pipe a UTF-16 (requerido por la API de Windows).
	pipeName, err := windows.UTF16PtrFromString(pipePeticiones)
	if err != nil {
		return nil, nil, fmt.Errorf("AbrirPipes: nombre de pipe inválido %q: %w", pipePeticiones, err)
	}

	// Crear el named pipe como servidor full-duplex.
	// Se permite una sola instancia simultánea (nMaxInstances = 1) porque
	// ctrllt es el único cliente que se conecta a cada servicio.
	handle, err := windows.CreateNamedPipe(
		pipeName,
		windows.PIPE_ACCESS_DUPLEX,
		windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT,
		1,    // máximo de instancias simultáneas
		4096, // tamaño del búfer de salida
		4096, // tamaño del búfer de entrada
		0,    // tiempo de espera predeterminado (50 ms)
		nil,  // atributos de seguridad predeterminados
	)
	if err != nil {
		return nil, nil, fmt.Errorf("AbrirPipes: CreateNamedPipe falló: %w", err)
	}

	// Esperar a que el cliente se conecte (bloqueante).
	if err := windows.ConnectNamedPipe(handle, nil); err != nil && err != windows.ERROR_PIPE_CONNECTED {
		windows.CloseHandle(handle)
		return nil, nil, fmt.Errorf("AbrirPipes: ConnectNamedPipe falló: %w", err)
	}

	// Envolver el handle en un *os.File para poder usar bufio sobre él.
	f := os.NewFile(uintptr(handle), pipePeticiones)
	// Devolvemos el mismo archivo para lectura y escritura (full-duplex).
	return f, f, nil
}
