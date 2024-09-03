import socket
import logging
import signal
import threading
import sys
from .lottery import recv_batches,recv,recv_intro_msg,handle_winner_request
from .constants import MESSAGE_TYPE_BETDATA, MESSAGE_TYPE_REQWIN,NUM_AGENCIES
from concurrent.futures import ThreadPoolExecutor


class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_flag = threading.Event()
        self._handled_agencies = 0
        self._handled_agencies_lock = threading.Lock()
        self.executor = ThreadPoolExecutor(max_workers=20) 
        signal.signal(signal.SIGTERM, self.__handle_shutdown_signal)
        signal.signal(signal.SIGINT, self.__handle_shutdown_signal)
        
    def increment_handled_agencies(self):
        with self._handled_agencies_lock:
            self._handled_agencies += 1

    def get_handled_agencies(self):
        with self._handled_agencies_lock:
            return self._handled_agencies

    def __handle_shutdown_signal(self, signum, frame):
        signal_name = 'SIGTERM' if signum == signal.SIGTERM else 'SIGINT'
        logging.info(f'action: receive_signal | signal: {signal_name} | result: in_progress')
        self._shutdown_flag.set()  
        self._server_socket.close()  
        logging.info('action: shutdown_server | result: success')
        sys.exit(0)  

    def run(self):
        while not self._shutdown_flag.is_set():
            try:
                client_sock = self.__accept_new_connection()
                self.executor.submit(self.__handle_client_connection, client_sock)
            except OSError:
                break

    def __handle_client_connection(self, client_sock):
        try:
            addr = client_sock.getpeername()
            while True:
                msg_type = recv_intro_msg(client_sock)
                if msg_type == MESSAGE_TYPE_BETDATA:
                    recv_batches(client_sock)
                    self.increment_handled_agencies()
                    if self.get_handled_agencies() == NUM_AGENCIES:
                        logging.info(f'action: sorteo | result: success')
                elif msg_type == MESSAGE_TYPE_REQWIN:
                    if handle_winner_request(client_sock, self.get_handled_agencies()):
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

