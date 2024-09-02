cd .data

get_max_length() {
    field_index=$1
    max_length=0

    unzip -p dataset.zip "agency-*.csv" | awk -F, -v col="$field_index" '
    {
        if (length($col) > max) {
            max = length($col)
        }
    }
    END { print max }'
}

MAX_NOMBRE_LEN=$(get_max_length 1)
MAX_APELLIDO_LEN=$(get_max_length 2)
MAX_DOCUMENTO_LEN=$(get_max_length 3)
MAX_NACIMIENTO_LEN=$(get_max_length 4)
MAX_NUMERO_LEN=$(get_max_length 5)
MAX_AGENCY_LEN=1

MAX_LINE_SIZE=$((MAX_NOMBRE_LEN + MAX_APELLIDO_LEN + MAX_DOCUMENTO_LEN + MAX_NACIMIENTO_LEN + MAX_NUMERO_LEN + MAX_AGENCY_LEN + 5 + 1))
MAX_LINES_IN_8KB=$((8192 / MAX_LINE_SIZE))

echo "$MAX_LINES_IN_8KB"