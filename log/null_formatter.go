package log

type NullFormatter struct{}

func (n *NullFormatter) Format(e Entry) ([]byte, error) {
	return nil, nil
}
