import socket
import logging
import signal
import threading
import sys
from .lottery import recv_batches,recv,recv_intro_msg,handle_winner_request
from .constants import MESSAGE_TYPE_BETDATA, MESSAGE_TYPE_REQWIN,NUM_AGENCIES


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_flag = threading.Event()
        self._handled_agencies = 0
        
        

        # Signal handlers 
        signal.signal(signal.SIGTERM, self.__handle_shutdown_signal)
        signal.signal(signal.SIGINT, self.__handle_shutdown_signal)

    def __handle_shutdown_signal(self, signum, frame):
        signal_name = 'SIGTERM' if signum == signal.SIGTERM else 'SIGINT'
        logging.info(f'action: receive_signal | signal: {signal_name} | result: in_progress')
        self._shutdown_flag.set()  
        self._server_socket.close()  
        logging.info('action: shutdown_server | result: success')
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
            msg_type = recv_intro_msg(client_sock)
            if msg_type == MESSAGE_TYPE_BETDATA:
                recv_batches(client_sock)
                self._handled_agencies += 1
                if (self._handled_agencies == NUM_AGENCIES):
                    logging.info(f'action: sorteo | result: success')
            elif msg_type == MESSAGE_TYPE_REQWIN:
                handle_winner_request(client_sock,self._handled_agencies)
            else:
                logging.info(f'Unknown message received')
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

