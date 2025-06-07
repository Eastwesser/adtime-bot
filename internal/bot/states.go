package bot

import (
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type UserState struct {
	CurrentMenu    string `json:"current_menu"`
    PrintingPage   int    `json:"printing_page"` 
    SelectedProduct string `json:"selected_product"`
	Step        string `json:"step"`
	Service     string `json:"service"`
	ServiceType string `json:"service_type"`
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

func (s *StateStorage) SetLastBotMessageID(ctx context.Context, chatID int64, messageID int) error {

	data, err := json.Marshal(messageID)
	if err != nil {
		return fmt.Errorf("failed to marshal message ID: %w", err)
	}

	key := fmt.Sprintf("last_msg:%d", chatID)
	if err := s.redis.Set(ctx, key, data, s.ttl); err != nil {
		return fmt.Errorf("failed to set last message ID: %w", err)
	}
	return nil
}

func (s *StateStorage) GetLastBotMessageID(ctx context.Context, chatID int64) (int, error) {
	key := fmt.Sprintf("last_msg:%d", chatID)
	data, err := s.redis.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get last message ID: %w", err)
	}

	var messageID int
	if err := json.Unmarshal(data, &messageID); err != nil {
		return 0, fmt.Errorf("failed to unmarshal message ID: %w", err)
	}

	return messageID, nil
}

func NewStateStorage(redis *redis.Client) *StateStorage {
	return &StateStorage{
		redis: redis,
		ttl:   24 * time.Hour,
	}
}

func (s *StateStorage) SetStep(ctx context.Context, chatID int64, step string) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}
	state.Step = step
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

func (s *StateStorage) SetTexture(ctx context.Context, chatID int64, textureID string, price float64) error {
	if textureID == "" {
		return errors.New("empty texture ID")
	}
	if price <= 0 {
		return fmt.Errorf("invalid price for texture %s: %.2f", textureID, price)
	}

	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}
	state.TextureID = textureID
	state.Price = fmt.Sprintf("%.2f", price)
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

func (s *StateStorage) SetServiceType(ctx context.Context, chatID int64, serviceType string) error {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		state = UserState{}
	}
	state.ServiceType = serviceType
	return s.Save(ctx, chatID, state)
}

func getStateKey(chatID int64) string {
	return fmt.Sprintf("state:%d", chatID)
}

func (s *StateStorage) GetFullState(ctx context.Context, chatID int64) (UserState, error) {
	return s.Get(ctx, chatID)
}

func (s *StateStorage) GetStep(ctx context.Context, chatID int64) (string, error) {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}
	return state.Step, nil
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

func (s *StateStorage) GetTexture(ctx context.Context, chatID int64) (*storage.Texture, error) {
	state, err := s.Get(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if state.TextureID == "" {
		return nil, fmt.Errorf("no texture selected")
	}

	// Return a basic texture with just the ID
	return &storage.Texture{
		ID:          state.TextureID,
		Name:        "Unknown Texture",
		PricePerDM2: 0.0,
	}, nil
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

func (s *StateStorage) SaveOrderState(ctx context.Context, chatID int64) error {
	return s.ResetOrderState(ctx, chatID)
}

func (s *StateStorage) ResetOrderState(ctx context.Context, chatID int64) error {
    // Get current state to preserve phone number
    currentState, err := s.Get(ctx, chatID)
    if err != nil {
        // If no state exists, start fresh
        currentState = UserState{}
    }

    // Reset all fields except phone number
    return s.Save(ctx, chatID, UserState{
        PhoneNumber: currentState.PhoneNumber, // Preserve phone
        Step:        StepServiceType,          // Or StepPrivacyAgreement if needed
    })
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

func (s *StateStorage) SetCurrentMenu(ctx context.Context, chatID int64, menu string) error {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        state = UserState{}
    }
    state.CurrentMenu = menu
    return s.Save(ctx, chatID, state)
}

func (s *StateStorage) GetCurrentMenu(ctx context.Context, chatID int64) (string, error) {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        return "", err
    }
    return state.CurrentMenu, nil
}

func (s *StateStorage) SetPrintingPage(ctx context.Context, chatID int64, page int) error {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        state = UserState{}
    }
    state.PrintingPage = page
    return s.Save(ctx, chatID, state)
}

func (s *StateStorage) GetPrintingPage(ctx context.Context, chatID int64) (int, error) {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        return 1, err
    }
    return state.PrintingPage, nil
}

func (s *StateStorage) SetSelectedProduct(ctx context.Context, chatID int64, product string) error {
    state, err := s.Get(ctx, chatID)
    if err != nil {
        state = UserState{}
    }
    state.SelectedProduct = product
    return s.Save(ctx, chatID, state)
}
