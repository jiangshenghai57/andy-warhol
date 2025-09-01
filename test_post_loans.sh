#!/bin/sh

# Create JSON array with 100 mortgage objects
data=""
for i in {1..1000}
do
  wac=$(echo "scale=2; 3.0 + ($i % 50) * 0.1" | bc)
  wam=$((60 + ($i % 3) * 90))
  face=$((100000 + ($i * 5000)))  # Add face value starting at 100,000
  staticdq=$((i % 2 == 0))
  
  if [ $staticdq -eq 1 ]; then
    staticdq_str="true"
  else
    staticdq_str="false"
  fi
  
  data="$data{\"id\": \"$i\", \"wac\": $wac, \"wam\": $wam, \"face\": $face, \"staticdq\": $staticdq_str}"
  
  if [ $i -lt 100 ]; then
    data="$data,"
  fi
done
data="$data"

echo "Posting the following data:" $data

curl http://localhost:8080/loans \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data "$data"