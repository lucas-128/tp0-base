NETWORK_NAME="tp0-base_testing_net"
SERVICE_NAME="server"  
PORT="12345"
MESSAGE="Test"
TIMEOUT_DURATION="5"  # Timeout duration in seconds (without 's' suffix)

docker run --rm --network "$NETWORK_NAME" busybox sh -c "
  RESPONSE=\$(echo \"$MESSAGE\" | timeout $TIMEOUT_DURATION nc $SERVICE_NAME $PORT)
  if [ -z \"\$RESPONSE\" ]; then
    echo 'action: test_echo_server | result: fail'
  elif [ \"\$RESPONSE\" = \"$MESSAGE\" ]; then
    echo 'action: test_echo_server | result: success'
  else
    echo 'action: test_echo_server | result: fail'
  fi
"
