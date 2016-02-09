// Code generated by flotilla/asset/pack.
// sources:
// assets/css/css_asset.css
// DO NOT EDIT!

package resources

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsCssCss_assetCss = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xd2\xd7\x52\x08\x49\x2d\x2e\x51\x48\xca\xcc\x4b\x2c\xaa\x54\x48\x2e\x2e\x56\x48\x2c\x2e\x4e\x2d\x51\xd0\xd2\xe7\x4a\xca\x4f\xa9\x54\xa8\x56\x48\x4a\x4c\xce\x4e\x2f\xca\x2f\xcd\x4b\xb1\x52\x48\xca\x01\x72\xac\x15\x92\xf3\x73\xf2\x8b\xac\x14\x8a\xd2\x93\x34\x2c\x0c\x74\x14\x20\x58\xd3\x5a\xa1\x96\x0b\x10\x00\x00\xff\xff\xf2\x8f\x4f\x39\x50\x00\x00\x00")

func assetsCssCss_assetCssBytes() ([]byte, error) {
	return bindataRead(
		_assetsCssCss_assetCss,
		"assets/css/css_asset.css",
	)
}

func assetsCssCss_assetCss() (*asset, error) {
	bytes, err := assetsCssCss_assetCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/css/css_asset.css", size: 80, mode: os.FileMode(436), modTime: time.Unix(1454959722, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var ResourceFS *bindataFS = &bindataFS{
	prefix: "",
	tree:   _bintree,
	data:   _bindata,
}

type bindataFS struct {
	prefix string
	tree   *bintree
	data   bindata
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func (b *bindataFS) Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := b.data[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func (b *bindataFS) MustAsset(name string) []byte {
	a, err := b.Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func (b *bindataFS) AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := b.data[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func (b *bindataFS) AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

type bindata map[string]func() (*asset, error)

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = bindata{
	"assets/css/css_asset.css": assetsCssCss_assetCss,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func (b *bindataFS) AssetDir(name string) ([]string, error) {
	node := b.tree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": &bintree{nil, map[string]*bintree{
		"css": &bintree{nil, map[string]*bintree{
			"css_asset.css": &bintree{assetsCssCss_assetCss, map[string]*bintree{}},
		}},
	}},
}}

func (b *bindataFS) HasAsset(requested string) (string, bool) {
	for _, filename := range b.AssetNames() {
		if path.Base(filename) == requested {
			return filename, true
		}
	}
	return "", false
}

func (b *bindataFS) AssetHttp(requested string) (http.File, error) {
	if has, ok := b.HasAsset(requested); ok {
		f, err := b.open(has)
		return f, err
	}
	return nil, errors.New(fmt.Sprintf("Asset %!s(MISSING) unavailable", requested))
}

func (b *bindataFS) open(name string) (http.File, error) {
	name = path.Join(b.prefix, name)
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if children, err := b.AssetDir(name); err == nil {
		return NewAssetDirectory(name, children, b), nil
	}
	bf, err := b.Asset(name)
	if err != nil {
		return nil, err
	}
	return NewAssetFile(name, bf), nil
}

type AssetDirectory struct {
	AssetFile
	ChildrenRead int
	Children     []os.FileInfo
}

func NewAssetDirectory(name string, children []string, fs *bindataFS) *AssetDirectory {
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

// RestoreAsset restores an asset under the given directory
func (b *bindataFS) RestoreAsset(dir, name string) error {
	data, err := b.Asset(name)
	if err != nil {
		return err
	}
	info, err := b.AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

// RestoreAssets restores an asset under the given directory recursively
func (b *bindataFS) RestoreAssets(dir, name string) error {
	children, err := b.AssetDir(name)
	// File
	if err != nil {
		return b.RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = b.RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}
