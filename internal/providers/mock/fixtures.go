package mock

import "github.com/aiservice/internal/providers"

var summarizeFixtures = map[MockMode][]providers.SummarizeFlow{
	ModeLight: {
		{Summarization: "Quick notes about backend sync."},
		{Summarization: "Short board summary with API ideas."},
		{Summarization: "Simple architecture discussion."},
		{Summarization: "Lightweight websocket planning notes."},
		{Summarization: "Frontend TODO list and UI ideas."},
		{Summarization: "Basic deployment checklist."},
		{Summarization: "Authentication flow draft."},
		{Summarization: "Canvas rendering experiments."},
		{Summarization: "Small discussion about caching."},
		{Summarization: "Notes about socket namespaces."},
	},

	ModeDefault: {
		{
			Summarization: "The board describes websocket synchronization architecture with namespace separation and state persistence.",
		},
		{
			Summarization: "Discussion about optimistic updates, rollback handling, and client synchronization.",
		},
		{
			Summarization: "The board contains API routes, middleware planning, and authorization strategy.",
		},
		{
			Summarization: "Frontend rendering pipeline for collaborative canvas interactions.",
		},
		{
			Summarization: "Database schema planning for board snapshots and event history.",
		},
		{
			Summarization: "Microservice communication diagram with retry and reconciliation logic.",
		},
		{
			Summarization: "Authentication architecture using JWT access and refresh tokens.",
		},
		{
			Summarization: "Realtime cursor synchronization and websocket lifecycle management.",
		},
		{
			Summarization: "Load balancing and horizontal scaling discussion for sync workers.",
		},
		{
			Summarization: "Notes about distributed event processing and conflict resolution.",
		},
	},

	ModeHeavy: {
		{
			Summarization: "Comprehensive summary of a distributed synchronization system including websocket multiplexing, event sourcing, snapshot persistence, optimistic state reconciliation, conflict resolution, horizontal scaling, and collaborative rendering architecture.",
		},
		{
			Summarization: "Detailed overview of a collaborative whiteboard backend with socket namespaces, Redis pub/sub, persistent event streams, recovery snapshots, and consistency guarantees.",
		},
		{
			Summarization: "The board documents a realtime infrastructure strategy with state replication, synchronization workers, retry queues, event deduplication, and distributed tracing.",
		},
		{
			Summarization: "Large architecture discussion about scalable canvas rendering, incremental synchronization, throttling strategies, and websocket backpressure handling.",
		},
		{
			Summarization: "Extensive planning notes covering API gateways, internal RPC communication, storage abstractions, observability, and deployment topology.",
		},
		{
			Summarization: "Complex synchronization pipeline with client prediction, reconciliation phases, CRDT experimentation, and distributed locking considerations.",
		},
		{
			Summarization: "Comprehensive engineering notes about event-driven architecture, queue durability, failure recovery, and transactional consistency.",
		},
		{
			Summarization: "Large-scale collaborative editing design including history persistence, operational transforms, and synchronization recovery logic.",
		},
		{
			Summarization: "Advanced infrastructure planning for high-load websocket coordination and distributed board state persistence.",
		},
		{
			Summarization: "Full system architecture overview involving frontend collaboration engine, backend synchronization workers, Redis coordination, and database snapshotting.",
		},
	},
}

var structurizeFixtures = map[MockMode][]providers.StructurizeFlow{
	ModeLight: {
		{AiTreeResponse: "root\n└── notes"},
		{AiTreeResponse: "project\n└── api"},
		{AiTreeResponse: "workspace\n├── todo\n└── ideas"},
		{AiTreeResponse: "docs\n└── draft"},
		{AiTreeResponse: "backend\n└── auth"},
		{AiTreeResponse: "frontend\n└── ui"},
		{AiTreeResponse: "sync\n└── websocket"},
		{AiTreeResponse: "board\n└── elements"},
		{AiTreeResponse: "service\n└── handlers"},
		{AiTreeResponse: "storage\n└── cache"},
	},

	ModeDefault: {
		{
			AiTreeResponse: `project
├── backend
│   ├── api
│   └── sync
└── frontend`,
		},
		{
			AiTreeResponse: `workspace
├── auth
├── websocket
└── storage`,
		},
		{
			AiTreeResponse: `system
├── handlers
├── middleware
└── database`,
		},
		{
			AiTreeResponse: `service
├── queue
├── events
└── snapshots`,
		},
		{
			AiTreeResponse: `application
├── ui
├── api
└── workers`,
		},
		{
			AiTreeResponse: `platform
├── frontend
├── backend
└── infrastructure`,
		},
		{
			AiTreeResponse: `sync-service
├── sockets
├── state
└── persistence`,
		},
		{
			AiTreeResponse: `api
├── auth
├── routes
└── validation`,
		},
		{
			AiTreeResponse: `board-system
├── renderer
├── sync
└── history`,
		},
		{
			AiTreeResponse: `cluster
├── gateway
├── workers
└── redis`,
		},
	},

	ModeHeavy: {
		{
			AiTreeResponse: `enterprise-platform
├── gateway
│   ├── auth
│   │   ├── jwt
│   │   ├── refresh
│   │   └── permissions
│   ├── routing
│   └── rate-limit
├── collaboration
│   ├── websocket
│   │   ├── cursors
│   │   ├── presence
│   │   └── rooms
│   ├── synchronization
│   │   ├── snapshots
│   │   ├── reconciliation
│   │   └── persistence
│   └── history
├── llm
│   ├── summarize
│   ├── structurize
│   ├── templates
│   └── embeddings
└── infrastructure
    ├── redis
    ├── postgres
    ├── s3
    └── monitoring`,
		},
		{
			AiTreeResponse: `distributed-sync-system
├── frontend
│   ├── canvas
│   │   ├── renderer
│   │   ├── viewport
│   │   └── interactions
│   ├── state
│   └── websocket
├── backend
│   ├── handlers
│   ├── middleware
│   ├── queues
│   └── workers
│       ├── sync
│       ├── snapshots
│       └── recovery
├── persistence
│   ├── events
│   ├── boards
│   └── versions
└── observability
    ├── tracing
    ├── metrics
    └── logs`,
		},
		{
			AiTreeResponse: `collaboration-engine
├── api
│   ├── public
│   ├── internal
│   └── admin
├── realtime
│   ├── sockets
│   ├── state-sync
│   ├── event-stream
│   └── presence
├── storage
│   ├── snapshots
│   ├── history
│   ├── exports
│   └── cache
├── ai
│   ├── summarization
│   ├── template-generation
│   ├── image-recognition
│   └── text-generation
└── deployment
    ├── docker
    ├── kubernetes
    └── ci-cd`,
		},
		{
			AiTreeResponse: `microservices
├── auth-service
│   ├── sessions
│   ├── oauth
│   └── permissions
├── sync-service
│   ├── websocket
│   ├── state-machine
│   ├── snapshots
│   └── persistence
├── llm-service
│   ├── prompts
│   ├── providers
│   ├── pipelines
│   └── parsers
├── storage-service
│   ├── postgres
│   ├── redis
│   └── object-storage
└── monitoring-service
    ├── tracing
    ├── alerting
    └── dashboards`,
		},
		{
			AiTreeResponse: `whiteboard-platform
├── frontend
│   ├── editor
│   ├── collaboration
│   ├── components
│   └── state
├── backend
│   ├── api
│   ├── websocket
│   ├── queues
│   └── processors
├── ai-pipelines
│   ├── summarize
│   ├── structurize
│   ├── templates
│   └── image-analysis
├── infrastructure
│   ├── ingress
│   ├── load-balancer
│   ├── redis-cluster
│   └── postgres-cluster
└── observability
    ├── logs
    ├── metrics
    └── tracing`,
		},
		{
			AiTreeResponse: `event-driven-platform
├── producers
│   ├── frontend-events
│   ├── sync-events
│   └── ai-events
├── queues
│   ├── persistence
│   ├── retries
│   └── dead-letter
├── consumers
│   ├── sync-workers
│   ├── recovery-workers
│   └── analytics-workers
├── storage
│   ├── snapshots
│   ├── events
│   └── archives
└── infrastructure
    ├── kafka
    ├── redis
    └── monitoring`,
		},
		{
			AiTreeResponse: `ai-platform
├── prompts
│   ├── summarize
│   ├── structurize
│   └── templates
├── providers
│   ├── openai
│   ├── anthropic
│   └── mock
├── pipelines
│   ├── preprocessing
│   ├── generation
│   └── postprocessing
├── parsers
│   ├── markdown
│   ├── html
│   └── json
└── infrastructure
    ├── cache
    ├── queues
    └── telemetry`,
		},
		{
			AiTreeResponse: `scalable-backend
├── ingress
├── api
│   ├── graphql
│   ├── rest
│   └── websocket
├── workers
│   ├── synchronization
│   ├── persistence
│   ├── exports
│   └── cleanup
├── storage
│   ├── postgres
│   ├── redis
│   └── object-storage
└── monitoring
    ├── prometheus
    ├── grafana
    └── tracing`,
		},
		{
			AiTreeResponse: `knowledge-platform
├── editor
│   ├── pages
│   ├── graphs
│   └── references
├── collaboration
│   ├── realtime
│   ├── comments
│   └── presence
├── ai
│   ├── summarization
│   ├── classification
│   └── embeddings
├── search
│   ├── indexing
│   ├── ranking
│   └── retrieval
└── infrastructure
    ├── databases
    ├── queues
    └── telemetry`,
		},
		{
			AiTreeResponse: `production-cluster
├── edge
│   ├── cdn
│   ├── gateway
│   └── routing
├── applications
│   ├── frontend
│   ├── api
│   ├── sync
│   └── ai
├── storage
│   ├── postgres
│   ├── redis
│   ├── backups
│   └── snapshots
└── operations
    ├── monitoring
    ├── alerting
    └── deployments`,
		},
	},
}

var templateFixtures = map[MockMode][]providers.TemplateGenerationFlow{
	ModeLight: {
		{
			BoardType:   "simple",
			Title:       "Meeting",
			Description: "Meeting notes template",
		},
		{
			BoardType:   "simple",
			Title:       "Auth Notes",
			Description: "Authentication planning",
		},
		{
			BoardType:   "simple",
			Title:       "Frontend",
			Description: "Frontend notes",
		},
		{
			BoardType:   "simple",
			Title:       "Backend",
			Description: "Backend planning",
		},
		{
			BoardType:   "simple",
			Title:       "Sync",
			Description: "Sync service notes",
		},
	},
	ModeDefault: {
		{
			BoardType:   "simple",
			Title:       "Backend Architecture",
			Description: "API and synchronization architecture overview",
			Elements: []providers.TemplateElement{
				{
					Type:         "rectangle",
					X:            100,
					Y:            100,
					Width:        240,
					Height:       100,
					Fill:         "#E3F2FD",
					Stroke:       "#1E88E5",
					StrokeWidth:  2,
					Content:      "API Gateway",
					CornerRadius: 12,
				},
				{
					Type:         "rectangle",
					X:            420,
					Y:            100,
					Width:        240,
					Height:       100,
					Fill:         "#E8F5E9",
					Stroke:       "#43A047",
					StrokeWidth:  2,
					Content:      "Sync Service",
					CornerRadius: 12,
				},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Realtime Collaboration",
			Description: "Realtime synchronization topology",
			GraphNodes: []providers.TemplateNode{
				{ID: "frontend", Type: "customNode", Title: "Frontend"},
				{ID: "gateway", Type: "customNode", Title: "Gateway"},
				{ID: "sync", Type: "customNode", Title: "Sync Worker"},
				{ID: "redis", Type: "customNode", Title: "Redis"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "frontend", Target: "gateway", Label: "ws"},
				{ID: "e2", Source: "gateway", Target: "sync", Label: "events"},
				{ID: "e3", Source: "sync", Target: "redis", Label: "pub/sub"},
			},
		},
	},
	ModeHeavy: {
		{
			BoardType:   "graph",
			Title:       "Realtime Knowledge Processing Platform",
			Description: "Distributed realtime processing and indexing infrastructure",
			GraphNodes: []providers.TemplateNode{
				{ID: "cdn", Type: "customNode", Title: "CDN Edge"},
				{ID: "gateway", Type: "customNode", Title: "Gateway"},
				{ID: "auth", Type: "customNode", Title: "Auth Cluster"},
				{ID: "api", Type: "customNode", Title: "API Workers"},
				{ID: "queue", Type: "customNode", Title: "Queue Bus"},
				{ID: "indexer", Type: "customNode", Title: "Indexer"},
				{ID: "search", Type: "customNode", Title: "Search Engine"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "cdn", Target: "gateway", Label: "proxy"},
				{ID: "e2", Source: "gateway", Target: "auth", Label: "validate"},
				{ID: "e3", Source: "auth", Target: "api", Label: "authorize"},
				{ID: "e4", Source: "api", Target: "queue", Label: "publish"},
				{ID: "e5", Source: "queue", Target: "indexer", Label: "consume"},
				{ID: "e6", Source: "indexer", Target: "search", Label: "update"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Distributed AI Rendering Pipeline",
			Description: "Large-scale rendering and AI orchestration system",
			GraphNodes: []providers.TemplateNode{
				{ID: "ingress", Type: "customNode", Title: "Ingress"},
				{ID: "router", Type: "customNode", Title: "Traffic Router"},
				{ID: "llm", Type: "customNode", Title: "LLM Cluster"},
				{ID: "renderer", Type: "customNode", Title: "Renderer"},
				{ID: "cache", Type: "customNode", Title: "Distributed Cache"},
				{ID: "storage", Type: "customNode", Title: "Artifact Storage"},
				{ID: "analytics", Type: "customNode", Title: "Analytics"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "ingress", Target: "router", Label: "traffic"},
				{ID: "e2", Source: "router", Target: "llm", Label: "prompts"},
				{ID: "e3", Source: "llm", Target: "renderer", Label: "instructions"},
				{ID: "e4", Source: "renderer", Target: "cache", Label: "cache"},
				{ID: "e5", Source: "cache", Target: "storage", Label: "persist"},
				{ID: "e6", Source: "storage", Target: "analytics", Label: "metrics"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Enterprise Event Streaming Platform",
			Description: "Scalable distributed event streaming architecture",
			GraphNodes: []providers.TemplateNode{
				{ID: "clients", Type: "customNode", Title: "Clients"},
				{ID: "gateway", Type: "customNode", Title: "Gateway"},
				{ID: "broker", Type: "customNode", Title: "Event Broker"},
				{ID: "workers", Type: "customNode", Title: "Workers"},
				{ID: "stream", Type: "customNode", Title: "Stream Processor"},
				{ID: "warehouse", Type: "customNode", Title: "Data Warehouse"},
				{ID: "bi", Type: "customNode", Title: "BI Platform"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "clients", Target: "gateway", Label: "requests"},
				{ID: "e2", Source: "gateway", Target: "broker", Label: "events"},
				{ID: "e3", Source: "broker", Target: "workers", Label: "dispatch"},
				{ID: "e4", Source: "workers", Target: "stream", Label: "transform"},
				{ID: "e5", Source: "stream", Target: "warehouse", Label: "store"},
				{ID: "e6", Source: "warehouse", Target: "bi", Label: "reports"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Cloud Native Deployment Pipeline",
			Description: "Production-grade deployment automation platform",
			GraphNodes: []providers.TemplateNode{
				{ID: "git", Type: "customNode", Title: "Git Repository"},
				{ID: "ci", Type: "customNode", Title: "CI Pipeline"},
				{ID: "registry", Type: "customNode", Title: "Container Registry"},
				{ID: "orchestrator", Type: "customNode", Title: "Orchestrator"},
				{ID: "cluster", Type: "customNode", Title: "Kubernetes Cluster"},
				{ID: "monitoring", Type: "customNode", Title: "Monitoring"},
				{ID: "alerts", Type: "customNode", Title: "Alerting"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "git", Target: "ci", Label: "commit"},
				{ID: "e2", Source: "ci", Target: "registry", Label: "publish"},
				{ID: "e3", Source: "registry", Target: "orchestrator", Label: "deploy"},
				{ID: "e4", Source: "orchestrator", Target: "cluster", Label: "manage"},
				{ID: "e5", Source: "cluster", Target: "monitoring", Label: "metrics"},
				{ID: "e6", Source: "monitoring", Target: "alerts", Label: "notify"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Distributed Collaboration Engine",
			Description: "Realtime collaborative editing infrastructure",
			GraphNodes: []providers.TemplateNode{
				{ID: "frontend", Type: "customNode", Title: "Frontend"},
				{ID: "socket", Type: "customNode", Title: "Socket Layer"},
				{ID: "presence", Type: "customNode", Title: "Presence Service"},
				{ID: "sync", Type: "customNode", Title: "Sync Engine"},
				{ID: "history", Type: "customNode", Title: "History Store"},
				{ID: "snapshots", Type: "customNode", Title: "Snapshots"},
				{ID: "recovery", Type: "customNode", Title: "Recovery Engine"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "frontend", Target: "socket", Label: "ws"},
				{ID: "e2", Source: "socket", Target: "presence", Label: "presence"},
				{ID: "e3", Source: "presence", Target: "sync", Label: "state"},
				{ID: "e4", Source: "sync", Target: "history", Label: "events"},
				{ID: "e5", Source: "history", Target: "snapshots", Label: "persist"},
				{ID: "e6", Source: "snapshots", Target: "recovery", Label: "restore"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "AI Inference Processing Pipeline",
			Description: "High-throughput AI inference infrastructure",
			GraphNodes: []providers.TemplateNode{
				{ID: "clients", Type: "customNode", Title: "Clients"},
				{ID: "balancer", Type: "customNode", Title: "Load Balancer"},
				{ID: "inference", Type: "customNode", Title: "Inference Cluster"},
				{ID: "scheduler", Type: "customNode", Title: "Task Scheduler"},
				{ID: "gpu", Type: "customNode", Title: "GPU Workers"},
				{ID: "results", Type: "customNode", Title: "Results Store"},
				{ID: "dashboards", Type: "customNode", Title: "Dashboards"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "clients", Target: "balancer", Label: "requests"},
				{ID: "e2", Source: "balancer", Target: "inference", Label: "route"},
				{ID: "e3", Source: "inference", Target: "scheduler", Label: "schedule"},
				{ID: "e4", Source: "scheduler", Target: "gpu", Label: "tasks"},
				{ID: "e5", Source: "gpu", Target: "results", Label: "outputs"},
				{ID: "e6", Source: "results", Target: "dashboards", Label: "analytics"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Global Media Distribution Platform",
			Description: "Distributed content delivery infrastructure",
			GraphNodes: []providers.TemplateNode{
				{ID: "upload", Type: "customNode", Title: "Upload Service"},
				{ID: "processing", Type: "customNode", Title: "Media Processing"},
				{ID: "transcoding", Type: "customNode", Title: "Transcoding"},
				{ID: "cdn", Type: "customNode", Title: "CDN"},
				{ID: "cache", Type: "customNode", Title: "Edge Cache"},
				{ID: "delivery", Type: "customNode", Title: "Delivery API"},
				{ID: "clients", Type: "customNode", Title: "Clients"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "upload", Target: "processing", Label: "assets"},
				{ID: "e2", Source: "processing", Target: "transcoding", Label: "convert"},
				{ID: "e3", Source: "transcoding", Target: "cdn", Label: "publish"},
				{ID: "e4", Source: "cdn", Target: "cache", Label: "replicate"},
				{ID: "e5", Source: "cache", Target: "delivery", Label: "serve"},
				{ID: "e6", Source: "delivery", Target: "clients", Label: "stream"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Security Monitoring Infrastructure",
			Description: "Enterprise-grade security analysis platform",
			GraphNodes: []providers.TemplateNode{
				{ID: "agents", Type: "customNode", Title: "Agents"},
				{ID: "collector", Type: "customNode", Title: "Log Collector"},
				{ID: "pipeline", Type: "customNode", Title: "Processing Pipeline"},
				{ID: "detection", Type: "customNode", Title: "Threat Detection"},
				{ID: "response", Type: "customNode", Title: "Response Engine"},
				{ID: "storage", Type: "customNode", Title: "Secure Storage"},
				{ID: "soc", Type: "customNode", Title: "SOC Dashboard"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "agents", Target: "collector", Label: "logs"},
				{ID: "e2", Source: "collector", Target: "pipeline", Label: "stream"},
				{ID: "e3", Source: "pipeline", Target: "detection", Label: "analyze"},
				{ID: "e4", Source: "detection", Target: "response", Label: "alerts"},
				{ID: "e5", Source: "response", Target: "storage", Label: "archive"},
				{ID: "e6", Source: "storage", Target: "soc", Label: "visualize"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Distributed Payment Processing System",
			Description: "Scalable payment transaction infrastructure",
			GraphNodes: []providers.TemplateNode{
				{ID: "checkout", Type: "customNode", Title: "Checkout"},
				{ID: "gateway", Type: "customNode", Title: "Payment Gateway"},
				{ID: "fraud", Type: "customNode", Title: "Fraud Detection"},
				{ID: "ledger", Type: "customNode", Title: "Transaction Ledger"},
				{ID: "settlement", Type: "customNode", Title: "Settlement"},
				{ID: "banking", Type: "customNode", Title: "Banking API"},
				{ID: "reporting", Type: "customNode", Title: "Reporting"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "checkout", Target: "gateway", Label: "payments"},
				{ID: "e2", Source: "gateway", Target: "fraud", Label: "verify"},
				{ID: "e3", Source: "fraud", Target: "ledger", Label: "record"},
				{ID: "e4", Source: "ledger", Target: "settlement", Label: "settle"},
				{ID: "e5", Source: "settlement", Target: "banking", Label: "transfer"},
				{ID: "e6", Source: "banking", Target: "reporting", Label: "reports"},
			},
		},
		{
			BoardType:   "graph",
			Title:       "Realtime Analytics Infrastructure",
			Description: "High-volume realtime analytics processing system",
			GraphNodes: []providers.TemplateNode{
				{ID: "sdk", Type: "customNode", Title: "Client SDK"},
				{ID: "ingestion", Type: "customNode", Title: "Ingestion API"},
				{ID: "stream", Type: "customNode", Title: "Stream Engine"},
				{ID: "aggregation", Type: "customNode", Title: "Aggregation"},
				{ID: "warehouse", Type: "customNode", Title: "Warehouse"},
				{ID: "dashboards", Type: "customNode", Title: "Dashboards"},
				{ID: "alerts", Type: "customNode", Title: "Alert Engine"},
			},
			GraphEdges: []providers.TemplateEdge{
				{ID: "e1", Source: "sdk", Target: "ingestion", Label: "events"},
				{ID: "e2", Source: "ingestion", Target: "stream", Label: "stream"},
				{ID: "e3", Source: "stream", Target: "aggregation", Label: "aggregate"},
				{ID: "e4", Source: "aggregation", Target: "warehouse", Label: "store"},
				{ID: "e5", Source: "warehouse", Target: "dashboards", Label: "visualize"},
				{ID: "e6", Source: "dashboards", Target: "alerts", Label: "thresholds"},
			},
		},
	},
}
