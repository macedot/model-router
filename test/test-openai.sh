#!/bin/bash
time curl -v http://127.0.0.1:12345/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d '{
        "model": "openai",
        "messages": [{"role": "user", "content": "Say hello"}]
    }'
