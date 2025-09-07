# andy-warhol API

A mortgage loan amortization calculation service.

## Endpoints

### GET /info
Returns service information, available endpoints, and capabilities.

**Response:**
```json
{
  "service": "andy-warhol",
  "description": "Mortgage loan amortization calculation service",
  "version": "1.0.0",
  "endpoints": {
    "GET /info": "Get service information and capabilities",
    "GET /loans": "Retrieve list of processed loans",
    "POST /loans": "Submit loan data for amortization calculation"
  },
  "capabilities": [...],
  "loan_parameters": {...}
}
```

### GET /loans
Retrieves the list of loans that have been processed.

### POST /loans
Submits loan data for amortization calculation. Accepts an array of loan objects.

**Request Body:**
```json
[
  {
    "id": "LOAN001",
    "wam": 360,
    "wac": 4.5,
    "face": 250000,
    "prepay_cpr": 0.06
  }
]
```

## Running the Service

1. Build: `go build -o andy-warhol`
2. Run: `./andy-warhol`
3. Access: `http://localhost:8080`