package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Conf struct {
	Level    string
	FileName string
}

type Logger struct {
	*logrus.Logger
	logFile *os.File
}

type CleanFormatter struct{}

func (f *CleanFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message

	return []byte(fmt.Sprintf("%s [%s] %s\n", level, timestamp, message)), nil
}

func New(conf Conf) (*Logger, error) {
	logger := logrus.New()

	logLevel, err := logrus.ParseLevel(conf.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(logLevel)

	logger.SetFormatter(&CleanFormatter{})

	var logFile *os.File
	var output io.Writer = os.Stdout

	if conf.FileName != "" {
		file, err := os.OpenFile(conf.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logFile = file
		output = io.MultiWriter(os.Stdout, file)
	}

	logger.SetOutput(output)

	return &Logger{
		Logger:  logger,
		logFile: logFile,
	}, nil
}

func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

func (l *Logger) SetLogLevel(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	l.Logger.SetLevel(logLevel)
	l.Infof("change log level on: %s", level)
	return nil
}
