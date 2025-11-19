package toUpbitPoint

type UpLimitType uint8

const (
	UpLimit115 UpLimitType = iota //1.15
	UpLimit110
	UpLimit105
	UpLimit125
	UpLimit130
	UpLimit140
	UpLimitErr
)

type PointLevel uint8

const (
	Point3 PointLevel = iota
	Point5
	Point10
	Point12
	Point14
	Point15
)

var (
	u64Arr115 = [6]float64{
		1.03,
		1.05,
		1.10,
		1.12,
		1.14,
		1.15,
	}
)

func GetPoint(upT UpLimitType, pointT PointLevel) float64 {
	switch upT {
	case UpLimit105:
	case UpLimit110:
	case UpLimit115:
		return u64Arr115[pointT]
	case UpLimit130:
	case UpLimit140:
	}
	return 0
}

func GetPointPre5(upT UpLimitType) float64 {
	switch upT {
	case UpLimit105:
	case UpLimit110:
	case UpLimit115:
		return u64Arr115[Point5]
	}
	return 0
}

func GetPointPre3(upT UpLimitType) float64 {
	switch upT {
	case UpLimit105:
	case UpLimit110:
	case UpLimit115:
		return u64Arr115[Point3]
	}
	return 0
}
