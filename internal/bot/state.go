package bot

import (
	"adtime-bot/pkg/redis"
	"context"
	"fmt"
	"time"
)

// BOT STATE

type UserState struct {
	Step        string `json:"step"`
	Service     string `json:"service"`
	Date        string `json:"date"`
	PhoneNumber string `json:"phone_number"`
	WidthCM     int    `json:"width_cm"`
	HeightCM    int    `json:"height_cm"`
	TextureID   string `json:"texture_id"`
	Price       string `json:"price"`
}

type StateStorage struct {
	redis *redis.Client
	ttl   time.Duration
}


func (s *StateStorage) ClearState(chatID int64) {
	_ = s.Clear(context.Background(), chatID)
}

func (s *StateStorage) SetWaitingForPhoneNumber(chatID int64) {
	_ = s.SetStep(context.Background(), chatID, "waiting_phone_number")
}

func (s *StateStorage) SetWaitingForDateConfirmation(chatID int64) {
	_ = s.SetStep(context.Background(), chatID, "waiting_date_confirmation")
}

func (s *StateStorage) IsWaitingForManualDateInput(chatID int64) bool {
	state, _ := s.Get(context.Background(), chatID)
	return state.Step == "waiting_manual_date_input"
}

func (s *StateStorage) SetWaitingForManualDateInput(chatID int64) {
	_ = s.SetStep(context.Background(), chatID, "waiting_manual_date_input")
}

func (s *StateStorage) SetWaitingDimensions(chatID int64) {
	_ = s.SetStep(context.Background(), chatID, "waiting_dimensions")
}

func (s *StateStorage) SetWaitingForServiceInput(chatID int64) {
	_ = s.SetStep(context.Background(), chatID, "waiting_service_input")
}

func (s *StateStorage) SetWaitingForDateSelection(chatID int64) {
	_ = s.SetStep(context.Background(), chatID, "waiting_date_selection")
}

func NewStateStorage(redis *redis.Client) *StateStorage {
	return &StateStorage{
		redis: redis,
		ttl:   24 * time.Hour,
	}
}

func (s *StateStorage) Save(ctx context.Context, chatID int64, state UserState) error {
	return s.redis.SaveState(ctx, chatID, state)
}

func (s *StateStorage) Get(ctx context.Context, chatID int64) (UserState, error) {
	var state UserState
	err := s.redis.GetState(ctx, chatID, &state)
	return state, err
}

func (s *StateStorage) Clear(ctx context.Context, chatID int64) error {
	return s.redis.ClearState(ctx, chatID)
}

func (s *StateStorage) SetStep(ctx context.Context, chatID int64, step string) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}

	state.Step = step
	return s.Save(ctx, chatID, state)
}

func (s *StateStorage) SetDimensions(ctx context.Context, chatID int64, width, height int) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}

	state.WidthCM = width
	state.HeightCM = height
	return s.Save(ctx, chatID, state)
}

func (s *StateStorage) IsWaitingForPrivacyAgreement(chatID int64) bool {
	state, _ := s.Get(context.Background(), chatID)
	return state.Step == "waiting_privacy_agreement"
}

func (s *StateStorage) SetWaitingForPrivacyAgreement(chatID int64) error {
	return s.SetStep(context.Background(), chatID, "waiting_privacy_agreement")
}

func (s *StateStorage) IsWaitingForServiceSelection(chatID int64) bool {
	state, _ := s.Get(context.Background(), chatID)
	return state.Step == "waiting_service_selection"
}

func (s *StateStorage) SetWaitingForServiceSelection(chatID int64) error {
	return s.SetStep(context.Background(), chatID, "waiting_service_selection")
}

func (s *StateStorage) SetService(chatID int64, service string) error {
	state, err := s.Get(context.Background(), chatID)
	if err != nil {
		state = UserState{}
	}

	state.Service = service
	return s.Save(context.Background(), chatID, state)
}

func (s *StateStorage) SetDate(chatID int64, date string) error {
	state, err := s.Get(context.Background(), chatID)
	if err != nil {
		state = UserState{}
	}

	state.Date = date
	return s.Save(context.Background(), chatID, state)
}

func (s *StateStorage) SetPhoneNumber(chatID int64, phone string) error {
	state, err := s.Get(context.Background(), chatID)
	if err != nil {
		state = UserState{}
	}

	state.PhoneNumber = phone
	return s.Save(context.Background(), chatID, state)
}

func (s *StateStorage) IsWaitingDimensions(id int64) bool {
	state, _ := s.Get(context.Background(), id)
	return state.Step == "waiting_dimensions"
}

func (s *StateStorage) IsWaitingContact(id int64) bool {
	state, _ := s.Get(context.Background(), id)
	return state.Step == "waiting_contact"
}

func (s *StateStorage) IsWaitingForDateSelection(id int64) bool {
	state, _ := s.Get(context.Background(), id)
	return state.Step == "waiting_date_selection"
}

func (s *StateStorage) IsWaitingForPhoneNumber(id int64) bool {
	state, _ := s.Get(context.Background(), id)
	return state.Step == "waiting_phone_number"
}

func (s *StateStorage) GetDimensions(chatID int64) (width, height int) {
	state, _ := s.Get(context.Background(), chatID)
	return state.WidthCM, state.HeightCM
}

func (s *StateStorage) SetTexture(chatID int64, textureID string, price float64) error {
	state, err := s.Get(context.Background(), chatID)
	if err != nil {
		state = UserState{}
	}

	state.TextureID = textureID
	state.Price = fmt.Sprintf("%.2f", price)
	return s.Save(context.Background(), chatID, state)
}
