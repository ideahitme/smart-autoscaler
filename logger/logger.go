package logger

import "go.uber.org/zap"

var Log *zap.Logger

func init() {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	// 	zap.NewJSONEncoder(
	// 		zap.RFC3339Formatter("@timestamp"),
	// 		zap.MessageKey("@message"),
	// 		zap.LevelString("@level"),
	// 	),
	// 	zap.DebugLevel,
	// )
}
