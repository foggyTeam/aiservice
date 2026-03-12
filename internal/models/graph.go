package models

// GNode represents a React Flow graph node
type GNode struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"` // customNode, externalLinkNode, internalLinkNode, nodeLinkNode
	Position GPosition `json:"position"`
	Data     GNodeData `json:"data"`
}

// GPosition represents the position of a graph node
type GPosition struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// GNodeData represents the data of a graph node
type GNodeData struct {
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Shape       string        `json:"shape,omitempty"`
	Color       string        `json:"color,omitempty"`
	URL         string        `json:"url,omitempty"`
	Domain      string        `json:"domain,omitempty"`
	Favicon     string        `json:"favicon,omitempty"`
	ThumbnailURL string       `json:"thumbnailUrl,omitempty"`
	Element     *GraphElement `json:"element,omitempty"`
}

// GraphElement represents an internal link element
type GraphElement struct {
	Type  string   `json:"type"`
	Title string   `json:"title"`
	Path  []string `json:"path"`
}

// GEdge represents a React Flow graph edge
type GEdge struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"` // default, smoothstep, straight, step, simplebezier
	Source       string      `json:"source"`
	Target       string      `json:"target"`
	SourceHandle string      `json:"sourceHandle,omitempty"`
	TargetHandle string      `json:"targetHandle,omitempty"`
	Label        string      `json:"label,omitempty"`
	Animated     bool        `json:"animated,omitempty"`
	Style        *GEdgeStyle `json:"style,omitempty"`
}

// GEdgeStyle represents the style of a graph edge
type GEdgeStyle struct {
	Stroke        string `json:"stroke,omitempty"`
	StrokeWidth   int    `json:"strokeWidth,omitempty"`
	StrokeLinecap string `json:"strokeLinecap,omitempty"`
	StrokeDasharray string `json:"strokeDasharray,omitempty"`
}

// SemanticGraph represents a semantic representation of a graph for LLM
type SemanticGraph struct {
	Nodes []SemanticNode `json:"nodes"`
	Edges []SemanticEdge `json:"edges"`
}

// SemanticNode represents a semantic graph node
type SemanticNode struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Kind        string `json:"kind"` // component, service, database, external_link, internal_link, navigation
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// SemanticEdge represents a semantic graph edge
type SemanticEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}
