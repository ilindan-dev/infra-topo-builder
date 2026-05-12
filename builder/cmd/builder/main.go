package main

import "fmt"

func main() {
	fmt.Println("not implemented")
}

// func run(cfg config.Config, log *slog.Logger) error {
//	return nil
// }

// func mustMakeLogger(logLevel string) *slog.Logger {
//	var level slog.Level
//	switch logLevel {
//	case "DEBUG":
//		level = slog.LevelDebug
//	case "INFO":
//		level = slog.LevelInfo
//	case "ERROR":
//		level = slog.LevelError
//	default:
//		panic("unknown log level: " + logLevel)
//	}
//	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
//	return slog.New(handler)
// }
