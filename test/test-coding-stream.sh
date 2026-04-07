#!/bin/bash
time curl -s http://127.0.0.1:12345/v1/messages \
    -H "Content-Type: application/json" \
    -d '{
        "stream": true,
        "model": "coding",
        "messages": [{"role": "user", "content": "Say hello"}]
    }'
