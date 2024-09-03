if [ "$#" -ne 2 ]; then
    echo "Uso: $0 <output_file_name> <number_of_clients>"
    exit 1
fi

output_file=$1
num_clients=$2

python3 mi-generador.py "$output_file" "$num_clients"
#echo "Docker Compose file '$output_file' generado con $num_clients clientes."
