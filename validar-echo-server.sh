NETWORK_NAME="tp0-base_testing_net"
SERVICE_NAME="server"
PORT="12345"
MESSAGE="Test"
TIMEOUT_DURATION="5"  


NETWORK_EXISTS=$(docker network ls --filter name="$NETWORK_NAME" --format '{{.Name}}')

if [ "$NETWORK_EXISTS" != "$NETWORK_NAME" ]; then
  echo 'action: test_echo_server | result: net not exist'
  exit 1
fi

SERVICE_REACHABLE=$(docker run --rm --network "$NETWORK_NAME" alpine sh -c "nc -z $SERVICE_NAME $PORT 2>/dev/null && echo 'reachable' || echo 'unreachable'")

if [ "$SERVICE_REACHABLE" != "reachable" ]; then
  echo 'action: test_echo_server | result: unreach'
  exit 1
fi

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
