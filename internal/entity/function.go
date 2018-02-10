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

func (f Functions) Remove(name string) Functions {
	fns := Functions{}
	for _, fn := range f {
		if fn.Name == name {
			continue
		}
		fns = append(fns, fn)
	}
	return fns
}

type Function struct {
	Name       string `toml:"name"`
	Arn        string `toml:"arn"`
	MemorySize int64  `toml:"memory_size"`
	Timeout    int64  `toml:"timeout"`
}
