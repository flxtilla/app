package asset

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/thrisp/flotilla/xrr"
)

type FakeFile struct {
	Path string
	Dir  bool
	Len  int64
}

func (f *FakeFile) Name() string {
	_, name := filepath.Split(f.Path)
	return name
}

func (f *FakeFile) Mode() os.FileMode {
	mode := os.FileMode(0644)
	if f.Dir {
		return mode | os.ModeDir
	}
	return mode
}

func (f *FakeFile) ModTime() time.Time {
	return time.Unix(0, 0)
}

func (f *FakeFile) Size() int64 {
	return f.Len
}

func (f *FakeFile) IsDir() bool {
	return f.Mode().IsDir()
}

func (f *FakeFile) Sys() interface{} {
	return nil
}

type AssetFile struct {
	*bytes.Reader
	io.Closer
	FakeFile
}

func NewAssetFile(name string, content []byte) *AssetFile {
	return &AssetFile{
		bytes.NewReader(content),
		ioutil.NopCloser(nil),
		FakeFile{name, false, int64(len(content))},
	}
}

func (f *AssetFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}

func (f *AssetFile) Size() int64 {
	return f.FakeFile.Size()
}

func (f *AssetFile) Stat() (os.FileInfo, error) {
	return f, nil
}

type AssetDirectory struct {
	AssetFile
	ChildrenRead int
	Children     []os.FileInfo
}

func NewAssetDirectory(name string, children []string, fs AssetFS) *AssetDirectory {
	fileinfos := make([]os.FileInfo, 0, len(children))
	for _, child := range children {
		_, err := fs.AssetDir(filepath.Join(name, child))
		fileinfos = append(fileinfos, &FakeFile{child, err == nil, 0})
	}
	return &AssetDirectory{
		AssetFile{
			bytes.NewReader(nil),
			ioutil.NopCloser(nil),
			FakeFile{name, true, 0},
		},
		0,
		fileinfos}
}

func (f *AssetDirectory) Readdir(count int) ([]os.FileInfo, error) {
	if count <= 0 {
		return f.Children, nil
	}
	if f.ChildrenRead+count > len(f.Children) {
		count = len(f.Children) - f.ChildrenRead
	}
	rv := f.Children[f.ChildrenRead : f.ChildrenRead+count]
	f.ChildrenRead += count
	return rv, nil
}

func (f *AssetDirectory) Stat() (os.FileInfo, error) {
	return f, nil
}

type AssetFS interface {
	Asset(string) ([]byte, error)
	AssetDir(string) ([]string, error)
	AssetNames() []string
	HttpAsset(string) (http.File, error)
}

type assetFS struct {
	asset      func(string) ([]byte, error)
	assetDir   func(string) ([]string, error)
	assetNames func() []string
	prefix     string
}

func NewAssetFS(afn func(string) ([]byte, error), adfn func(string) ([]string, error), anfn func() []string, prefix string) AssetFS {
	return &assetFS{
		asset:      afn,
		assetDir:   adfn,
		assetNames: anfn,
		prefix:     prefix,
	}
}

func (fs *assetFS) Asset(in string) ([]byte, error) {
	return fs.asset(in)
}

func (fs *assetFS) AssetDir(in string) ([]string, error) {
	return fs.assetDir(in)
}

func (fs *assetFS) AssetNames() []string {
	return fs.assetNames()
}

func (fs *assetFS) HasAsset(requested string) (string, bool) {
	for _, filename := range fs.AssetNames() {
		if path.Base(filename) == requested {
			return filename, true
		}
	}
	return "", false
}

var AssetUnavailable = xrr.NewXrror("Asset %s unavailable").Out

func (fs *assetFS) HttpAsset(requested string) (http.File, error) {
	if hasasset, ok := fs.HasAsset(requested); ok {
		f, err := fs.Open(hasasset)
		return f, err
	}
	return nil, AssetUnavailable(requested)
}

func (fs *assetFS) Open(name string) (http.File, error) {
	name = path.Join(fs.prefix, name)
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if children, err := fs.AssetDir(name); err == nil {
		return NewAssetDirectory(name, children, fs), nil
	}
	b, err := fs.Asset(name)
	if err != nil {
		return nil, err
	}
	return NewAssetFile(name, b), nil
}

type Assets interface {
	GetAsset(string) (http.File, error)
	GetAssetByte(string) ([]byte, error)
	SetAssets(...AssetFS)
	ListAssetFS() []AssetFS
}

func New(af ...AssetFS) Assets {
	as := &assets{
		a: make([]AssetFS, 0),
	}
	as.SetAssets(af...)
	return as
}

type assets struct {
	a []AssetFS
}

func (a *assets) GetAsset(requested string) (http.File, error) {
	for _, x := range a.a {
		f, err := x.HttpAsset(requested)
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

func (a *assets) SetAssets(af ...AssetFS) {
	a.a = append(a.a, af...)
	//spew.Dump(a, af)
}

func (a *assets) ListAssetFS() []AssetFS {
	return a.a
}
