package entity

type Functions []*Function

func (f Functions) Exists(name string) bool {
	for _, fn := range f {
		if fn.Name == name {
			return true
		}
	}
	return false
}

func (f Functions) Find(name string) *Function {
	for _, fn := range f {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

type Function struct {
	Name string `toml:"name"`
}
