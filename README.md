## Instrucciones de uso

El repositorio cuenta con un **Makefile** que posee encapsulado diferentes comandos utilizados recurrentemente en el proyecto en forma de targets. Los targets se ejecutan mediante la invocación de:

- **make \<target\>**:
  Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de _debugging_ y _troubleshooting_.

Los targets disponibles son:

- **docker-compose-up**: Inicializa el ambiente de desarrollo (buildear docker images del servidor y cliente, inicializar la red a utilizar por docker, etc.) y arranca los containers de las aplicaciones que componen el proyecto.
- **docker-compose-down**: Realiza un `docker-compose stop` para detener los containers asociados al compose y luego realiza un `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene.
- **docker-compose-logs**: Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose.
- **docker-image**: Buildea las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para testear nuevos cambios en las imágenes antes de arrancar el proyecto.
- **build**: Compila la aplicación cliente para ejecución en el _host_ en lugar de en docker. La compilación de esta forma es mucho más rápida pero requiere tener el entorno de Golang instalado en la máquina _host_.

## Cómo ejecutar cada ejercicio

Antes de la ejecución, se debe cambiar a la rama correspondiente del ejercicio que se desea ejecutar. Por ejemplo, para el ejercicio 4:

```bash
git checkout ej4
```

### Ejercicio 1:

El comando: `./generar-compose.sh <filename> <client_num>` genera un archivo de Docker Compose con el nombre filename, que configura una cantidad de clientes especificada por client_num.

### Ejercicio 2:

Con el comando `make docker-compose-up` se inician los clientes y el servidor. Para visualizar los logs, usar `make docker-compose-logs`. Finalmente, se puede tener la ejecución con `make docker-compose-down`.

### Ejercicio 3:

Inicia el servidor y los clientes con `make docker-compose-up`. Luego, utiliza el comando `./validar-echo-server.sh` para verificar el correcto funcionamiento del servidor Echo.

### Ejercicio 4:

Inicia el servidor y los clientes con `make docker-compose-up`. Después, puedes enviar una señal **SIGTERM** a un cliente específico o al servidor con el comando `docker kill --signal=SIGTERM nombre_o_id`. Para observar el proceso de apagado del elemento seleccionado, utiliza `make docker-compose-logs`.

### Ejercicios 5, 6, 7 y 8:

Inicia los clientes con el comando `make docker-compose-up`, visualiza los logs con make `docker-compose-logs` y detén la ejecución con `make docker-compose-down`.

En el ejercicio 6, puedes calcular el tamaño máximo de cada campo enviado utilizando el comando `./hallar-maximo-campos.sh`. Este valor se utiliza para calcular maxAmount y asegurar que los paquetes no excedan 8Kb.

## Protocolo de comunicación:

El protocolo de comunicación implementado sigue estos pasos:

1 - Antes de enviar cualquier paquete, el cliente manda un paquete de 4 bytes que especifica el tamaño del siguiente paquete de datos. Este valor le indica al servidor cuántos bytes debe leer.

2 - Cuando el cliente inicia la comunicación con el servidor y quiere transferir los chunks con datos de apuestas, envía un mensaje **"BETDATA"**. Esto le avisa al servidor que la agencia va a comenzar la carga de apuestas.

3 - Después de enviar **"BETDATA"**, la agencia empieza a transferir los datos de las apuestas. Antes de cada chunk de datos, el cliente envía el tamaño del chunk. Cada chunk consiste de los campos de las apuestas, separados por comas, y al final de cada línea se incluye el ID de la agencia que envía los datos.

4 - Una vez que la agencia terminó de enviar todas sus apuestas, envía un **"0"** (en lugar del tamaño del chunk). Esto le indica al servidor que ya no quedan más apuestas por transferir.

5 - Al finalizar la transferencia de apuestas, el cliente solicita los resultados enviando un mensaje **"REQWINN"**. El servidor responde según si ya tiene o no los ganadores disponibles.

6 - Si los ganadores no están listos, el servidor responde con **"NOWINN"**. En ese caso, el cliente espera unos segundos antes de volver a intentar. Cuando los ganadores estén disponibles, el servidor envía un mensaje **"WINNERS"**, seguido de los documentos correspondientes a la agencia que solicitó los resultados.

7 - El cliente, al recibir el mensaje **"WINNERS"**, sabe que debe proceder a leer los paquetes que contienen los documentos de los ganadores.

### Nota: Para agregar nuevas agencias

Si es necesario sumar más agencias al sistema, hay que seguir estos pasos:

- Agregar los archivos de datos de las nuevas agencias (agency-n.csv).
- Ejecutar **./generar-compose.sh** especificando la cantidad de agencias deseadas en el archivo de salida.
- Actualizar la constante **NUM_AGENCIES** en `/server/common/constants.py/` con el nuevo número de agencias.

## Concurrencia y sincronización:

El servidor puede responder a varios clientes de forma simultánea gracias a un threadpool, que asigna un worker para manejar cada conexión entrante.

Utilizamos los siguientes locks para asegurar la sincronización entre los distintos hilos:

1 - handled_agencies_lock:
Este lock se usa para proteger el atributo `handled_agencies`. Para saber si ya se pueden devolver los resultados, se consulta este atributo para verificar si llegaron los datos de todas las agencias. Además, cada vez que una agencia termina de enviar todas sus apuestas, se incrementa este valor en 1, indicando que una nueva agencia completó la carga de sus datos.

2 - store_bets_lock:
Este lock se emplea para controlar el acceso a la función `store_bets`, ya que no es thread-safe. Con este lock nos aseguramos de que solo un hilo acceda a la función a la vez, evitando posibles conflictos de concurrencia.

3 - winners_lock:
Similar al lock anterior, pero en este caso se utiliza para controlar el acceso a la función `load_bets`, garantizando que solo un hilo a la vez pueda cargar las apuestas.
