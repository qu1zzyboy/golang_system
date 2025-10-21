package accountConfig

import (
	"context"
)

const (
	MODULE_ID = "account_config" //全局定时任务
)

type Boot struct {
}

func NewBoot() *Boot {
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{}
}

func (s *Boot) Start(ctx context.Context) error {
	Trades = append(Trades, Config{
		AccountId:  0,
		ApiKeyHmac: "PNzK8MP5fuOITM1eRBj2SJqSzv6IdczWxKCHzEA9OP3Whc20CxgRw8kkEgWqkUpI",
		SecretHmac: "GjeDmsX1irPEuVEIIHwiVQOEg6gN4i66ySJBv6ic3TqZBMK1o0rpGR1mihPm1ib7",
		Email:      "z2282915646@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  1,
		ApiKeyHmac: "TBwrcCUb7kIIM3x58AuxbeEN62i6QwFg1f6tz6ScuVVpvEvqTUKByoa903HHfC6G",
		SecretHmac: "iG6xTgyUfAeReVkR1LgB6IdHFxjr7mQ2qM3YUPcgyMmw40VmQfOoa9Sixd5YxdDo",
		Email:      "office_quantity1@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  2,
		ApiKeyHmac: "5EKpaIl6kV3PsWAYJywWutmfvl5Oh30ST2NdsuIzdeTIpzPmKzmngJMOx1igbcAd",
		SecretHmac: "46vN4aLas5BOlwxV1xi92zMlf9wRPfog2IOG1OH2Q39dbElKBUv0WgzNVy5v9O5s",
		Email:      "office_quantity2@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  3,
		ApiKeyHmac: "ZWMMEcssJO69hr3vKiQRp0xG76tKTcdo4aaOEJaAhIX9osnQzH5mkMDIn6xkpEPT",
		SecretHmac: "z08fE9EzynBMWJaAdI5Lumv76H8Hus3zbGCAIxT0hjImAzexw4eahYYJFwPftrOO",
		Email:      "office_quantity3@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  4,
		ApiKeyHmac: "qBKaZoJtXr4fdGS1JqwZpbgxroJQTvpNdGFSYVKt7bUVoigFZWZemEIip8aubsZU",
		SecretHmac: "cuLAljwx7MRjHqFCMGFG97FKiBVSqI47njxL8FIjiqbH2Arwix6kQMwz7E8xMmFA",
		Email:      "office_quantity4@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  5,
		ApiKeyHmac: "XJ4rlGR9JAycrO7c20yH8pfymmXngOeRxPMk2Om0Vft4WeiVQM2v5eg5UrQRYaBN",
		SecretHmac: "ZyXjc87p3NxBQEoHNS84Yns5Hhkzf60GH08dfJOrVFUjSyhlKfrzBlo1scC6K1Uy",
		Email:      "office_quantity5@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  6,
		ApiKeyHmac: "tMM23WljGGnpIisztHjZ5vWrnKPSqpOfMEAt5COFYsub6ju2LTGCOx0vBFTWx8uQ",
		SecretHmac: "lHKbCatL9Gz2g9Hg9Tr5dneZDhdLMwMrc6WtS0DakKssrdCj4CavuDg7TWY0TTLl",
		Email:      "office_quantity6@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  7,
		ApiKeyHmac: "c2Y1zMXaZcz85k6MrQY1Qo3FEEy81ookmcI3js0KJrMfT0EL5pgvwSgHKfSbu7aH",
		SecretHmac: "AN6jUNWjRjuFLj85uVd41TSsz8y1qAxuYlgdyXcLK30tzhrmjIuAJLzhGAdKg6CT",
		Email:      "office_quantity7@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  8,
		ApiKeyHmac: "0CIVowNTr0K1tPF7MNTbZMhncEfZpsgdPbUapZzaYgfsG03k3ewbSpc5vZc5BpaI",
		SecretHmac: "vuHz3IRWbRPl9numqHOTeOAvfewsHOJtWDNT7gKoNqQpYP9166t0IMwgme7YEacY",
		Email:      "office_quantity8@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  9,
		ApiKeyHmac: "ZKV6gexsFQphxRFyXWKPzSHVRK15yEiCIocnD7CREiVnCfwKSAkVLMyYVMx9FN9w",
		SecretHmac: "N6MREsxp3CZMbaHrBbv9ruVmekTlpLYoiQYd1yLX0WuQsioOUMnvuy8CRdx6ysJU",
		Email:      "office_quantity9@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  10,
		ApiKeyHmac: "9df3d8lIJtG9NeLaUUaQEIxjNaQ3KZAf6RGjcUoOZq3R4UH1rFkzJFTGaJF5cvkU",
		SecretHmac: "YhfptaP6cCsLg2vNohny0C5HqwsImNV0aY1WuhCOMxL16FhiHX36J9p4K7QnIMCX",
		Email:      "office_quantity10@163.com",
	})
	Trades = append(Trades, Config{
		AccountId:  11,
		ApiKeyHmac: "j0G54QHYH68pETkd0K9keDW2U02woj9w5mJ8EoFrv8tlRi0fCzp8XjwTXqFVNtho",
		SecretHmac: "F7HmmiZnDUpgVgOW0CAFrVn7FcqFUN8t9UbYxQRKqZlIURE9sZiy7hqhXbkMIOC2",
		Email:      "chunpingzhan888@gmail.com",
	})
	Monitors = append(Monitors, Config{
		AccountId:  0,
		ApiKeyHmac: "PNzK8MP5fuOITM1eRBj2SJqSzv6IdczWxKCHzEA9OP3Whc20CxgRw8kkEgWqkUpI",
		SecretHmac: "GjeDmsX1irPEuVEIIHwiVQOEg6gN4i66ySJBv6ic3TqZBMK1o0rpGR1mihPm1ib7",
		Email:      "z2282915646@163.com",
	})
	Monitors = append(Monitors, Config{
		AccountId:  1,
		ApiKeyHmac: "tHxmOwZ24QDCiopCSn00lOM0VMc4dLIKMmUUIH8tg3AeWEG8PFpWqcwfOucw5LMz",
		SecretHmac: "HvRUrDWcp8jxqi77XnnjHaGh1QsYmlDn4GcgRDFutYIZElcNYwyiUamiqjbG8BCX",
	})
	Monitors = append(Monitors, Config{
		AccountId:  2,
		ApiKeyHmac: "vbixBgNmZHx8AdeuC6FD3BNhMIDFNdB4FVV0EgjWpm1Oye2GSgE5DU9cUlAtQh6x",
		SecretHmac: "hBnBcgceIfXlwk1FjDY9PTD1ZyT3RzCnUkhmBjTKFw64FkQp8CqnXTtgyu5svnI9",
	})
	Monitors = append(Monitors, Config{
		AccountId:  3,
		ApiKeyHmac: "5HMFBsrlhCveVyQOtQuDuwE1w9aobQsdn83rDBnk1j2VSP28ZYMsh8UfymQfsb5g",
		SecretHmac: "sGFMOryiNv4Zx2bg1VjVz4US70SnR1TwCOP14zf9t7AbVZUEbVvoBw2UtH2P3DIs",
	})
	Monitors = append(Monitors, Config{
		AccountId:  4,
		ApiKeyHmac: "ASaZWqFrXYAPymhfUoOk8OSxI74eUD4m8n2FnoOgmL5IroLxh6v5EQNknJE0KDCM",
		SecretHmac: "FcdHvEVBSNa3bg2UNHLvWZdnPLL7WIDqoX6glbyycDuBTufOEybGD8ru2IIrqXwT",
	})
	Monitors = append(Monitors, Config{
		AccountId:  5,
		ApiKeyHmac: "LVvealSl1dcbwVRZo27os8YZur5w5NIvZUjYiWTHDGllviThlTh72t1VIAtQ6OYI",
		SecretHmac: "6IpgKGXI4JjYDBDf3J2D8F5Y70bqcW3zWSKXLDgbKau1I9r5y1AqaEHMsyELnzOT",
	})
	Monitors = append(Monitors, Config{
		AccountId:  6,
		ApiKeyHmac: "5mh5dEX6jQEQDpvaz60K2CdQmGcmBERHJKvI3Dw0hQd027AT6xlKao2bFTEYwQna",
		SecretHmac: "kudx9ZwZ4yIx5hqzCqttJMvGPciMp9GIcX287FBJNwtGFasBT6aF5QWZB5flkCP3",
	})
	Monitors = append(Monitors, Config{
		AccountId:  7,
		ApiKeyHmac: "RCzGiKQgwpoEzF18NAMvqqx4HKdHufAHaCXnah4Z7fvzRkpyfnSvFlpQXZlfLvLG",
		SecretHmac: "BaE7ngmD1hEFafNkns7RYS33WkTalPtddBPnh0G2xBNPMHtjKqjpBWAL7ZCIuq9e",
	})
	Monitors = append(Monitors, Config{
		AccountId:  8,
		ApiKeyHmac: "wDB048vbdXWcoEzy3J3bsWSXbrQHSEX1hQ5vO9uyMAdkWALfBHupbuATHtXCeiod",
		SecretHmac: "4Hsh5iP6DNhIrgK7Xten7aqBCn7aVCsgq8ovp78lpyqjLc9pvuQb80QDds2N1TVr",
	})
	Monitors = append(Monitors, Config{
		AccountId:  9,
		ApiKeyHmac: "DDRqNzZM30ierlPgS3V4glNS6mTummWB45OsfeOhGEn1wLN8liwrp86BA3Fm73F0",
		SecretHmac: "RAx9doyoCsdL49q5DovT0dOUc0mrdMZt94urxRewegp1Nlb9o0kt84BfR4sNoeDh",
	})
	Monitors = append(Monitors, Config{
		AccountId:  10,
		ApiKeyHmac: "QAuecbVrNgLGJIiOZ8cAUcVFKrjaNjM1cJu23gzxiuacnkg3skTvp8LPBsCH4OdX",
		SecretHmac: "uoh4HrdVCjhy78p1xFlqOQ7kvrIJmiWn96HBRtwdgwmbwUqRCeHWvJRHwUaKh0Ug",
	})
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
