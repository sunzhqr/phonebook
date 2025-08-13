package logger

import "go.uber.org/zap"

const (
	prod = "prod"
)

type Logger struct {
	*zap.Logger
}

type Field = zap.Field

func KV(k string, v any) Field {
	return zap.Any(k, v)
}

func Err(err error) Field {
	return zap.Error(err)
}

func New(env any) *Logger {
	var l *zap.Logger
	var err error
	if env == prod {
		cfg := zap.NewProductionConfig()
		//cfg.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}
		l, err = cfg.Build(zap.AddCaller())
	} else {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err) // кидать панику уместно при иницилизации
	}
	return &Logger{l}
}
