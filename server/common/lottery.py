import socket
import struct
import logging
from .utils import Bet, store_bets

def recv_all(sock, length):
    data = bytearray()
    while len(data) < length:
        packet = sock.recv(length - len(data))
        data.extend(packet)
    return data

def recv(client_sock):
    size_data = recv_all(client_sock, 4)
    data_size = int.from_bytes(size_data, byteorder='big')

    bet_data = recv_all(client_sock, data_size).decode('utf-8')
    
    bet_parts = bet_data.split('|')
    
    bet = Bet(
        agency=bet_parts[0],
        first_name=bet_parts[1],
        last_name=bet_parts[2],
        document=bet_parts[3],
        birthdate=bet_parts[4],
        number=bet_parts[5]
    )
    
    store_bets([bet])

    msg = f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}'
    logging.info(msg)
    
    client_sock.send((msg + '\n').encode('utf-8'))
