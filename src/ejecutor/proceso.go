package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/SofiAlfonso/Ejecutor_de_Lotes/src/common"
)

// cerrarArchivos cierra una lista de archivos ignorando los nil.
func cerrarArchivos(archivos []*os.File) {
	for _, f := range archivos {
		if f != nil {
			f.Close()
		}
	}
}

// LanzarPipeline lanza una cadena de procesos conectados por tuberias anonimas.
// El stdin del primero viene de idStdin, el stdout del ultimo va a idStdout.
// El stderr de TODOS va al mismo idStderr.
// Devuelve un id-ejecucion que representa al grupo completo.
func LanzarPipeline(idProgramas []string, idStdin, idStdout, idStderr string) (string, error) {
	if len(idProgramas) == 0 {
		return "", fmt.Errorf("no se pudo ejecutar el programa: lista de programas vacia")
	}

	// Verificar que todos los programas existen antes de abrir ficheros
	type binMeta struct {
		ruta string
		meta metadataPrograma
	}
	programas := make([]binMeta, len(idProgramas))
	for i, id := range idProgramas {
		ruta, meta, err := VerificarPrograma(id)
		if err != nil {
			return "", fmt.Errorf("no se pudo ejecutar el programa: %s: %w", id, err)
		}
		programas[i] = binMeta{ruta, meta}
	}

	// --- Ficheros de E/S (opcionales) ---
	var fIn, fOut, fErr *os.File
	var archivos []*os.File

	if idStdin != "" {
		ruta, err := VerificarFichero(idStdin)
		if err != nil {
			return "", err
		}
		fIn, err = os.Open(ruta)
		if err != nil {
			return "", fmt.Errorf("no se pudo abrir stdin: %w", err)
		}
		archivos = append(archivos, fIn)
	}

	if idStdout != "" {
		ruta, err := VerificarFichero(idStdout)
		if err != nil {
			cerrarArchivos(archivos)
			return "", err
		}
		fOut, err = os.OpenFile(ruta, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			cerrarArchivos(archivos)
			return "", fmt.Errorf("no se pudo abrir stdout: %w", err)
		}
		archivos = append(archivos, fOut)
	}

	if idStderr != "" {
		ruta, err := VerificarFichero(idStderr)
		if err != nil {
			cerrarArchivos(archivos)
			return "", err
		}
		fErr, err = os.OpenFile(ruta, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			cerrarArchivos(archivos)
			return "", fmt.Errorf("no se pudo abrir stderr: %w", err)
		}
		archivos = append(archivos, fErr)
	}

	// Generar id-ejecucion para el grupo
	idEjecucion, err := common.GenerarIDEjecucion()
	if err != nil {
		cerrarArchivos(archivos)
		return "", fmt.Errorf("no se pudo ejecutar el programa: %w", err)
	}

	n := len(programas)
	cmds := make([]*exec.Cmd, n)

	// Construir todos los comandos
	for i, p := range programas {
		cmds[i] = exec.Command(p.ruta, p.meta.Args...)
		cmds[i].Env = append(os.Environ(), p.meta.Env...)
		// Asignar stderr (puede ser nil si no se abrió)
		cmds[i].Stderr = fErr
	}

	// Asignar stdin del primer proceso (si se abrió)
	if fIn != nil {
		cmds[0].Stdin = fIn
	}

	// Crear tuberías anónimas entre procesos consecutivos
	pipes := make([]*os.File, 0, 2*(n-1))
	for i := 0; i < n-1; i++ {
		pr, pw, err := os.Pipe()
		if err != nil {
			cerrarArchivos(archivos)
			for _, p := range pipes {
				p.Close()
			}
			return "", fmt.Errorf("no se pudo crear tuberia entre procesos: %w", err)
		}
		cmds[i].Stdout = pw
		cmds[i+1].Stdin = pr
		pipes = append(pipes, pr, pw)
	}

	// Asignar stdout del último proceso (si se abrió)
	if fOut != nil {
		cmds[n-1].Stdout = fOut
	}

	// Lanzar todos los procesos
	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			// Matar los que ya arrancaron
			for j := 0; j < i; j++ {
				if cmds[j].Process != nil {
					cmds[j].Process.Kill()
				}
			}
			cerrarArchivos(archivos)
			for _, p := range pipes {
				p.Close()
			}
			//MarcarTerminado(idEjecucion, -1)
			return "", fmt.Errorf("no se pudo ejecutar el programa: %s: %w", idProgramas[i], err)
		}
	}

	// Registrar el grupo con todos los comandos
	info := RegistrarProceso(idEjecucion, idProgramas[0], cmds)
	_ = GuardarEjecucion(info)

	// Cerrar write-ends en el padre para que los hijos reciban EOF
	for i := 0; i < n-1; i++ {
		pipes[2*i+1].Close()
	}

	// Goroutine monitor: espera que TODOS terminen
	go func() {
		defer cerrarArchivos(archivos)
		defer func() {
			for i := 0; i < n-1; i++ {
				pipes[2*i].Close() // cerrar read-ends
			}
		}()

		codigoSalida := 0
		for i, cmd := range cmds {
			err := cmd.Wait()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					// Tomar el codigo del ultimo proceso (convencion Unix)
					if i == n-1 {
						codigoSalida = exitErr.ExitCode()
					}
				} else {
					codigoSalida = -1
				}
			}
		}
		MarcarTerminado(idEjecucion, codigoSalida)
		_ = GuardarEjecucion(info)
	}()

	return idEjecucion, nil
}

// MatarProceso termina forzosamente todos los procesos del pipeline.
func MatarProceso(idEjecucion string) error {
	info, err := ObtenerProceso(idEjecucion)
	if err != nil {
		return fmt.Errorf("proceso no encontrado")
	}
	info.mu.RLock()
	terminado := info.Terminado
	cmds := info.Cmds
	info.mu.RUnlock()

	if terminado || len(cmds) == 0 {
		return fmt.Errorf("proceso no encontrado o ya terminado")
	}

	// Matar todos los procesos del pipeline
	for _, cmd := range cmds {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	// Marcar como terminado inmediatamente (código -1 indica "matado")
	info.mu.Lock()
	if !info.Terminado {
		info.Estado = ProcesoTerminado
		info.CodigoSalida = -1
		info.Terminado = true
		info.mu.Unlock()

		// Decrementar contador de procesos activos
		muProcesos.Lock()
		contadorActivos--
		activos := contadorActivos
		muProcesos.Unlock()

		// Si el servicio estaba en Parando y ya no hay activos, terminarlo
		if activos == 0 {
			servicio.mu.RLock()
			parando := servicio.current == estadoParando
			servicio.mu.RUnlock()
			if parando {
				TerminarServicio()
			}
		}
	} else {
		info.mu.Unlock()
	}

	return nil
}
