// Package main implementa el servicio ejecutor de procesos de lotes.
package main

import (
	"bufio"
	"fmt"
	"io"

	"github.com/SofiAlfonso/Ejecutor_de_Lotes/src/common"
)

// maxMsgLen es el tamaño máximo de mensaje definido en el protocolo.
const maxMsgLen = 4096

// Servidor escucha peticiones desde el pipe de entrada y escribe respuestas
// en el pipe de salida. Lanza una goroutine por petición para no bloquear
// ejecuciones paralelas. Corre hasta que el servicio pase a estado Terminado.
func Servidor(pipePeticiones, pipeRespuestas string) error {
	entrada, salida, err := common.AbrirPipes(pipePeticiones, pipeRespuestas)
	if err != nil {
		return fmt.Errorf("servidor: %w", err)
	}
	defer entrada.Close()
	defer salida.Close()

	scanner := bufio.NewScanner(entrada)
	scanner.Buffer(make([]byte, maxMsgLen), maxMsgLen)
	writer := bufio.NewWriter(salida)

	for scanner.Scan() {
		linea := scanner.Bytes()
		if len(linea) == 0 {
			continue
		}

		respuesta := ProcesarPeticion(linea)

		if _, err := writer.Write(respuesta); err != nil {
			return fmt.Errorf("servidor: error escribiendo respuesta: %w", err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("servidor: error escribiendo newline: %w", err)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("servidor: error en flush: %w", err)
		}

		if ServicioEstaTerminado() {
			break
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("servidor: error leyendo pipe: %w", err)
	}
	return nil
}
