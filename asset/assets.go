package asset

import "net/http"

type AssetFS interface {
	Asset(string) ([]byte, error)
	AssetHttp(string) (http.File, error)
	AssetDir(string) ([]string, error)
	AssetNames() []string
}

type Assets interface {
	GetAsset(string) (http.File, error)
	GetAssetByte(string) ([]byte, error)
	SetAssetFS(...AssetFS)
	ListAssetFS() []AssetFS
}

func New(af ...AssetFS) Assets {
	as := &assets{
		a: make([]AssetFS, 0),
	}
	as.SetAssetFS(af...)
	return as
}

type assets struct {
	a []AssetFS
}

var AssetUnavailable = Xrror("Asset %s unavailable").Out

func (a *assets) GetAsset(requested string) (http.File, error) {
	for _, x := range a.a {
		f, err := x.AssetHttp(requested)
		if err == nil {
			return f, nil
		}
	}
	return nil, AssetUnavailable(requested)
}

func (a *assets) GetAssetByte(requested string) ([]byte, error) {
	for _, x := range a.a {
		b, err := x.Asset(requested)
		if err == nil {
			return b, nil
		}
	}
	return nil, AssetUnavailable(requested)
}

func (a *assets) SetAssetFS(af ...AssetFS) {
	a.a = append(a.a, af...)
}

func (a *assets) ListAssetFS() []AssetFS {
	return a.a
}
