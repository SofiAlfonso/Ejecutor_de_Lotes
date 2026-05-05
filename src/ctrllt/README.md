# ctrllt — Pasarela central

## Descripción

`ctrllt` es el único punto de entrada para el cliente. Recibe todas las peticiones, registra el pipe de respuesta individual de cada cliente y enruta cada mensaje al servicio correspondiente (`gesfich`, `gesprog` o `ejecutor`). Devuelve la respuesta al pipe privado del cliente que originó la petición.

## Sinopsis

```
ctrllt -c <pipe-clientes-req> \
       -f <pipe-fich-req>     \
       -b <pipe-fich-res>     \
       -p <pipe-prog-req>     \
       -g <pipe-prog-res>     \
       -e <pipe-ejec-req>     \
       -d <pipe-ejec-res>
```

| Flag | Significado                          |
|------|--------------------------------------|
| `-c` | Pipe compartido donde escuchan clientes |
| `-f` | Pipe de peticiones hacia gesfich     |
| `-b` | Pipe de respuestas desde gesfich     |
| `-p` | Pipe de peticiones hacia gesprog     |
| `-g` | Pipe de respuestas desde gesprog     |
| `-e` | Pipe de peticiones hacia ejecutor    |
| `-d` | Pipe de respuestas desde ejecutor    |

## Responsabilidades

1. Escuchar el pipe compartido `-c` en un bucle continuo.
2. Al recibir un mensaje, registrar el `client_id` y su pipe de respuesta individual.
3. Enrutar la petición al servicio indicado en el campo `service` del JSON.
4. Leer la respuesta del servicio y reenviarla al pipe privado del cliente.
5. Manejar múltiples clientes simultáneos con goroutines (una por petición en vuelo).

## Estados del proceso

```
Inicio → Corriendo → Terminando → Terminado
```

- **Inicio:** abre y verifica todos los pipes antes de aceptar clientes.
- **Corriendo:** acepta y despacha peticiones.
- **Terminando:** deja de aceptar nuevas peticiones; espera a que terminen las goroutines activas.
- **Terminado:** cierra todos los pipes y sale.
