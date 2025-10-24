package idGen

import (
	"github.com/bwmarrin/snowflake"
)

func GetSnowIdInt64() (int64, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return -1, err
	}
	return node.Generate().Int64(), nil
}
