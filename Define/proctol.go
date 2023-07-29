package Define

const (
	ActionAdd         = "add"
	ActionEdit        = "edit"
	ActionDelete      = "delete"
	ActionUploaded    = "uploaded"
	ActionTranscoding = "transcoding"
	ActionReview      = "review"

	VideoTopic    = "video"
	ProgramTopic  = "program"
	EpisodesTopic = "episode"
	AdTopic       = "ad"
	PayTopic      = "pay"
	OrderTopic    = "order"
	RecordTopic   = "record"
	MessageTopic  = "message"

	StatusCreated             = 1
	StatusProcessing          = 2
	StatusComplete            = 3
	StatusCancel              = 4
	StatusCompletePayment     = 5
	StatusCompleteUpload      = 6
	StatusCompleteTranscoding = 7
	StatusCompleteScreenshot  = 8
	StatusCompleteReview      = 9

	MachineCancelConditionOverduePayment     = 2
	MachineCancelConditionOvertimeNotShipped = 3
)

type ICP[model any] struct {
	ItemId    string `json:"item_id"`
	ItemType  string `json:"item_type"`
	ExtraData model  `json:"extra_data"`
	Action    string `json:"action"`
	Status    int    `json:"status"`
}

type CancelOrder struct {
	OrderId                  string `json:"order_id"`
	Reason                   string `json:"reason"`
	IsSelfCancel             bool   `json:"is_self_cancel"`
	MachineClosingConditions int    `json:"machine_closing_conditions"`
}
