package accountConfig

type Config struct {
	ApiKeyHmac string
	SecretHmac string
	Email      string
	AccountId  uint8
}

var (
	Trades   []Config
	Monitors []Config
)
