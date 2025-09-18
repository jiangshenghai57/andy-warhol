#!/bin/sh

# Create JSON array with mortgage objects
data="["
for i in {1..1}
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

  prepay_cpr=$(echo "scale=4; 0.02 + ($i % 5) * 0.01" | bc)

  if [[ $prepay_cpr =~ ^\0. ]]; then
    prepay_cpr="0$prepay_cpr"
  fi

  # FIX: Put entire JSON object on one line
  data="$data{\"id\": \"$i\", \"wac\": $wac, \"wam\": $wam, \"face\": $face, \"static_dq\": $staticdq_str, \"prepay_cpr\": $prepay_cpr}"

  if [ $i -lt 1 ]; then  # Fixed: should be 1, not 1000 since loop is 1..1
    data="$data"
  fi
done
data="$data]"

echo "$data" | python3 -c "import json,sys; json.load(sys.stdin); print('✅ JSON is valid')" 2>/dev/null || echo "❌ JSON is invalid"


echo "Data to be sent: $data"

# Add verbose flag to see the actual error
curl http://localhost:8080/loans \
    --include \
    --header "Content-Type: application/json" \
    --data "$data" \
    --verbose