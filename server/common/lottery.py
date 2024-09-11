import socket
import struct
import logging
from typing import List, Optional
from .utils import Bet, store_bets,load_bets,has_won
from .constants import LENGTH_BYTES, MESSAGE_TYPE_NOWINN, MESSAGE_TYPE_WINNERS,NUM_AGENCIES

# Receives all bytes of data from a socket until the specified length is met
def recv_all(sock, length):
    data = bytearray()
    while len(data) < length:
        packet = sock.recv(length - len(data))
        data.extend(packet)
    return data

# Receives an introductory message from the client socket
def recv_intro_msg(client_sock):
    size_data = recv_all(client_sock, LENGTH_BYTES)
    data_size = int.from_bytes(size_data, byteorder='big')
    message_data = recv_all(client_sock, data_size)    
    message = message_data.decode('utf-8')
    return message
  
# Sends a message to the client socket, including its length
def send_message_len(client_sock, msg: str):
    msg_bytes = msg.encode('utf-8')    
    size = len(msg_bytes)
    size_bytes = size.to_bytes(LENGTH_BYTES, byteorder='big')
    send_all(client_sock, size_bytes)
    send_all(client_sock, msg_bytes)  

# Handles a winner request from the client socket and sends appropriate responses
def handle_winner_request(client_sock, handled_agencies, server):
    
    size_bytes = recv_all(client_sock, LENGTH_BYTES)
    size = int.from_bytes(size_bytes, byteorder='big')   
    agency_id_bytes = recv_all(client_sock, size)
    agency_id = agency_id_bytes.decode('utf-8')
    
    if handled_agencies < NUM_AGENCIES:
        msg = MESSAGE_TYPE_NOWINN
        send_message_len(client_sock,msg)
        return False  
    else:
        msg = MESSAGE_TYPE_WINNERS
        send_message_len(client_sock,msg)
        winners_docs = get_winners(agency_id,server)
        send_message_len(client_sock, winners_docs)
        return True
  
# Retrieves the list of winners for a given agency ID
def get_winners(agency_id, server):
    with server._winners_array_lock:
        if not server._winners:
            bets = load_bets()
            for bet in bets:
                if has_won(bet):
                    server._winners.append((bet.document, bet.agency))

        winners_for_agency = [doc for doc, agency in server._winners if agency == int(agency_id)]
        result = ','.join(winners_for_agency)
        return result 
        
# Sends all data to the given connection                   
def send_all(conn: socket.socket, data: bytes):
    total_sent = 0
    while total_sent < len(data):
        sent = conn.send(data[total_sent:])
        total_sent += sent
        
# Receives data in batches from the client socket, processes and stores it
def recv_batches(client_sock, store_bets_lock):
    while True:
        size_data = recv_all(client_sock, LENGTH_BYTES)
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
            with store_bets_lock:
                store_bets(parsed_bets)    
            msg = f'action: apuesta_recibida  | result: success | cantidad: {len(parsed_bets)}'
            logging.info(msg)
            client_sock.send((msg + '\n').encode('utf-8'))

# Parses bet data from a string into a list of Bet objects       
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

# Receives a single bet from the client socket and stores it                 
def recv(client_sock):
    
    size_data = recv_all(client_sock, LENGTH_BYTES)
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


