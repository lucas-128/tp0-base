import socket
import logging
import signal
import threading
import sys
from .lottery import recv_batches,recv

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_flag = threading.Event()
        self._clients_socks = []
        signal.signal(signal.SIGTERM, self.__handle_shutdown_signal)
        signal.signal(signal.SIGINT, self.__handle_shutdown_signal)

    def __handle_shutdown_signal(self, signum, frame):
        """
        Method to handle shutdown signals (SIGTERM or SIGINT).
        This method is triggered when the program receives a termination or interrupt signal.
        """
        signal_name = 'SIGTERM' if signum == signal.SIGTERM else 'SIGINT'
        logging.info(f'action: receive_signal | result: success | signal: {signal_name} ')
        self._shutdown_flag.set()  
        for client_sock in self._clients_socks:
            client_sock.close() 
            logging.info('action: disconnect_client | result: success')
        logging.info('action: shutdown | result: success')
        sys.exit(0)
          

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communication
        finishes, servers starts to accept new connections again
        """

        while not self._shutdown_flag.is_set():
            try:
                client_sock = self.__accept_new_connection()
                self._clients_socks.append(client_sock)
                self.__handle_client_connection(client_sock)
            except OSError:
                break

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            addr = client_sock.getpeername()
            recv_batches(client_sock)
                        
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            logging.info(f'action: close_client_socket | result: success | ip: {addr[0]}')

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
