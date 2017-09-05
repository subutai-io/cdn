#!/bin/bash

while [[ $# -gt 1 ]]; do

case $1 in
    -e|--environment)
    	ENV="$2"
    	shift
    ;;
    -a|--address)
    	addr="$2"
    	shift
    ;;
    *)
	echo "Unknown option $2"
	exit 1
    ;;
esac
shift 
done

for type in template raw; do
	echo $type
	curl -s -k https://${ENV}cdn.subut.ai:8338/kurjun/rest/$type/info | jq '.[] | .id' | tr -d '"'|
	while IFS= read -r ID 
	do
     		echo "https://$addr:8338/kurjun/rest/$type/get?id=$ID" 
     		curl -m 3 -k "https://$addr:8338/kurjun/rest/$type/get?id=$ID" -o /dev/null
	done
done
