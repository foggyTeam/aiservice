#!/bin/bash

# Curl requests for AI Service Analyze Handler
# Base URL - adjust as needed
BASE_URL="http://localhost:8080"

# ============================================
# 1. TEXT INPUT - Simple Request
# ============================================
echo "=== Test 1: Text Input ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "text",
    "input": {
      "type": "text",
      "text": "Create a timeline from 2020 to 2025"
    }
  }' | jq

exit 0
# ============================================
# 2. TEXT INPUT - With Context
# ============================================
echo -e "\n=== Test 2: Text Input with Context ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "text",
    "input": {
      "type": "text",
      "text": "Draw a flowchart for user authentication"
    },
    "context": {
      "board_theme": "whiteboard",
      "existing_elements": 5,
      "user_language": "en"
    }
  }' | jq .

# ============================================
# 3. INK INPUT - Simple Handwriting
# ============================================
echo -e "\n=== Test 3: Ink Input (Handwriting) ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "ink",
    "input": {
      "type": "ink",
      "strokes": [
        [
          {"x": 10.5, "y": 20.3, "t": 1700000000, "pressure": 0.8},
          {"x": 11.2, "y": 21.1, "t": 1700000010, "pressure": 0.85},
          {"x": 12.0, "y": 22.0, "t": 1700000020, "pressure": 0.9}
        ],
        [
          {"x": 50.0, "y": 30.0, "t": 1700000050, "pressure": 0.75},
          {"x": 51.5, "y": 31.2, "t": 1700000060, "pressure": 0.8}
        ]
      ],
      "meta": {
        "pen_type": "stylus",
        "resolution": 300,
        "content_type": "text"
      }
    }
  }' | jq .

# ============================================
# 4. INK INPUT - Complex Shape/Diagram
# ============================================
echo -e "\n=== Test 4: Ink Input (Diagram/Shape) ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "ink",
    "input": {
      "type": "ink",
      "strokes": [
        [
          {"x": 0, "y": 0, "t": 1700000000, "pressure": 0.8},
          {"x": 100, "y": 0, "t": 1700000010, "pressure": 0.8},
          {"x": 100, "y": 100, "t": 1700000020, "pressure": 0.8},
          {"x": 0, "y": 100, "t": 1700000030, "pressure": 0.8},
          {"x": 0, "y": 0, "t": 1700000040, "pressure": 0.8}
        ]
      ],
      "meta": {
        "content_type": "shape",
        "shape_type": "rectangle",
        "element_type": "diagram"
      }
    }
  }' | jq .

# ============================================
# 5. IMAGE INPUT - With URL
# ============================================
echo -e "\n=== Test 5: Image Input (URL) ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "image",
    "input": {
      "type": "image",
      "image_url": "https://example.com/diagram.png",
      "meta": {
        "format": "png",
        "width": 800,
        "height": 600
      }
    }
  }' | jq .

# ============================================
# 6. IMAGE INPUT - With Base64
# ============================================
echo -e "\n=== Test 6: Image Input (Base64) ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "image",
    "input": {
      "type": "image",
      "base64": "iVBORw0KGgoAAAANSUhEUgAAAAUA...",
      "meta": {
        "format": "png"
      }
    }
  }' | jq .

# ============================================
# 7. TEXT INPUT - With Callback (Async)
# ============================================
echo -e "\n=== Test 7: Async Processing with Callback ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "text",
    "input": {
      "type": "text",
      "text": "Generate a complex data visualization"
    },
    "callback_url": "https://miro.example.com/webhook/analyze",
    "request_id": "req_12345"
  }' | jq .

# ============================================
# 8. INK INPUT - Full Example with All Parameters
# ============================================
echo -e "\n=== Test 8: Full INK Input with All Parameters ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_miro_001",
    "user_id": "user_john_doe",
    "type": "ink",
    "input": {
      "type": "ink",
      "strokes": [
        [
          {"x": 10.0, "y": 20.0, "t": 1702080000000, "pressure": 0.7, "tilt": 45.0},
          {"x": 15.0, "y": 25.0, "t": 1702080010000, "pressure": 0.75, "tilt": 45.5},
          {"x": 20.0, "y": 30.0, "t": 1702080020000, "pressure": 0.8, "tilt": 46.0},
          {"x": 25.0, "y": 35.0, "t": 1702080030000, "pressure": 0.85, "tilt": 46.5}
        ]
      ],
      "meta": {
        "pen_type": "apple_pencil",
        "resolution": 300,
        "content_type": "text",
        "language": "en",
        "board_width": 1920,
        "board_height": 1080
      }
    },
    "context": {
      "board_theme": "dark",
      "existing_elements": ["shape_1", "text_2", "image_3"],
      "user_permissions": ["edit", "create", "delete"],
      "board_language": "en",
      "collaboration_mode": true
    },
    "callback_url": "https://your-miro-instance.com/api/webhooks/analyze",
    "request_id": "req_abc_123"
  }' | jq .

# ============================================
# 9. Get Job Status
# ============================================
echo -e "\n=== Test 9: Get Job Status ==="
curl -X GET "$BASE_URL/jobs/job_1702080000123456789" \
  -H "Content-Type: application/json" | jq .

# ============================================
# 10. Health Check
# ============================================
echo -e "\n=== Test 10: Health Check ==="
curl -X GET "$BASE_URL/health" | jq .

# ============================================
# 11. TEXT INPUT - Create Sticky Note
# ============================================
echo -e "\n=== Test 11: Create Sticky Note ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_sticky",
    "user_id": "user_alice",
    "type": "text",
    "input": {
      "type": "text",
      "text": "Important: Complete project proposal by Friday"
    },
    "context": {
      "action_type": "sticky_note",
      "color": "yellow",
      "position": {"x": 100, "y": 200}
    }
  }' | jq .

# ============================================
# 12. INK INPUT - Mathematical Formula
# ============================================
echo -e "\n=== Test 12: Mathematical Formula Recognition ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_math",
    "user_id": "user_scientist",
    "type": "ink",
    "input": {
      "type": "ink",
      "strokes": [
        [
          {"x": 0, "y": 10, "t": 1700000000},
          {"x": 10, "y": 10, "t": 1700000010},
          {"x": 20, "y": 10, "t": 1700000020}
        ]
      ],
      "meta": {
        "content_type": "formula",
        "expected_format": "latex"
      }
    }
  }' | jq .

# ============================================
# 13. ERROR TEST - Missing Required Fields
# ============================================
echo -e "\n=== Test 13: Error - Missing Required Fields ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123"
  }' | jq .

# ============================================
# 14. ERROR TEST - Invalid JSON
# ============================================
echo -e "\n=== Test 14: Error - Invalid JSON ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{invalid json}' | jq .

# ============================================
# 15. BULK REQUEST - Save to File
# ============================================
echo -e "\n=== Test 15: Save Response to File ==="
curl -X POST "$BASE_URL/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_123",
    "user_id": "user_456",
    "type": "text",
    "input": {
      "type": "text",
      "text": "Create a diagram"
    }
  }' | jq . > response.json && echo "Response saved to response.json"
