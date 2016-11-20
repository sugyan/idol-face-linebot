package app

type postbackAction string

var (
	postbackActionAccept = postbackAction("accept")
	postbackActionReject = postbackAction("reject")
)

type postbackData struct {
	Action      postbackAction `json:"action"`
	FaceID      int            `json:"face_id"`
	InferenceID int            `json:"inference_id"`
}
