package state_manager

import (
	"adtime-bot/internal/bot"
	"adtime-bot/internal/storage/redis"
	"context"
	"fmt"
)

type UserDialogStateManager struct {
	redisStorage RedisStorage
}

func New(redisStorage RedisStorage) *UserDialogStateManager {
	return &UserDialogStateManager{redisStorage: redisStorage}
}

func (u *UserDialogStateManager) GetUserDialogState(ctx context.Context, chatId int64) (*redis.UserState, error) {
	state, err := u.redisStorage.GetUserDialogState(ctx, chatId)
	if err != nil {
		return nil, fmt.Errorf("redisStorage.GetUserDialogState failed: %w", err)
	}
	return state, nil
}

func (u *UserDialogStateManager) setUserDialogState(ctx context.Context, chatId int64, state *redis.UserState) error {
	return u.redisStorage.SetUserDialogState(ctx, chatId, state)
}

func (u *UserDialogStateManager) SetStep(ctx context.Context, chatID int64, step string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.Step = step

	return u.setUserDialogState(ctx, chatID, state)
}

/*func (u *UserDialogStateManager) SetService(ctx context.Context, chatID int64, service string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.Service = service

	return u.setUserDialogState(ctx, chatID, state)
}*/

/*func (u *UserDialogStateManager) SetTexture(ctx context.Context, chatID int64, textureID string, price int64) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.TextureID = textureID
	state.Price = fmt.Sprintf("%d", price)
	return u.setUserDialogState(ctx, chatID, state)
}
*/

/*func (u *UserDialogStateManager) SetDimensions(ctx context.Context, chatID int64, width, height int) error {

	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.WidthCM = width
	state.HeightCM = height

	return u.setUserDialogState(ctx, chatID, state)
}*/

/*func (u *UserDialogStateManager) SetDate(ctx context.Context, chatID int64, date string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}
	state.Date = date
	return u.setUserDialogState(ctx, chatID, state)
}*/

/*func (u *UserDialogStateManager) SetPhoneNumber(ctx context.Context, chatID int64, phone string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}
	state.PhoneNumber = phone
	return u.setUserDialogState(ctx, chatID, state)
}*/

/*func (u *UserDialogStateManager) SetServiceType(ctx context.Context, chatID int64, serviceType string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.ServiceType = serviceType
	return u.setUserDialogState(ctx, chatID, state)
}

*/

func (u *UserDialogStateManager) ResetDialogState(ctx context.Context, chatID int64) error {
	// Get current state to preserve phone number
	prevState, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	// Reset all fields except phone number
	return u.setUserDialogState(ctx, chatID, &redis.UserState{
		Step: bot.StepMainMenu,
		Userdata: &redis.UserData{
			PhoneNumber: prevState.Userdata.PhoneNumber,
		},
	})
}

func (u *UserDialogStateManager) ClearState(ctx context.Context, chatID int64) error {
	return u.redisStorage.DropUserDialogState(ctx, chatID)
}

func (u *UserDialogStateManager) SetCurrentMenu(ctx context.Context, chatID int64, menu string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.CurrentMenu = menu
	return u.setUserDialogState(ctx, chatID, state)
}

func (u *UserDialogStateManager) SetPrintingPage(ctx context.Context, chatID int64, page int) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.PrintingPage = page
	return u.setUserDialogState(ctx, chatID, state)
}

func (u *UserDialogStateManager) SetSelectedProduct(ctx context.Context, chatID int64, product string) error {
	state, err := u.GetUserDialogState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("GetUserDialogState failed: %w", err)
	}

	state.SelectedProduct = product
	return u.setUserDialogState(ctx, chatID, state)
}
