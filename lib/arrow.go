package lib

import (
	"fmt"

	commonPB "go.viam.com/api/common/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var defaultColor = Color{R: 255, G: 255, B: 0}

// ArrowJSON represents an arrow configuration in JSON format.
// It contains all the necessary information to create a visual arrow in the world state.
type ArrowJSON struct {
	Pose        PoseJSON `json:"pose"`                   // Position and orientation (required)
	Name        string   `json:"name"`                   // Name of the arrow frame (optional, defaults to "arrow-{uuid}")
	UUID        string   `json:"uuid,omitempty"`         // UUID string (optional, generates new UUID if not provided)
	Color       Color    `json:"color,omitempty"`        // RGB color (optional, defaults to yellow)
	ParentFrame string   `json:"parent_frame,omitempty"` // Parent reference frame (optional, defaults to "world")
}

// Arrow is a type alias for commonPB.Transform representing a visual arrow in the world state.
type Arrow = commonPB.Transform

// CreateArrow creates a new arrow transform from individual components.
// It generates a UUID if none is provided and uses default values for optional parameters.
//
// Parameters:
//   - pose: Position and orientation of the arrow (required)
//   - name: Name for the arrow frame (empty string will generate "arrow-{uuid}")
//   - uuid: Optional UUID bytes (generates new UUID if nil)
//   - color: Optional color (defaults to yellow if nil)
//   - parentFrame: Optional parent frame (defaults to "world" if nil)
//
// Returns the created arrow transform or an error if creation fails.
func CreateArrow(pose *commonPB.Pose, name string, uuid []byte, color *Color, parentFrame string) (*Arrow, error) {
	if pose == nil {
		return nil, fmt.Errorf("pose is required")
	}

	var id UUID
	if uuid == nil {
		id = GenerateUUID()
	} else {
		parsed, err := UUIDFromBytes(uuid)
		if err != nil {
			return nil, err
		}

		id = *parsed
	}

	if name == "" {
		name = defaultName(id)
	}

	metadataColor := map[string]any{"r": defaultColor.R, "g": defaultColor.G, "b": defaultColor.B}
	if color != nil {
		metadataColor = map[string]any{
			"r": int(color.R),
			"g": int(color.G),
			"b": int(color.B),
		}
	}

	metadata, err := structpb.NewStruct(map[string]any{
		"shape": "arrow",
		"color": metadataColor,
	})

	if err != nil {
		return nil, err
	}

	parent := "world"
	if parentFrame != "" {
		parent = parentFrame
	}

	return &Arrow{
		ReferenceFrame: name,
		PoseInObserverFrame: &commonPB.PoseInFrame{
			ReferenceFrame: parent,
			Pose:           pose,
		},
		Uuid:     id.Bytes(),
		Metadata: metadata,
	}, nil
}

// ParseArrows parses an array of arrows from JSON data.
// It expects an array of arrow objects and returns a slice of parsed arrows.
//
// Parameters:
//   - drawData: JSON array containing arrow objects
//
// Returns a slice of parsed arrows or an error if parsing fails.
func ParseArrows(drawData any) ([]*Arrow, error) {
	arrowArray, ok := drawData.([]any)
	if !ok {
		return nil, fmt.Errorf("Expected array of arrows, got %T", drawData)
	}

	arrows := make([]*Arrow, 0, len(arrowArray))
	for i, item := range arrowArray {
		arrowData, err := ParseArrow(item)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse arrow at index %d: %w", i, err)
		}
		arrows = append(arrows, arrowData)
	}
	return arrows, nil
}

// ParseArrow parses a single arrow from JSON data.
// It expects an arrow object with required pose and optional fields.
//
// Parameters:
//   - item: JSON object containing arrow data
//
// Returns the parsed arrow or an error if parsing fails.
func ParseArrow(item any) (*Arrow, error) {
	arrowMap, ok := item.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Expected arrow object, got %T", item)
	}

	poseData, ok := arrowMap["pose"]
	if !ok {
		return nil, fmt.Errorf("Missing required 'pose' field")
	}

	pose, err := ParsePose(poseData)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse pose: %w", err)
	}

	var id UUID
	idData, ok := arrowMap["uuid"].(string)
	if !ok {
		id = GenerateUUID()
	} else {
		parsed, err := UUIDFromString(idData)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse UUID: %w", err)
		}
		id = *parsed
	}

	name, ok := arrowMap["name"].(string)
	if !ok {
		if arrowMap["name"] == nil {
			name = defaultName(id)
		} else {
			return nil, fmt.Errorf("Expected string for name, got %T", arrowMap["name"])
		}
	}
	if name == "" {
		name = defaultName(id)
	}

	parentFrame := "world"
	if frameData, ok := arrowMap["parent_frame"]; ok {
		if frameStr, ok := frameData.(string); ok {
			parentFrame = frameStr
		} else {
			return nil, fmt.Errorf("Expected string for parent frame, got %T", frameData)
		}
	}

	var color *Color
	if colorData, ok := arrowMap["color"]; ok {
		parsed, err := ParseColor(colorData, defaultColor)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse color: %w", err)
		}

		color = &parsed
	} else {
		color = &defaultColor
	}

	bytes := id.Bytes()
	result, err := CreateArrow(pose, name, bytes, color, parentFrame)
	if err != nil {
		return nil, fmt.Errorf("Failed to create arrow: %w", err)
	}

	return result, nil
}

func defaultName(uuid UUID) string {
	return fmt.Sprintf("arrow-%s", uuid.String())
}
