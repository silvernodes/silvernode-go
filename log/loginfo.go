package log

import (
	"fmt"
	"time"

	"github.com/silvernodes/silvernode-go/utils/jsonutil"
	"github.com/silvernodes/silvernode-go/utils/timeutil"
)

type LogInfo struct {
	Level    int
	Created  string
	Source   string
	Message  string
	Category string
}

func NewLogInfo(level int, created string, source string, message string, category string) *LogInfo {
	info := new(LogInfo)
	info.Level = level
	info.Created = created
	info.Source = source
	info.Message = message
	info.Category = category
	return info
}

func ParseLogInfo(str string) (*LogInfo, error) {
	l := new(LogInfo)
	err := jsonutil.Unmarshal(str, l)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *LogInfo) SetCreated(tm time.Time) {
	l.Created = tm.Format(timeutil.FORMAT_NOW_A)
}

func (l *LogInfo) ToJson() (string, error) {
	return jsonutil.Marshal(l)
}

func (l *LogInfo) FormatString() string {
	if l.Source == "" {
		return fmt.Sprintf("[%s] [%s] [%s] %s",
			l.Created,
			l.Category,
			LevelToString(l.Level),
			l.Message)
	}
	return fmt.Sprintf("[%s] [%s] [%s] (%s) %s",
		l.Created,
		l.Category,
		LevelToString(l.Level),
		l.Source,
		l.Message)
}
