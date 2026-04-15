#!/bin/bash
time curl -v http://127.0.0.1:12345/v1/messages \
    -H "Content-Type: application/json" \
    -d '{
        "stream": false,
        "model": "coding",
        "messages": [{"role": "user", "content": "Say hello"}]
    }'
