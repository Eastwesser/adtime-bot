package bot

import (
	"adtime-bot/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
    StepPrivacyAgreement = "privacy_agreement"
    StepServiceSelection = "service_selection"
    StepServiceInput     = "service_input"
    StepServiceType      = "service_type"  // New state
    StepDimensions       = "dimensions"
    StepDateSelection    = "date_selection"
    StepManualDateInput  = "manual_date_input"
    StepDateConfirmation = "date_confirmation"
    StepContactMethod    = "contact_method" // New state
    StepPhoneNumber      = "phone_number"
)

// Update UserState struct to include ServiceType
type UserState struct {
    Step         string `json:"step"`
    Service      string `json:"service"`
    ServiceType  string `json:"service_type"` // New field
    Date         string `json:"date"`
    PhoneNumber  string `json:"phone_number"`
    WidthCM      int    `json:"width_cm"`
    HeightCM     int    `json:"height_cm"`
    TextureID    string `json:"texture_id"`
    Price        string `json:"price"`
}

type StateStorage struct {
	redis *redis.Client
	ttl   time.Duration
}

type State interface {
    GetTextureID(ctx context.Context, chatID int64) (string, error)
    SetTexture(ctx context.Context, chatID int64, textureID string, price float64) error
}

func NewStateStorage(redis *redis.Client) *StateStorage {
	return &StateStorage{
		redis: redis,
		ttl:   24 * time.Hour,
	}
}

func (s *StateStorage) Save(ctx context.Context, chatID int64, state UserState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := s.redis.Set(ctx, getStateKey(chatID), data, s.ttl); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

func (s *StateStorage) Get(ctx context.Context, chatID int64) (UserState, error) {
	data, err := s.redis.Get(ctx, getStateKey(chatID))
	if err != nil {
		return UserState{}, fmt.Errorf("failed to get state: %w", err)
	}

	var state UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return UserState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	return state, nil
}

func (s *StateStorage) GetFullState(ctx context.Context, chatID int64) (UserState, error) {
	return s.Get(ctx, chatID)
}

func (s *StateStorage) Clear(ctx context.Context, chatID int64) error {
	if err := s.redis.Del(ctx, getStateKey(chatID)); err != nil {
		return fmt.Errorf("failed to clear state: %w", err)
	}
	return nil
}

func (s *StateStorage) ClearState(ctx context.Context, chatID int64) error {
    return s.Clear(ctx, chatID)
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

func (s *StateStorage) SetService(ctx context.Context, chatID int64, service string) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}
	state.Service = service
	return s.Save(ctx, chatID, state)
}

func (s *StateStorage) SetDate(ctx context.Context, chatID int64, date string) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}
	state.Date = date
	return s.Save(ctx, chatID, state)
}

func (s *StateStorage) SetPhoneNumber(ctx context.Context, chatID int64, phone string) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}
	state.PhoneNumber = phone
	return s.Save(ctx, chatID, state)
}

func (s *StateStorage) SetTexture(ctx context.Context, chatID int64, textureID string, price float64) error {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        state = UserState{}
    }
    state.TextureID = textureID
    state.Price = fmt.Sprintf("%.2f", price)
    return s.Save(ctx, chatID, state)
}

func (s *StateStorage) GetTextureID(ctx context.Context, chatID int64) (string, error) {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        return "", err
    }
    return state.TextureID, nil
}

func (s *StateStorage) GetDimensions(ctx context.Context, chatID int64) (width, height int, err error) {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		return 0, 0, err
	}
	return state.WidthCM, state.HeightCM, nil
}

func getStateKey(chatID int64) string {
	return fmt.Sprintf("state:%d", chatID)
}

func (s *StateStorage) SetServiceType(ctx context.Context, chatID int64, serviceType string) error {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        state = UserState{}
    }
    state.ServiceType = serviceType
    return s.Save(ctx, chatID, state)
}


