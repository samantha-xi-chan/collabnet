package logrus_wrap

import (
	"context"
	"github.com/sirupsen/logrus"
)

func GetContextLogger(ctx context.Context) (x *logrus.Logger) {
	k := "logger"
	v, ok := ctx.Value(k).(*logrus.Logger)
	if !ok {
		return nil
	}
	return v
}
func SetContextLogger(ctx context.Context, v interface{}) (x context.Context) {
	k := "logger"
	return context.WithValue(ctx, k, v)
}
