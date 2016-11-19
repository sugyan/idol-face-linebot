package message

type postbackAction string

// values of postback action
var (
	PostbackActionAccept = postbackAction("accept")
	PostbackActionReject = postbackAction("reject")
)

// PostbackData type
type PostbackData struct {
	Action      postbackAction `json:"action"`
	FaceID      int            `json:"face_id"`
	InferenceID int            `json:"inference_id"`
}
