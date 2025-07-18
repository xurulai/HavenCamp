package zlog

import (
	"go.uber.org/zap"
	"haven_camp_server/pkg/zlog"
	"testing"
)

func TestInfo(t *testing.T) {
	zlog.Info("this is a info", zap.String("name", "apylee"))
}
