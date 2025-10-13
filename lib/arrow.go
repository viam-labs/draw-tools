package lib

import (
	"fmt"

	"github.com/google/uuid"
	commonPB "go.viam.com/api/common/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var defaultColor = Color{R: 255, G: 255, B: 0}

type Arrow struct {
	Pose        Pose   `json:"pose"`                   // required
	UUID        string `json:"uuid,omitempty"`         // optional, creates a new transform if not provided
	Color       Color  `json:"color,omitempty"`        // optional, defaults to black
	ParentFrame string `json:"parent_frame,omitempty"` // optional, defaults to "world"
}

func CreateArrowTransforms(arrows []Arrow) ([]commonPB.Transform, error) {
	data := []commonPB.Transform{}
	for _, arrow := range arrows {
		metadata, err := structpb.NewStruct(map[string]any{
			"shape": "arrow",
			"color": map[string]any{
				"r": int(arrow.Color.R),
				"g": int(arrow.Color.G),
				"b": int(arrow.Color.B),
			},
		})

		if err != nil {
			return nil, err
		}

		var id []byte
		if arrow.UUID == "" {
			id = GenerateUUID()
		} else {
			parsedId, err := uuid.Parse(arrow.UUID)
			if err != nil {
				return nil, err
			}
			id = parsedId[:]
		}

		data = append(data,
			commonPB.Transform{
				ReferenceFrame: fmt.Sprintf("arrow-%x", id),
				PoseInObserverFrame: &commonPB.PoseInFrame{
					ReferenceFrame: arrow.ParentFrame,
					Pose: &commonPB.Pose{
						X:     arrow.Pose.X,
						Y:     arrow.Pose.Y,
						Z:     arrow.Pose.Z,
						OX:    arrow.Pose.OX,
						OY:    arrow.Pose.OY,
						OZ:    arrow.Pose.OZ,
						Theta: arrow.Pose.Theta,
					},
				},
				Uuid:     id,
				Metadata: metadata,
			},
		)

	}

	return data, nil
}

func ParseArrows(drawData any) ([]Arrow, error) {
	arrowArray, ok := drawData.([]any)
	if !ok {
		return nil, fmt.Errorf("Expected array of arrows, got %T", drawData)
	}

	arrows := make([]Arrow, 0, len(arrowArray))
	for i, item := range arrowArray {
		arrowData, err := ParseArrow(item)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse arrow at index %d: %w", i, err)
		}
		arrows = append(arrows, arrowData)
	}
	return arrows, nil
}

func ParseArrow(item any) (Arrow, error) {
	arrowMap, ok := item.(map[string]any)
	if !ok {
		return Arrow{}, fmt.Errorf("Expected arrow object, got %T", item)
	}

	poseData, ok := arrowMap["pose"]
	if !ok {
		return Arrow{}, fmt.Errorf("Missing required 'pose' field")
	}

	pose, err := ParsePose(poseData)
	if err != nil {
		return Arrow{}, fmt.Errorf("Failed to parse pose: %w", err)
	}

	parentFrame := "world"
	if frameData, ok := arrowMap["parent_frame"]; ok {
		if frameStr, ok := frameData.(string); ok {
			parentFrame = frameStr
		} else {
			return Arrow{}, fmt.Errorf("Expected string for parent frame, got %T", frameData)
		}
	}

	result := Arrow{
		Pose:        pose,
		ParentFrame: parentFrame,
	}

	if colorData, ok := arrowMap["color"]; ok {
		parsed, err := ParseColor(colorData, defaultColor)
		if err != nil {
			return Arrow{}, fmt.Errorf("Failed to parse color: %w", err)
		}

		result.Color = parsed
	} else {
		result.Color = defaultColor
	}

	return result, nil
}
