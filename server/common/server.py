import socket
import logging
import signal
import sys
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
from .lottery import recv_batches, recv_intro_msg, handle_winner_request
from .constants import MESSAGE_TYPE_BETDATA, MESSAGE_TYPE_REQWIN, NUM_AGENCIES

class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_flag = threading.Event()
        self._handled_agencies = 0
        self._clients_socks = []
        
        # Lock to synchronize access to the handled agencies counter
        self._handled_agencies_lock = threading.Lock()
        
        # Lock to synchronize access to storing bets
        self._store_bets_lock = threading.Lock()
                
        # Lock to synchronize access to the winners array
        self._winners_array_lock = threading.Lock()       
        self._winners = []  # List of tuples (winner_document: str, agency_id: str or int)
        
        # Thread pool executor to handle concurrent tasks
        self._executor = ThreadPoolExecutor(max_workers=NUM_AGENCIES)
        
        # Register signal handlers for graceful shutdown
        signal.signal(signal.SIGTERM, self.__handle_shutdown_signal)
        signal.signal(signal.SIGINT, self.__handle_shutdown_signal)
        
    def increment_handled_agencies(self):
        """
        Increments the count of handled agencies in a thread-safe manner.
        """
        with self._handled_agencies_lock:
            self._handled_agencies += 1

    def get_handled_agencies(self):
        """
        Returns the current count of handled agencies in a thread-safe manner.
        """
        with self._handled_agencies_lock:
            return self._handled_agencies

    def __handle_shutdown_signal(self, signum, frame):
        """
        Method to handle shutdown signals (SIGTERM or SIGINT).
        This method is triggered when the program receives a termination or interrupt signal.
        """
        signal_name = 'SIGTERM' if signum == signal.SIGTERM else 'SIGINT'
        logging.info(f'action: receive_signal | result: success | signal: {signal_name}')
        self._shutdown_flag.set()  
        for client_sock in self._clients_socks:
            try:
                client_sock.close()
                logging.info('action: disconnect_client | result: success')
            except Exception as e:
                logging.error(f'Failed to close client socket: {e}')
        
        logging.info('action: shutdown | result: success')
        self._server_socket.close()  

    def run(self):
        """
        Main server loop, handles accepting new connections and processing client requests.
        """
        while not self._shutdown_flag.is_set():
            try:
                client_sock = self.__accept_new_connection()
                self._clients_socks.append(client_sock)
                # Submit a task to the thread pool executor to handle the client connection
                self._executor.submit(self.__handle_client_connection, client_sock)

            except OSError:
                if self._shutdown_flag.is_set():
                    break  
                logging.error('Server accept failed due to an unexpected error')
                
        # Shutdown the thread pool executor and wait for all currently running tasks to complete
        self._executor.shutdown(wait=True)

    def __handle_client_connection(self, client_sock):
        """
        Handles communication with a client connection.
        """
        try:
            addr = client_sock.getpeername()
            while not self._shutdown_flag.is_set():  
                
                # Receive the introduction message to determine the type of request
                msg_type = recv_intro_msg(client_sock)
                
                if msg_type == MESSAGE_TYPE_BETDATA:
                    # Process bet data received from the client
                    recv_batches(client_sock, self._store_bets_lock)
                    self.increment_handled_agencies()
                    
                    # Check if the number of handled agencies matches the expected number
                    if self.get_handled_agencies() == NUM_AGENCIES:
                        logging.info(f'action: sorteo | result: success')
                        
                elif msg_type == MESSAGE_TYPE_REQWIN:
                # Handle a winner request and break the loop if successful
                    if handle_winner_request(client_sock, self.get_handled_agencies(), self):
                        break
                else:
                    logging.info(f'Unknown message received')
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            logging.info(f'action: close_client_socket | result: success | ip: {addr[0]}')

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
