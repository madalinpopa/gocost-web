package income

const (
	maxSourceLength = 255
)

type SourceVO struct {
	value string
}

func NewSourceVO(value string) (SourceVO, error) {
	if len(value) > maxSourceLength {
		return SourceVO{}, ErrSourceTooLong
	}
	return SourceVO{value: value}, nil
}

func (s SourceVO) Value() string {
	return s.value
}

func (s SourceVO) String() string {
	return s.value
}

func (s SourceVO) Equals(other SourceVO) bool {
	return s.value == other.value
}
