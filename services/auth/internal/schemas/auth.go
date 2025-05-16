package schemas

type (
	SendConfirmationCodeSchema struct {
		email string
	}
	SignUpSchema struct {
		Username         string
		Email            string
		FirstName        string
		LastName         string
		ConfirmationCode string
	}
)
