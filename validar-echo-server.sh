NETWORK_NAME="tp0-base_testing_net"
SERVICE_NAME="server"  
PORT="12345"
MESSAGE="Test"
TIMEOUT_DURATION="5s" 

docker run --rm --network "$NETWORK_NAME" busybox sh -c "
  RESPONSE=\$(timeout $TIMEOUT_DURATION sh -c 'echo \"$MESSAGE\" | nc $SERVICE_NAME $PORT')
  if [ -z \"\$RESPONSE\" ]; then
    echo 'action: test_echo_server | result: fail'
  elif [ \"\$RESPONSE\" = \"$MESSAGE\" ]; then
    echo 'action: test_echo_server | result: success'
  else
    echo 'action: test_echo_server | result: fail'
  fi
"
