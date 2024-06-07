package model

type ProductQuantity struct {
	ProductName string `json:"productName"`
	Quantity    int32  `json:"quantity"`
}

type OrderInfo struct {
	CustomerId        string            `json:"customerId"`
	ProductQuantities []ProductQuantity `json:"productQuantities"`
	TotalPrice        float32           `json:"totalPrice"`
}

type Status string

const (
	CREATED   Status = "CREATED"
	VALIDATED Status = "VALIDATED"
	CANCELED  Status = "CANCELED"
	FAILED    Status = "FAILED"
)

type Order struct {
	OrderInfo
	Id     string `json:"id"`
	Status Status `json:"status"`
}

type OrderEvent struct {
	Timestamp    int64  `json:"timestamp"`
	Order        Order  `json:"order"`
	ChangeReason string `json:"changeReason"`
}
