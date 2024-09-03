NETWORK_NAME="tp0-base_testing_net"
SERVICE_NAME="server:12345"  
MESSAGE="Test"
TIMEOUT=5

RESPONSE=$(echo "$MESSAGE" | nc -w $TIMEOUT $SERVICE_NAME)

if [ $? -eq 0 ] && [ "$RESPONSE" = "$MESSAGE" ]; then
  echo 'action: test_echo_server | result: success'
else
  echo 'action: test_echo_server | result: fail'
fi