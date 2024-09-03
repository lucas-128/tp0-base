NETWORK_NAME="tp0-base_testing_net"
SERVICE_NAME="server"  
PORT="12345"
MESSAGE="Test"

docker run --rm --network "$NETWORK_NAME" alpine sh -c "
  RESPONSE=\$(echo '$MESSAGE' | nc $SERVICE_NAME $PORT)
  if [ \"\$RESPONSE\" = \"$MESSAGE\" ]; then
    echo 'action: test_echo_server | result: success'
  else
    echo 'action: test_echo_server | result: fail'
  fi
"