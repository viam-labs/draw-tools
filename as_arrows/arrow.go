package as_arrows

import (
	"drawmotionplan/lib"
	"fmt"

	commonPB "go.viam.com/api/common/v1"
	"go.viam.com/rdk/spatialmath"
	"google.golang.org/protobuf/types/known/structpb"
)

func (service *drawMotionPlanAsArrows) drawArrows(arrows []arrow) ([]commonPB.Transform, error) {
	data := []commonPB.Transform{}
	index := 0

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
			service.logger.Errorw("failed to create metadata", "error", err)
			return nil, err
		}

		data = append(data,
			commonPB.Transform{
				ReferenceFrame: fmt.Sprintf("arrow-%d", index),
				PoseInObserverFrame: &commonPB.PoseInFrame{
					ReferenceFrame: arrow.ParentFrame,
					Pose:           spatialmath.PoseToProtobuf(arrow.Pose),
				},
				Uuid:     lib.GenerateUUID(),
				Metadata: metadata,
			},
		)

		index++
	}

	return data, nil
}

func (service *drawMotionPlanAsArrows) parseArrows(drawData any) ([]arrow, error) {
	// Always expect an array of arrow objects
	arrowArray, ok := drawData.([]any)
	if !ok {
		return nil, fmt.Errorf("expected array of arrows, got %T", drawData)
	}

	arrows := make([]arrow, 0, len(arrowArray))
	for i, item := range arrowArray {
		arrowData, err := service.parseArrow(item)
		if err != nil {
			return nil, fmt.Errorf("failed to parse arrow at index %d: %w", i, err)
		}
		arrows = append(arrows, arrowData)
	}
	return arrows, nil
}

func (service *drawMotionPlanAsArrows) parseArrow(item any) (arrow, error) {
	arrowMap, ok := item.(map[string]any)
	if !ok {
		return arrow{}, fmt.Errorf("expected arrow object, got %T", item)
	}

	poseData, ok := arrowMap["pose"]
	if !ok {
		return arrow{}, fmt.Errorf("missing required 'pose' field")
	}

	pose, err := lib.ParsePose(poseData)
	if err != nil {
		return arrow{}, fmt.Errorf("failed to parse pose: %w", err)
	}

	color := lib.Color{R: 0, G: 0, B: 0}
	if colorData, ok := arrowMap["color"]; ok {
		parsed, err := lib.ParseColor(colorData)
		if err != nil {
			return arrow{}, fmt.Errorf("failed to parse color: %w", err)
		}

		color = parsed
	}

	parentFrame := "world"
	if frameData, ok := arrowMap["parent_frame"]; ok {
		if frameStr, ok := frameData.(string); ok {
			parentFrame = frameStr
		}
	}

	return arrow{
		Pose:        pose,
		Color:       color,
		ParentFrame: parentFrame,
	}, nil
}
