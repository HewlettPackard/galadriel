#!/bin/bash

url="https://fulcio.sigstore.dev/api/v2/trustBundle"
response=$(curl -s "$url")
certs=$(echo "$response" | jq -r '.chains[].certificates[]')

count=1
buffer=""

IFS=$'\n'
for line in $certs; do
  if [[ $line == "-----BEGIN CERTIFICATE-----" ]]; then
    if [ ! -z "$buffer" ]; then
      echo -e "$buffer" > "certificate_$count.pem"
      count=$((count+1))
      buffer=""
    fi
  fi
  buffer="$buffer\n$line"
done

if [ ! -z "$buffer" ]; then
  echo -e "$buffer" > "certificate_$count.pem"
fi

if [ -f "certificate_1.pem" ]; then
  echo "Assuming certificate_1.pem is the root certificate."
  mv "certificate_1.pem" "root.pem"
else
  echo "Error: certificate_1.pem not found."
fi

if [ -f "certificate_2.pem" ]; then
  echo "Assuming certificate_2.pem is the intermediate certificate."
  mv "certificate_2.pem" "intermediate.pem"
else
  echo "Error: certificate_2.pem not found."
fi
