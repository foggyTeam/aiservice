package preprocessing

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aiservice/internal/models"
	"github.com/firebase/genkit/go/ai"
)

// SpatialThresholds defines distance thresholds for spatial analysis
type SpatialThresholds struct {
	ClusterDistance    float32 // Distance threshold for clustering elements
	AlignmentTolerance float32 // Tolerance for alignment detection
}

// DefaultSpatialThresholds provides reasonable defaults
var DefaultSpatialThresholds = SpatialThresholds{
	ClusterDistance:    100.0,
	AlignmentTolerance: 10.0,
}

// ElementRelationship represents a relationship between two elements
type ElementRelationship struct {
	SourceID string
	TargetID string
	Type     string // "proximity", "alignment", "containment", "flow"
	Distance float32
	Angle    float32 // Angle in degrees
}

// SpatialCluster represents a group of related elements
type SpatialCluster struct {
	ID       string
	Elements []models.Element
	CenterX  float32
	CenterY  float32
	Bounds   BoundingBox
}

// BoundingBox represents rectangular bounds of an element or cluster
type BoundingBox struct {
	MinX, MinY, MaxX, MaxY float32
}

// Preprocessor transforms raw input data into structured formats for AI processing
type Preprocessor struct {
	thresholds SpatialThresholds
}

// NewPreprocessor creates a new preprocessor instance
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{
		thresholds: DefaultSpatialThresholds,
	}
}

// PreprocessSummarizeRequest transforms a raw summarize request into a structured format
func (p *Preprocessor) PreprocessSummarizeRequest(req models.SummarizeRequest) ([]*ai.Part, error) {
	// Preserve raw data
	rawData, err := json.Marshal(req.Board)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw board data: %w", err)
	}

	// Perform spatial analysis
	clusters := p.analyzeSpatialRelationships(req.Board.Elements)
	relationships := p.identifyElementRelationships(req.Board.Elements)

	// Create spatial analysis summary
	spatialAnalysis := p.createSpatialAnalysisSummary(clusters, relationships)

	// Create semantic annotations
	semanticAnnotations := p.annotateSemantics(req.Board.Elements)

	// Combine raw data with spatial and semantic information
	structuredPrompt := fmt.Sprintf(`BOARD ANALYSIS:
RAW DATA:
%s

SPATIAL ANALYSIS:
%s

SEMANTIC ANNOTATIONS:
%s

Please provide a summary of the key points and conclusions from this board, considering the spatial relationships and semantic groupings.`,
		string(rawData), spatialAnalysis, semanticAnnotations)

	parts := []*ai.Part{
		ai.NewTextPart(structuredPrompt),
	}

	// Add image if available
	if req.Board.ImageURL != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", req.Board.ImageURL))
	}

	return parts, nil
}

// PreprocessStructurizeRequest transforms a raw structurize request into a structured format
func (p *Preprocessor) PreprocessStructurizeRequest(req models.StructurizeRequest) ([]*ai.Part, error) {
	// Preserve raw data
	rawData, err := json.Marshal(req.Board)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw board data: %w", err)
	}

	// Perform spatial analysis
	clusters := p.analyzeSpatialRelationships(req.Board.Elements)
	relationships := p.identifyElementRelationships(req.Board.Elements)

	// Create spatial analysis summary
	spatialAnalysis := p.createSpatialAnalysisSummary(clusters, relationships)

	// Create semantic annotations
	semanticAnnotations := p.annotateSemantics(req.Board.Elements)

	// Create a structured representation of the file hierarchy
	fileStructure := p.createFileHierarchyDescription(req.File)

	// Combine raw data with spatial and semantic information
	structuredPrompt := fmt.Sprintf(`PROJECT STRUCTURIZATION REQUEST:
%s

FILE HIERARCHY REQUESTED:
%s

RAW BOARD DATA:
%s

SPATIAL ANALYSIS:
%s

SEMANTIC ANNOTATIONS:
%s

Please create a proper file structure based on the board content, considering the spatial relationships and semantic groupings.`,
		req.RequestType, fileStructure, string(rawData), spatialAnalysis, semanticAnnotations)

	parts := []*ai.Part{
		ai.NewTextPart(structuredPrompt),
	}

	// Add image if available
	if req.Board.ImageURL != "" {
		parts = append(parts, ai.NewMediaPart("image/jpeg", req.Board.ImageURL))
	}

	return parts, nil
}

// analyzeSpatialRelationships performs clustering and relationship analysis
func (p *Preprocessor) analyzeSpatialRelationships(elements []models.Element) []SpatialCluster {
	if len(elements) == 0 {
		return []SpatialCluster{}
	}

	// Create clusters based on proximity
	clusters := p.clusterElementsByProximity(elements)

	// Refine clusters based on alignment patterns
	clusters = p.refineClustersByAlignment(clusters)

	return clusters
}

// clusterElementsByProximity groups elements based on distance
func (p *Preprocessor) clusterElementsByProximity(elements []models.Element) []SpatialCluster {
	if len(elements) == 0 {
		return []SpatialCluster{}
	}

	visited := make(map[string]bool)
	var clusters []SpatialCluster

	for _, elem := range elements {
		if visited[elem.Id] {
			continue
		}

		// Start a new cluster with this element
		cluster := SpatialCluster{
			ID:       fmt.Sprintf("cluster_%s", elem.Id),
			Elements: []models.Element{elem},
		}

		// Add nearby elements to the cluster
		cluster = p.expandCluster(cluster, elements, visited)

		// Calculate cluster properties
		cluster.CenterX, cluster.CenterY = p.calculateClusterCenter(cluster.Elements)
		cluster.Bounds = p.calculateBoundingBox(cluster.Elements)

		clusters = append(clusters, cluster)
		visited[elem.Id] = true
	}

	return clusters
}

// expandCluster adds nearby elements to the cluster
func (p *Preprocessor) expandCluster(cluster SpatialCluster, elements []models.Element, visited map[string]bool) SpatialCluster {
	initialSize := len(cluster.Elements)

	for _, elem := range elements {
		if visited[elem.Id] {
			continue
		}

		// Check if this element is close enough to any element in the cluster
		isClose := false
		for _, clusterElem := range cluster.Elements {
			distance := p.calculateDistance(clusterElem, elem)
			if distance <= p.thresholds.ClusterDistance {
				isClose = true
				break
			}
		}

		if isClose {
			cluster.Elements = append(cluster.Elements, elem)
			visited[elem.Id] = true
		}
	}

	// If we added elements, recursively expand again to catch transitive relationships
	if len(cluster.Elements) > initialSize {
		cluster = p.expandCluster(cluster, elements, visited)
	}

	return cluster
}

// refineClustersByAlignment adjusts clusters based on alignment patterns
func (p *Preprocessor) refineClustersByAlignment(clusters []SpatialCluster) []SpatialCluster {
	// This is a simplified implementation - in a real system, we'd have more sophisticated alignment detection
	refinedClusters := make([]SpatialCluster, 0, len(clusters))

	for _, cluster := range clusters {
		// Check for horizontal/vertical alignment within the cluster
		alignedGroups := p.groupElementsByAlignment(cluster.Elements)

		if len(alignedGroups) > 1 {
			// Split the cluster into alignment-based sub-clusters
			for i, group := range alignedGroups {
				subCluster := SpatialCluster{
					ID:       fmt.Sprintf("%s_aligned_%d", cluster.ID, i),
					Elements: group,
				}
				subCluster.CenterX, subCluster.CenterY = p.calculateClusterCenter(group)
				subCluster.Bounds = p.calculateBoundingBox(group)
				refinedClusters = append(refinedClusters, subCluster)
			}
		} else {
			// Keep the original cluster if no alignment-based splits occurred
			refinedClusters = append(refinedClusters, cluster)
		}
	}

	return refinedClusters
}

// groupElementsByAlignment groups elements by alignment patterns
func (p *Preprocessor) groupElementsByAlignment(elements []models.Element) [][]models.Element {
	if len(elements) <= 1 {
		return [][]models.Element{elements}
	}

	var groups [][]models.Element

	// Group horizontally aligned elements
	horizontalGroups := p.findHorizontalAlignments(elements)
	groups = append(groups, horizontalGroups...)

	// Group vertically aligned elements
	verticalGroups := p.findVerticalAlignments(elements)
	groups = append(groups, verticalGroups...)

	// If no alignments found, return original elements as individual groups
	if len(groups) == 0 {
		for _, elem := range elements {
			groups = append(groups, []models.Element{elem})
		}
	}

	return groups
}

// findHorizontalAlignments finds groups of horizontally aligned elements
func (p *Preprocessor) findHorizontalAlignments(elements []models.Element) [][]models.Element {
	var groups [][]models.Element
	processed := make(map[string]bool)

	for i, elem1 := range elements {
		if processed[elem1.Id] {
			continue
		}

		group := []models.Element{elem1}
		processed[elem1.Id] = true

		for j, elem2 := range elements {
			if i == j || processed[elem2.Id] {
				continue
			}

			// Check if elements are horizontally aligned (similar Y coordinates)
			yDiff := math.Abs(float64(elem1.Y - elem2.Y))
			if float32(yDiff) <= p.thresholds.AlignmentTolerance {
				group = append(group, elem2)
				processed[elem2.Id] = true
			}
		}

		if len(group) > 1 {
			groups = append(groups, group)
		}
	}

	return groups
}

// findVerticalAlignments finds groups of vertically aligned elements
func (p *Preprocessor) findVerticalAlignments(elements []models.Element) [][]models.Element {
	var groups [][]models.Element
	processed := make(map[string]bool)

	for i, elem1 := range elements {
		if processed[elem1.Id] {
			continue
		}

		group := []models.Element{elem1}
		processed[elem1.Id] = true

		for j, elem2 := range elements {
			if i == j || processed[elem2.Id] {
				continue
			}

			// Check if elements are vertically aligned (similar X coordinates)
			xDiff := math.Abs(float64(elem1.X - elem2.X))
			if float32(xDiff) <= p.thresholds.AlignmentTolerance {
				group = append(group, elem2)
				processed[elem2.Id] = true
			}
		}

		if len(group) > 1 {
			groups = append(groups, group)
		}
	}

	return groups
}

// identifyElementRelationships identifies various types of relationships between elements
func (p *Preprocessor) identifyElementRelationships(elements []models.Element) []ElementRelationship {
	var relationships []ElementRelationship

	for i, elem1 := range elements {
		for j, elem2 := range elements {
			if i == j {
				continue
			}

			distance := p.calculateDistance(elem1, elem2)

			// Check for proximity relationship
			if distance <= p.thresholds.ClusterDistance {
				relationship := ElementRelationship{
					SourceID: elem1.Id,
					TargetID: elem2.Id,
					Type:     "proximity",
					Distance: distance,
					Angle:    p.calculateAngle(elem1, elem2),
				}
				relationships = append(relationships, relationship)
			}

			// Check for alignment relationship
			xDiff := math.Abs(float64(elem1.X - elem2.X))
			yDiff := math.Abs(float64(elem1.Y - elem2.Y))

			if float32(xDiff) <= p.thresholds.AlignmentTolerance {
				relationship := ElementRelationship{
					SourceID: elem1.Id,
					TargetID: elem2.Id,
					Type:     "vertical_alignment",
					Distance: distance,
					Angle:    p.calculateAngle(elem1, elem2),
				}
				relationships = append(relationships, relationship)
			}

			if float32(yDiff) <= p.thresholds.AlignmentTolerance {
				relationship := ElementRelationship{
					SourceID: elem1.Id,
					TargetID: elem2.Id,
					Type:     "horizontal_alignment",
					Distance: distance,
					Angle:    p.calculateAngle(elem1, elem2),
				}
				relationships = append(relationships, relationship)
			}
		}
	}

	return relationships
}

// calculateDistance calculates Euclidean distance between two elements' centers
func (p *Preprocessor) calculateDistance(elem1, elem2 models.Element) float32 {
	center1X := elem1.X + elem1.Width/2
	center1Y := elem1.Y + elem1.Height/2
	center2X := elem2.X + elem2.Width/2
	center2Y := elem2.Y + elem2.Height/2

	dx := center1X - center2X
	dy := center1Y - center2Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

// calculateAngle calculates angle between two elements in degrees
func (p *Preprocessor) calculateAngle(elem1, elem2 models.Element) float32 {
	center1X := elem1.X + elem1.Width/2
	center1Y := elem1.Y + elem1.Height/2
	center2X := elem2.X + elem2.Width/2
	center2Y := elem2.Y + elem2.Height/2

	dx := float64(center2X - center1X)
	dy := float64(center2Y - center1Y)
	angleRad := math.Atan2(dy, dx)
	angleDeg := float32(angleRad * 180 / math.Pi)

	// Normalize to 0-360 range
	if angleDeg < 0 {
		angleDeg += 360
	}

	return angleDeg
}

// calculateClusterCenter calculates the center point of a cluster
func (p *Preprocessor) calculateClusterCenter(elements []models.Element) (float32, float32) {
	if len(elements) == 0 {
		return 0, 0
	}

	var sumX, sumY float32
	for _, elem := range elements {
		centerX := elem.X + elem.Width/2
		centerY := elem.Y + elem.Height/2
		sumX += centerX
		sumY += centerY
	}

	count := float32(len(elements))
	return sumX / count, sumY / count
}

// calculateBoundingBox calculates the bounding box of elements
func (p *Preprocessor) calculateBoundingBox(elements []models.Element) BoundingBox {
	if len(elements) == 0 {
		return BoundingBox{}
	}

	minX := elements[0].X
	minY := elements[0].Y
	maxX := elements[0].X + elements[0].Width
	maxY := elements[0].Y + elements[0].Height

	for _, elem := range elements {
		elemMinX := elem.X
		elemMinY := elem.Y
		elemMaxX := elem.X + elem.Width
		elemMaxY := elem.Y + elem.Height

		if elemMinX < minX {
			minX = elemMinX
		}
		if elemMinY < minY {
			minY = elemMinY
		}
		if elemMaxX > maxX {
			maxX = elemMaxX
		}
		if elemMaxY > maxY {
			maxY = elemMaxY
		}
	}

	return BoundingBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}

// createSpatialAnalysisSummary creates a textual summary of spatial analysis
func (p *Preprocessor) createSpatialAnalysisSummary(clusters []SpatialCluster, relationships []ElementRelationship) string {
	var sb strings.Builder

	sb.WriteString("SPATIAL ANALYSIS RESULTS:\n")
	sb.WriteString(fmt.Sprintf("Number of spatial clusters identified: %d\n", len(clusters)))
	sb.WriteString(fmt.Sprintf("Number of element relationships identified: %d\n\n", len(relationships)))

	// Describe clusters
	sb.WriteString("CLUSTERS:\n")
	for i, cluster := range clusters {
		sb.WriteString(fmt.Sprintf("  Cluster %d (ID: %s):\n", i+1, cluster.ID))
		sb.WriteString(fmt.Sprintf("    Center: (%.2f, %.2f)\n", cluster.CenterX, cluster.CenterY))
		sb.WriteString(fmt.Sprintf("    Bounds: (%.2f, %.2f) to (%.2f, %.2f)\n",
			cluster.Bounds.MinX, cluster.Bounds.MinY, cluster.Bounds.MaxX, cluster.Bounds.MaxY))
		sb.WriteString(fmt.Sprintf("    Elements: %d\n", len(cluster.Elements)))

		// List element types in the cluster
		typeCounts := make(map[string]int)
		for _, elem := range cluster.Elements {
			typeCounts[elem.Type]++
		}

		sb.WriteString("    Element types: ")
		var typeList []string
		for elemType, count := range typeCounts {
			typeList = append(typeList, fmt.Sprintf("%s(%d)", elemType, count))
		}
		sb.WriteString(strings.Join(typeList, ", "))
		sb.WriteString("\n\n")
	}

	// Describe relationships
	sb.WriteString("RELATIONSHIPS:\n")

	// Group relationships by type
	proximityRels := make([]ElementRelationship, 0)
	horizontalAlignRels := make([]ElementRelationship, 0)
	verticalAlignRels := make([]ElementRelationship, 0)

	for _, rel := range relationships {
		switch rel.Type {
		case "proximity":
			proximityRels = append(proximityRels, rel)
		case "horizontal_alignment":
			horizontalAlignRels = append(horizontalAlignRels, rel)
		case "vertical_alignment":
			verticalAlignRels = append(verticalAlignRels, rel)
		}
	}

	sb.WriteString(fmt.Sprintf("  Proximity relationships: %d\n", len(proximityRels)))
	sb.WriteString(fmt.Sprintf("  Horizontal alignment relationships: %d\n", len(horizontalAlignRels)))
	sb.WriteString(fmt.Sprintf("  Vertical alignment relationships: %d\n", len(verticalAlignRels)))

	// Show some examples of relationships
	if len(proximityRels) > 0 {
		sb.WriteString("  Sample proximity relationships:\n")
		for i, rel := range proximityRels {
			if i >= 10 { // Limit to first 5 examples
				sb.WriteString("    ... (more relationships)\n")
				break
			}
			sb.WriteString(fmt.Sprintf("    - Element '%s' is %.2f units from element '%s'\n",
				rel.SourceID, rel.Distance, rel.TargetID))
		}
	}

	return sb.String()
}

// annotateSemantics adds semantic annotations to elements
func (p *Preprocessor) annotateSemantics(elements []models.Element) string {
	var sb strings.Builder

	sb.WriteString("SEMANTIC ANNOTATIONS:\n")

	// Sort elements by position for logical ordering
	sortedElements := make([]models.Element, len(elements))
	copy(sortedElements, elements)

	sort.Slice(sortedElements, func(i, j int) bool {
		// Sort by Y position first, then X position
		if sortedElements[i].Y != sortedElements[j].Y {
			return sortedElements[i].Y < sortedElements[j].Y
		}
		return sortedElements[i].X < sortedElements[j].X
	})

	// Analyze and annotate each element
	for i, elem := range sortedElements {
		sb.WriteString(fmt.Sprintf("Element %d (ID: %s):\n", i+1, elem.Id))
		sb.WriteString(fmt.Sprintf("  Type: %s\n", elem.Type))
		sb.WriteString(fmt.Sprintf("  Position: (%.2f, %.2f)\n", elem.X, elem.Y))
		sb.WriteString(fmt.Sprintf("  Size: (%.2f x %.2f)\n", elem.Width, elem.Height))

		// Semantic role inference
		role := p.inferSemanticRole(elem)
		sb.WriteString(fmt.Sprintf("  Inferred Role: %s\n", role))

		// Content analysis
		if elem.Content != "" {
			sb.WriteString(fmt.Sprintf("  Content: \"%s\"\n", elem.Content))

			// Analyze content type
			contentType := p.analyzeContentType(elem.Content)
			sb.WriteString(fmt.Sprintf("  Content Type: %s\n", contentType))
		}

		// Visual properties analysis
		if elem.Fill != "" {
			sb.WriteString(fmt.Sprintf("  Fill Color: %s\n", elem.Fill))
		}
		if elem.Stroke != "" {
			sb.WriteString(fmt.Sprintf("  Stroke Color: %s\n", elem.Stroke))
		}
		if elem.StrokeWidth != 0 {
			sb.WriteString(fmt.Sprintf("  Stroke Width: %d\n", elem.StrokeWidth))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// inferSemanticRole infers the semantic role of an element
func (p *Preprocessor) inferSemanticRole(elem models.Element) string {
	// Basic heuristics for semantic role inference
	if elem.Type == "text" {
		if elem.Width*elem.Height > 10000 { // Large text area likely a title or header
			return "header_or_title"
		} else if elem.Width > elem.Height*3 { // Wide text likely a label or description
			return "label_or_description"
		} else {
			return "content_text"
		}
	} else if elem.Type == "rect" || elem.Type == "rectangle" {
		if elem.Content != "" {
			return "text_container_with_label"
		} else if elem.Width > elem.Height*2 {
			return "horizontal_divider_or_section"
		} else if elem.Height > elem.Width*2 {
			return "vertical_divider_or_section"
		} else {
			return "container_or_card"
		}
	} else if elem.Type == "line" {
		if len(elem.Points) > 2 {
			return "drawing_or_signature"
		} else {
			return "connector_or_arrow"
		}
	} else {
		return "other_element"
	}
}

// analyzeContentType analyzes the type of content in text elements
func (p *Preprocessor) analyzeContentType(content string) string {
	content = strings.ToLower(content)

	// Check for common patterns
	if strings.Contains(content, "http://") || strings.Contains(content, "https://") {
		return "url_link"
	} else if strings.Contains(content, "@") && strings.Contains(content, ".") {
		return "email_address"
	} else if strings.Contains(content, ":") && (strings.Contains(content, "todo") || strings.Contains(content, "done")) {
		return "task_item"
	} else if strings.Contains(content, "- ") || strings.Contains(content, "* ") {
		return "list_item"
	} else if strings.Contains(content, "1.") || strings.Contains(content, "2.") || strings.Contains(content, "3.") {
		return "numbered_list_item"
	} else if strings.Contains(content, "title") || strings.Contains(content, "heading") || len(content) < 20 {
		return "title_or_heading"
	} else if len(content) > 200 {
		return "paragraph_or_description"
	} else {
		return "general_text"
	}
}

// createFileHierarchyDescription generates a description of the requested file hierarchy
func (p *Preprocessor) createFileHierarchyDescription(file models.File) string {
	var sb strings.Builder

	if file.IsEmpty() {
		sb.WriteString("No specific file structure requested. Create an appropriate structure based on board content.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Requested Structure: %s (%s)\n", file.Name, file.Type))

	if len(file.Children) > 0 {
		sb.WriteString("Child Elements:\n")
		p.writeFileTree(&sb, file.Children, 1)
	} else {
		sb.WriteString("No child elements specified.\n")
	}

	return sb.String()
}

// writeFileTree recursively writes the file tree structure
func (p *Preprocessor) writeFileTree(sb *strings.Builder, files []models.File, depth int) {
	indent := strings.Repeat("  ", depth)

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("%s- %s (%s)\n", indent, file.Name, file.Type))

		if len(file.Children) > 0 {
			p.writeFileTree(sb, file.Children, depth+1)
		}
	}
}

// PreprocessForSimilarity identifies similar requests to enable caching
func (p *Preprocessor) PreprocessForSimilarity(req models.AnalyzeRequest) string {
	switch req.RequestType {
	case models.SummarizeType:
		// Create a hashable representation for summarize requests
		return fmt.Sprintf("summarize:%s:%d",
			req.SummarizeRequest.Board.BoardID,
			len(req.SummarizeRequest.Board.Elements))
	case models.StructurizeType:
		// Create a hashable representation for structurize requests
		return fmt.Sprintf("structurize:%s:%d:%s",
			req.StructurizeRequest.Board.BoardID,
			len(req.StructurizeRequest.Board.Elements),
			req.StructurizeRequest.File.Name)
	default:
		return "unknown"
	}
}
