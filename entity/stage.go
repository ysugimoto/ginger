package entity

type Stage struct {
	Name string `toml:"name"`
}

type Stages []*Stage

func (s Stages) Exists(name string) bool {
	for _, v := range s {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (s Stages) Add(name string) {
	s = append(s, &Stage{
		Name: name,
	})
}

func (s Stages) Remove(name string) Stages {
	ss := Stages{}
	for _, v := range s {
		if v.Name != name {
			ss = append(ss, v)
		}
	}
	return ss
}
