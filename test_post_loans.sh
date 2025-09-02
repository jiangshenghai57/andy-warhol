#!/bin/sh

# Create JSON array with mortgage objects
data=""
for i in {1..10}
do
  wac=$(echo "scale=2; 3.0 + ($i % 50) * 0.1" | bc)
  wam=$((60 + ($i % 3) * 90))
  face=$((100000 + ($i * 5000)))
  staticdq=$((i % 2 == 0))
  
  if [ $staticdq -eq 1 ]; then
    staticdq_str="true"
  else
    staticdq_str="false"
  fi
  
  data="$data{\"id\": \"$i\", \"wac\": $wac, \"wam\": $wam, \"face\": $face, \"staticdq\": $staticdq_str}"
  
  if [ $i -lt 1000 ]; then  # Changed to match loop count
    data="$data,"
  fi
done
data="$data"  # Close the JSON array

curl http://localhost:8080/loans \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data "$data"