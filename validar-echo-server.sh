#!/bin/bash

NETWORK_NAME="tp0-base_testing_net"
SERVICE_NAME="server"
PORT="12345"
MESSAGE="Test"
TIMEOUT_DURATION="5"  # Timeout duration in seconds

# Check if the Docker network exists
NETWORK_EXISTS=$(docker network ls --filter name="$NETWORK_NAME" --format '{{.Name}}')

if [ "$NETWORK_EXISTS" != "$NETWORK_NAME" ]; then
  echo 'action: test_echo_server | result: fail'
  exit 1
fi

# Check if the service is reachable in the network
SERVICE_REACHABLE=$(docker run --rm --network "$NETWORK_NAME" alpine sh -c "nc -z $SERVICE_NAME $PORT 2>/dev/null && echo 'reachable' || echo 'unreachable'")

if [ "$SERVICE_REACHABLE" != "reachable" ]; then
  echo 'action: test_echo_server | result: fail'
  exit 1
fi

# If network and service are good, run the test
docker run --rm --network "$NETWORK_NAME" alpine sh -c "
  RESPONSE=\$(echo '$MESSAGE' | nc -w $TIMEOUT_DURATION $SERVICE_NAME $PORT)
  if [ \"\$RESPONSE\" = \"$MESSAGE\" ]; then
    echo 'action: test_echo_server | result: success'
  else
    if [ \$? -eq 1 ]; then
      echo 'action: test_echo_server | result: fail'
    else
      echo 'action: test_echo_server | result: fail'
    fi
  fi
"
