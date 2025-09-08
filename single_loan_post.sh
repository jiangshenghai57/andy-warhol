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
            "performing_transition": [0.93, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
            "dq30_transition": [0.90, 0.04, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
            "dq60_transition": [0.88, 0.05, 0.02, 0.01, 0.01, 0.01, 0.01, 0.01],
            "dq90_transition": [0.01, 0.01, 0.01, 0.93, 0.01, 0.01, 0.01, 0.01],
            "dq120_transition": [0.01, 0.01, 0.01, 0.01, 0.93, 0.01, 0.01, 0.01],
            "dq150_transition": [0.01, 0.01, 0.01, 0.01, 0.01, 0.93, 0.01, 0.01],
            "dq180_transition": [0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.93, 0.01],
            "default_transition": [0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.93]
        }
    ]' \
    --verbose
