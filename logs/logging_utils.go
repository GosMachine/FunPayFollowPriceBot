package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	config := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:      []string{"logs/logfile.txt"},   // Укажите желаемый путь для записи логов
		ErrorOutputPaths: []string{"logs/errorfile.txt"}, // Укажите путь для записи ошибок
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}
	logger, _ := config.Build()
	Logger = logger
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(logger)
}
