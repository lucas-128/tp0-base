import sys

def generate_docker_compose(output_file, num_clients):
    compose_content = """ 

services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    volumes:
      - ./server/config.ini:/config.ini
    networks:
      - testing_net
"""

    for i in range(1, num_clients + 1):
        client_name = f"client{i}"
        client_service = f"""  {client_name}:
    container_name: {client_name}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - CLI_LOG_LEVEL=DEBUG
    volumes:
      - ./client/config.yaml:/config.yaml
    networks:
      - testing_net
    depends_on:
      - server
"""
        compose_content += client_service

    compose_content += """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

    with open(output_file, 'w') as file:
        file.write(compose_content)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Uso: python3 mi-generador.py <output_file_name> <number_of_clients>")
        sys.exit(1)

    output_file = sys.argv[1]
    num_clients = int(sys.argv[2])

    generate_docker_compose(output_file, num_clients)
