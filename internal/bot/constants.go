package bot

const (
	StepPrivacyAgreement  = "privacy_agreement"
	StepServiceSelection  = "service_selection"
	StepServiceType       = "service_type"
	StepDimensions        = "dimensions"
	StepDateSelection     = "date_selection"
	StepManualDateInput   = "manual_date_input"
	StepDateConfirmation  = "date_confirmation"
	StepContactMethod     = "contact_method"
	StepPhoneNumber       = "phone_number"
	StepTextureSelection  = "texture_selection"
	CustomTextureInput    = "custom_texture_input"
	StepPrintingSelection = "printing_selection"
	StepVinylSelection    = "vinyl_selection"
	StepPrintingOptions   = "printing_options"
	StepVinylOptions      = "vinyl_options"
)

// new state constants
const (
	StepMainMenu = "main_menu"
)

type LeatherDescription struct {
	Name              string
	PricePerDecimeter int64
	ImgURL            string
}

var LeatherTexture = map[string]LeatherDescription{
	"11111111-1111-1111-1111-111111111111": {
		Name:              "Натуральная кожа",
		PricePerDecimeter: 2500,
		ImgURL:            "https://example.com/tex1.jpg",
	},
	"22222222-2222-2222-2222-222222222222": {
		Name:              "Искусственная кожа",
		PricePerDecimeter: 1550,
		ImgURL:            "https://example.com/tex2.jpg",
	},
	"33333333-3333-3333-3333-333333333333": {
		Name:              "Замша",
		PricePerDecimeter: 3000,
		ImgURL:            "https://example.com/tex3.jpg",
	},
}
