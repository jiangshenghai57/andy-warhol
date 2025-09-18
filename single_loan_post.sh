#!/bin/bash

# Script to generate and post 1000 loans
# filepath: /Users/shenghaijiang/repos/andy-warhol/single_loan_post.sh

# Start JSON array
json_data="["

# Generate 1000 loans with slightly different parameters
for i in $(seq 1 1)
do
    # Create unique loan ID with padding
    loan_id=$(printf "LOAN%04d" $i)
    
    # Vary parameters slightly to make data more realistic
    wac=$(echo "scale=2; 3.5 + ($i % 30) * 0.1" | bc)
    wam=$((360 - ($i % 12) * 10))
    face=$((100000 + ($i % 50) * 5000))
    prepay_cpr=$(echo "scale=4; 0.05 + ($i % 10) * 0.005" | bc)

    # Add comma between loans (except for the last one)
    if [ $i -gt 1 ]; then
        json_data="$json_data,"
    fi
    
    # Add loan object to JSON array
    json_data="$json_data{
        \"id\": \"$loan_id\",
        \"wac\": $wac,
        \"wam\": $wam,
        \"face\": $face,
        \"prepay_cpr\": 0$prepay_cpr,
        \"static_dq\": false,
        \"performing_transition\": [0.93, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
        \"dq30_transition\": [0.90, 0.04, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
        \"dq60_transition\": [0.88, 0.05, 0.02, 0.01, 0.01, 0.01, 0.01, 0.01],
        \"dq90_transition\": [0.01, 0.01, 0.01, 0.93, 0.01, 0.01, 0.01, 0.01],
        \"dq120_transition\": [0.01, 0.01, 0.01, 0.01, 0.93, 0.01, 0.01, 0.01],
        \"dq150_transition\": [0.01, 0.01, 0.01, 0.01, 0.01, 0.93, 0.01, 0.01],
        \"dq180_transition\": [0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.93, 0.01],
        \"default_transition\": [0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.93]
    }"
done

# Close JSON array
json_data="$json_data]"

# Save JSON to file (useful for large payloads)
echo "$json_data" > loans_payload.json

# Print stats about the generated data
echo "Generated 1000 loans with varying parameters:"
echo "- WAC range: 3.5% to 6.4%"
echo "- WAM range: 240 to 360 months"
echo "- Face values: $100,000 to $345,000"
echo "- CPR range: 5% to 9.5%"
echo "JSON size: $(echo "$json_data" | wc -c) bytes"

# Validate JSON
echo "Validating JSON..."
echo "$json_data" | python3 -c "import json,sys; json.load(sys.stdin); print('✅ JSON is valid')" 2>/dev/null || (echo "❌ JSON is invalid"; exit 1)

# Send request with output handling
echo "Sending request to http://localhost:8080/loans..."
curl -s -X POST http://localhost:8080/loans \
    -H "Content-Type: application/json" \
    --data @loans_payload.json \
    -o response.json

echo "Response saved to response.json"
echo "Request complete!"

# Optional: Show response summary
echo "Response summary:"
cat response.json | python3 -c "
import json,sys
try:
    data = json.load(sys.stdin)
    print(f\"Count: {data.get('count', 'N/A')}\")
    print(f\"Status: {data.get('message', 'N/A')}\")
except Exception as e:
    print(f\"Error parsing response: {e}\")
