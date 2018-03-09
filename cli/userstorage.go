// test project main.go
package main

type storage struct {
	s map[string]map[string]interface{}
}

func newStorage() *storage {
	return &storage{
		make(map[string]map[string]interface{}),
	}
}

func (s storage) Save(user map[string]interface{}) error {
	if id, ok := user["id"]; ok {
		if idStr, is := id.(string); is {
			s.s[idStr] = user
		}
	}
	return nil

}

func (s storage) Find(fbid string) (map[string]interface{}, error) {
	if found, ok := s.s[fbid]; ok {
		return found, nil
	}
	return nil, nil
}
