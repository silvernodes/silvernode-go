package log

type ConsoleLogWriter struct {
}

func (c *ConsoleLogWriter) Write(info *LogInfo) {
	if info == nil {
		return
	}
	info.Println()
}

func (c *ConsoleLogWriter) Close() {

}

func NewConsoleLogWriter() *ConsoleLogWriter {
	c := new(ConsoleLogWriter)
	return c
}
