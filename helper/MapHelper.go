package helper

type Compare func(s1 SliceItem, s2 SliceItem) bool

type SliceItem struct {
	Key   string
	Value interface{}
}

func MapToSlice(m map[string]interface{}) []SliceItem {
	var result []SliceItem
	for s, i := range m {
		result = append(result, SliceItem{Key: s, Value: i})
	}

	return result
}

func SortMap2Slice(m map[string]interface{}, fun Compare) []SliceItem {
	s := MapToSlice(m)
	count := len(s)
	for i := 0; i < count-1; i++ {
		for j := i + 1; j < count; j++ {
			b := fun(s[i], s[j])
			if b {
				s[i], s[j] = s[j], s[i]
			}
		}
	}

	return s
}
