package resources

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"os"
	"time"
	"io/ioutil"
	"path"
	"path/filepath"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name string
	size int64
	mode os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _assets_bin_conf = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x4a\xce\xcf\x4b\x4b\x2c\x2e\x4e\x2d\x51\xb0\x55\x70\x0b\xf2\xf7\x8d\x77\x0c\x0e\x76\x0d\xe1\x82\x0b\x1b\xe2\x10\x37\xc2\x21\x6e\x8c\x43\xdc\x04\x87\xb8\x29\xaa\x38\x20\x00\x00\xff\xff\x47\x15\xd0\x7c\x8f\x00\x00\x00")

func assets_bin_conf_bytes() ([]byte, error) {
	return bindata_read(
		_assets_bin_conf,
		"assets/bin.conf",
	)
}

func assets_bin_conf() (*asset, error) {
	bytes, err := assets_bin_conf_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/bin.conf", size: 143, mode: os.FileMode(436), modTime: time.Unix(1426175400, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _assets_templates_layout_asset_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x34\x8d\xc1\x0e\x82\x30\x0c\x86\xef\x3c\x45\xd9\x7d\xec\x05\xca\x12\xa3\x9e\xf5\xe0\xc5\x13\x19\x50\xb3\xc5\x01\x86\xf6\x42\xc8\xde\xdd\xe1\x62\x2f\x4d\xdb\xaf\xff\x87\xf5\xe5\x76\x7e\x3c\xef\x57\xf0\x32\x45\x5b\xe1\xbf\x91\x1b\x6d\x05\xb9\x50\x82\x44\xb2\xfb\x0e\xcd\x69\xde\x20\x25\x34\x65\x73\x1c\x0b\x51\x6b\x8d\x31\xcc\x6f\x58\x29\xb6\x8a\x65\x8b\xc4\x9e\x48\x14\xf8\x95\x5e\xad\x32\x2c\x4e\xc2\x60\x06\x66\xd3\x3b\x0e\x43\x27\x9e\x26\x6a\xf2\xac\x2c\x9a\xe3\xd5\x6a\x9d\xad\xa6\x68\xb1\x5f\xc6\xcd\x96\xf0\xec\x15\x9a\x3e\xd1\x09\x81\x72\xcc\x24\xdd\xba\x2c\x39\x3a\xa5\xcc\xff\xc0\xea\x1b\x00\x00\xff\xff\x9a\x59\xde\xb0\xc5\x00\x00\x00")

func assets_templates_layout_asset_html_bytes() ([]byte, error) {
	return bindata_read(
		_assets_templates_layout_asset_html,
		"assets/templates/layout_asset.html",
	)
}

func assets_templates_layout_asset_html() (*asset, error) {
	bytes, err := assets_templates_layout_asset_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/templates/layout_asset.html", size: 197, mode: os.FileMode(436), modTime: time.Unix(1426101179, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _assets_templates_test_asset_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x74\x8e\xc1\x0a\x02\x21\x10\x86\xef\xfb\x14\x83\x0f\x50\x2f\xb0\x04\x7b\xf0\x56\x10\xe8\x5d\x04\x27\x12\x5c\x5d\xd6\x29\x8a\x65\xdf\xbd\x19\xf1\x12\x94\x47\xbf\x8f\xef\x9f\x6d\x03\x7c\x11\xe6\x50\x41\x25\xff\x2e\x0f\x72\xbe\x56\xa4\xc3\x9d\xe6\xa4\x60\xdf\x87\x81\x95\x80\xb7\x98\x11\x54\x43\x6e\xf1\x2b\x45\x9f\x5c\xc9\xd8\x0c\xe0\x37\x86\xf8\x3c\x35\x0c\x1d\x03\xe3\xf1\x28\xdf\x52\xe0\x85\x9f\xad\xb5\x14\xfa\x8e\x58\x6d\x2c\x4c\xc6\x68\x0b\x56\x5f\xae\xe7\xc9\xea\x9e\x11\x85\x03\x84\xf3\x92\x3c\xfd\x3b\x47\xac\xbe\x28\xdd\x4f\x00\x00\x00\xff\xff\xd8\x4d\x45\xe3\xe1\x00\x00\x00")

func assets_templates_test_asset_html_bytes() ([]byte, error) {
	return bindata_read(
		_assets_templates_test_asset_html,
		"assets/templates/test_asset.html",
	)
}

func assets_templates_test_asset_html() (*asset, error) {
	bytes, err := assets_templates_test_asset_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/templates/test_asset.html", size: 225, mode: os.FileMode(436), modTime: time.Unix(1426101167, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _assets_css_css_asset_css = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func assets_css_css_asset_css_bytes() ([]byte, error) {
	return bindata_read(
		_assets_css_css_asset_css,
		"assets/css/css_asset.css",
	)
}

func assets_css_css_asset_css() (*asset, error) {
	bytes, err := assets_css_css_asset_css_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "assets/css/css_asset.css", size: 0, mode: os.FileMode(436), modTime: time.Unix(1426617848, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
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
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if (err != nil) {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/bin.conf": assets_bin_conf,
	"assets/templates/layout_asset.html": assets_templates_layout_asset_html,
	"assets/templates/test_asset.html": assets_templates_test_asset_html,
	"assets/css/css_asset.css": assets_css_css_asset_css,
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
func AssetDir(name string) ([]string, error) {
	node := _bintree
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
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() (*asset, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"assets": &_bintree_t{nil, map[string]*_bintree_t{
		"bin.conf": &_bintree_t{assets_bin_conf, map[string]*_bintree_t{
		}},
		"css": &_bintree_t{nil, map[string]*_bintree_t{
			"css_asset.css": &_bintree_t{assets_css_css_asset_css, map[string]*_bintree_t{
			}},
		}},
		"templates": &_bintree_t{nil, map[string]*_bintree_t{
			"layout_asset.html": &_bintree_t{assets_templates_layout_asset_html, map[string]*_bintree_t{
			}},
			"test_asset.html": &_bintree_t{assets_templates_test_asset_html, map[string]*_bintree_t{
			}},
		}},
	}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
        data, err := Asset(name)
        if err != nil {
                return err
        }
        info, err := AssetInfo(name)
        if err != nil {
                return err
        }
        err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
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

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
        children, err := AssetDir(name)
        if err != nil { // File
                return RestoreAsset(dir, name)
        } else { // Dir
                for _, child := range children {
                        err = RestoreAssets(dir, path.Join(name, child))
                        if err != nil {
                                return err
                        }
                }
        }
        return nil
}

func _filePath(dir, name string) string {
        cannonicalName := strings.Replace(name, "\\", "/", -1)
        return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

