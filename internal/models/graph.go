package models

// GNode represents a React Flow graph node (simplified - content only, no visual details)
type GNode struct {
	ID   string    `json:"id" jsonschema:"description=Уникальный идентификатор узла"`
	Type string    `json:"type" jsonschema:"description=Тип узла: customNode, externalLinkNode, internalLinkNode, nodeLinkNode"`
	Data GNodeData `json:"data" jsonschema:"description=Данные узла"`
}

// GNodeData represents the data of a graph node (content only, no visual details)
type GNodeData struct {
	Title       string        `json:"title,omitempty" jsonschema:"description=Заголовок узла"`
	Description string        `json:"description,omitempty" jsonschema:"description=Описание узла"`
	URL         string        `json:"url,omitempty" jsonschema:"description=URL для внешних ссылок"`
	Element     *GraphElement `json:"element,omitempty" jsonschema:"description=Элемент проекта для внутренних ссылок"`
}

// GraphElement represents an internal link element
type GraphElement struct {
	Type  string   `json:"type" jsonschema:"description=Тип элемента проекта"`
	Title string   `json:"title" jsonschema:"description=Заголовок элемента"`
	Path  []string `json:"path" jsonschema:"description=Путь к элементу"`
}

// GEdge represents a React Flow graph edge (simplified - connections only, no visual details)
type GEdge struct {
	ID     string `json:"id" jsonschema:"description=Уникальный идентификатор ребра"`
	Source string `json:"source" jsonschema:"description=ID исходного узла"`
	Target string `json:"target" jsonschema:"description=ID целевого узла"`
	Label  string `json:"label,omitempty" jsonschema:"description=Метка/описание связи"`
}

// SemanticGraph represents a semantic representation of a graph for LLM
type SemanticGraph struct {
	Nodes []SemanticNode `json:"nodes"`
	Edges []SemanticEdge `json:"edges"`
}

// SemanticNode represents a semantic graph node
type SemanticNode struct {
	ID          string `json:"id" jsonschema:"description=Уникальный идентификатор узла"`
	Label       string `json:"label" jsonschema:"description=Заголовок/название узла"`
	Kind        string `json:"kind" jsonschema:"description=Тип узла: component, service, database, external_link, internal_link, navigation"`
	Description string `json:"description,omitempty" jsonschema:"description=Описание узла"`
	URL         string `json:"url,omitempty" jsonschema:"description=URL для внешних ссылок"`
}

// SemanticEdge represents a semantic graph edge
type SemanticEdge struct {
	From  string `json:"from" jsonschema:"description=ID исходного узла"`
	To    string `json:"to" jsonschema:"description=ID целевого узла"`
	Label string `json:"label,omitempty" jsonschema:"description=Метка/описание связи"`
}
