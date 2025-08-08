#!/bin/bash

DIRECTORY="$1"

for file in "$DIRECTORY"/*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing: $file"
        curl --request POST \
             --url http://localhost:8080/v1/street-manager-relay/sns \
             --header 'content-type: application/json' \
             --header 'x-amz-sns-message-type: 2' \
             --data @"$file"
        echo

    else
        echo "No JSON files found in $DIRECTORY"
        break
    fi
done