package accountConfig

type Config struct {
	ApiKeyHmac    string
	SecretHmac    string
	ApiKeyEd25519 string
	SecretEd25519 string
	Email         string
	Uid           uint32
	AccountId     uint8
}

var (
	Trades   []Config
	Monitors []Config
)
