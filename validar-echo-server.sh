NETWORK_NAME="tp0_testing_net"
SERVICE_NAME="server"
PORT="12345"
MESSAGE="Test"
TIMEOUT_DURATION="5"  

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
