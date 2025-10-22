package errorx

import "github.com/hhh500/quantGoInfra/define/defineJson"

type Validate interface {
	TypeName() string
	Check() error
}

func ValidateWithWrap(v Validate) error {
	if err := v.Check(); err != nil {
		return TryAddMetadata(err, map[string]string{defineJson.ReqType: v.TypeName()})
	}
	return nil
}
