package log

type JsonFormatter struct{}

func (j *JsonFormatter) Format(e Entry) ([]byte, error) {
	return nil, nil
}
