package types

type Dict struct {
	d map[string]interface{}
}

func NewDict() *Dict {
	return &Dict{
		make(map[string]interface{}),
	}
}

func (d *Dict) Set(key Sds, value interface{}) {
	d.d[key.String()] = value
}

func (d *Dict) Get(key Sds) (interface{}, bool) {
	v := d.d[key.String()]
	if v == nil {
		return nil, false
	}
	return v, true
}

func (d *Dict) Delete(key Sds) {
	if _, ok := d.Get(key); !ok {
		return
	}
	delete(d.d, key.String())
}

func (d *Dict) Length() int {
	return len(d.d)
}
