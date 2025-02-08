package types

type Message struct {
	Body []byte `json:"body" validate:"required"`
	Type string `json:"type" validate:"oneof=name_change message"`
	From string
}
