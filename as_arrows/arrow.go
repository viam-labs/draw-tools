package as_arrows

import (
	"fmt"

	commonPB "go.viam.com/api/common/v1"
	"go.viam.com/rdk/spatialmath"
	"google.golang.org/protobuf/types/known/structpb"
)

// DrawArrowsFromPoses draws a list of poses in the visualizer as arrows.
// Parameters:
//   - poses: a list of poses
//   - colors: Individual arrow color
func (service *drawMotionPlanAsArrows) drawArrowsFromPoses(poses []spatialmath.Pose) ([]commonPB.Transform, error) {
	data := []commonPB.Transform{}
	color := service.color
	index := 0

	for _, pose := range poses {
		metadata, err := structpb.NewStruct(map[string]any{
			"shape": "arrow",
			"color": color,
		})
		if err != nil {
			service.logger.Errorw("failed to create metadata", "error", err)
			return nil, err
		}

		data = append(data,
			commonPB.Transform{
				ReferenceFrame: fmt.Sprintf("arrow-%d", index),
				PoseInObserverFrame: &commonPB.PoseInFrame{
					ReferenceFrame: service.parentFrame,
					Pose:           spatialmath.PoseToProtobuf(pose),
				},
				Uuid:     generateUUID(),
				Metadata: metadata,
			},
		)

		index++
	}

	return data, nil
}
