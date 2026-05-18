//go:build linux

package common

import (
	"fmt"
	"os"
	"syscall"
)

// AbrirPipes abre los dos FIFOs para comunicación half-duplex en Linux.
// Crea los FIFOs si no existen todavía.
//
// Uso esperado (servidor):
//
//	entrada, salida, err := common.AbrirPipes("/tmp/svc_in", "/tmp/svc_out")
//
// La llamada bloquea hasta que el otro extremo (cliente / ctrllt) abra
// los FIFOs desde su lado, lo cual es el comportamiento estándar de los FIFOs.
func AbrirPipes(pipePeticiones, pipeRespuestas string) (*os.File, *os.File, error) {
	// Crear FIFOs si aún no existen.
	if err := crearFIFO(pipePeticiones); err != nil {
		return nil, nil, err
	}
	if err := crearFIFO(pipeRespuestas); err != nil {
		return nil, nil, err
	}

	// Abrir FIFO de entrada en modo solo-lectura.
	// Bloquea hasta que el escritor (ctrllt) abra su extremo.
	entrada, err := os.OpenFile(pipePeticiones, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("AbrirPipes: abrir entrada %q: %w", pipePeticiones, err)
	}

	// Abrir FIFO de salida en modo solo-escritura.
	// Bloquea hasta que el lector (ctrllt) abra su extremo.
	salida, err := os.OpenFile(pipeRespuestas, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		entrada.Close()
		return nil, nil, fmt.Errorf("AbrirPipes: abrir salida %q: %w", pipeRespuestas, err)
	}

	return entrada, salida, nil
}

// crearFIFO crea un FIFO en la ruta indicada si todavía no existe.
func crearFIFO(ruta string) error {
	if _, err := os.Stat(ruta); os.IsNotExist(err) {
		if err := syscall.Mkfifo(ruta, 0666); err != nil {
			return fmt.Errorf("crearFIFO %q: %w", ruta, err)
		}
	}
	return nil
}
