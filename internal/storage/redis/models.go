package redis

type UserState struct {
	Step     string    `json:"step"`
	Userdata *UserData `json:"user_data,omitempty"`
	Order    *Order    `json:"order,omitempty"`
}

type Order struct {
	SelectedProduct *string `json:"selected_product,omitempty"`

	Leather    *Leather    `json:"leather,omitempty"`
	Typography *Typography `json:"typography,omitempty"`
	Sticker    *Stickers   `json:"sticker,omitempty"`

	Price    *string   `json:"price,omitempty"`
	Delivery *Delivery `json:"delivery,omitempty"`
}

type Delivery struct {
	Date *string `json:"date,omitempty"`
}

type UserData struct {
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type Leather struct {
	// текстура кожи
	TextureID *string `json:"texture_id,omitempty"`
	// размер
	WidthCM  *int `json:"width_cm,omitempty"`
	HeightCM *int `json:"height_cm,omitempty"`
}

type Typography struct {
}

type Stickers struct {
}
