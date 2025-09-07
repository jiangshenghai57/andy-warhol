curl http://localhost:8080/loans \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '[
        {
            "id": "LOAN001",
            "wac": 3.5,
            "wam": 360,
            "face": 100000,
            "prepay_cpr": 0.06,
            "static_dq": false,
            "performing_transition": [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01]
        }
    ]' \
    --verbose
