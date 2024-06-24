package entity

type Entity interface {
    Handle()    uint64
    Owner()     uint64

    GetLayerName() *string
}

type entity struct {
    handle      uint64
    owner       uint64

    LayerName   string
}

func (e *entity) Handle() uint64 {
    return e.handle
}

func (e *entity) Owner() uint64 {
    return e.owner
}

func (e *entity) GetLayerName() *string {
    return &e.LayerName
}
