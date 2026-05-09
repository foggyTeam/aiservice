go test \
  -coverprofile=coverage.out \
  ./cmd/server/... \
  ./internal/config/... \
  ./internal/digitalink/... \
  ./internal/handlers/... \
  ./internal/log/... \
  ./internal/middleware/... \
  ./internal/models/... \
  ./internal/preprocessing/... \
  ./internal/providers/ \
  ./internal/providers/ \
  ./internal/utils/... \
  ./internal/services/analysis/... \
  ./internal/services/database/... \
  ./internal/services/graph/... \
  ./internal/services/image/... \
  ./internal/services/jobService/... \
  ./internal/services/pipeline/... \
  ./internal/services/storage/... 

