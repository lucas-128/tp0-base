import socket
import struct
import logging
from typing import List, Optional
from .utils import Bet, store_bets

def recv_all(sock, length):
    data = bytearray()
    while len(data) < length:
        packet = sock.recv(length - len(data))
        data.extend(packet)
    return data

def recv_batches(client_sock):
    
    while True:
        size_data = recv_all(client_sock, 4)
        data_size = int.from_bytes(size_data, byteorder='big')
    
        if not data_size:
            break
        
        bet_data = recv_all(client_sock, data_size).decode('utf-8')
        parsed_bets = parse_bets(bet_data)
        
        if parsed_bets is None:
            msg = f'action: apuesta_recibida | result: fail | cantidad: {len(bet_data.strip().splitlines())}'
            logging.error(msg)
            client_sock.send((msg + '\n').encode('utf-8'))
        else:     
            store_bets(parsed_bets)
            msg = f'action: apuesta_recibida  | result: success | cantidad: {len(parsed_bets)}'
            logging.info(msg)
            client_sock.send((msg + '\n').encode('utf-8'))
        
def parse_bets(bet_data: str) -> Optional[List[Bet]]:
    bets = []
    bet_entries = bet_data.strip().split('\n')
    
    for bet_entry in bet_entries:
        try:
            bet_parts = bet_entry.split(',')
            
            if len(bet_parts) != 6:  
                return None 

            bet = Bet(
                first_name=bet_parts[0],
                last_name=bet_parts[1],
                document=bet_parts[2],
                birthdate=bet_parts[3],
                number=bet_parts[4],
                agency=bet_parts[5],
            )
            bets.append(bet)
        except (IndexError, ValueError):
            return None 
    
    return bets
         
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


