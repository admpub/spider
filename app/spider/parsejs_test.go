package spider

import (
	"testing"

	"github.com/webx-top/com"
)

func TestGetSpiderModels(t *testing.T) {
	ms := getSpiderModels()
	com.Dump(ms)
}
