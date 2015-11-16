package flotilla

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/thrisp/flotilla/xrr"
)

type Store map[string]map[string]*StoreItem

func defaultStore() Store {
	s := make(Store)
	s.addDefault("upload", "size", "10000000")             // bytes
	s.addDefault("secret", "key", "Flotilla;Secret;Key;1") // weak default value
	s.addDefault("session", "cookiename", "session")
	s.addDefault("session", "lifetime", "2629743")
	s.add("static", "directories", workingStatic)
	s.add("template", "directories", workingTemplates)
	return s
}

var NoItem = xrr.NewXrror("Store item (section: %s, key: %s) does not exist.").Out

func (s Store) query(key string) (*StoreItem, error) {
	var sec, seckey string
	base := strings.Split(key, "_")
	if len(base) == 1 {
		sec, seckey = "", base[0]
	} else {
		sec, seckey = base[0], base[1]
	}
	if k, ok := s[strings.ToUpper(sec)]; ok {
		if i, ok := k[seckey]; ok {
			return i, nil
		}
	}
	return nil, NoItem(sec, seckey)
}

func (s Store) String(key string) string {
	if i, err := s.query(key); err == nil {
		return i.String()
	}
	return ""
}

func (s Store) Bool(key string) bool {
	if i, err := s.query(key); err == nil {
		return i.Bool()
	}
	return false
}

func (s Store) Float(key string) float64 {
	if i, err := s.query(key); err == nil {
		return i.Float()
	}
	return 0.0
}

func (s Store) Int(key string) int {
	if i, err := s.query(key); err == nil {
		return i.Int()
	}
	return 0
}

func (s Store) Int64(key string) int64 {
	if i, err := s.query(key); err == nil {
		return i.Int64()
	}
	return -1
}

func (s Store) List(key string) []string {
	if i, err := s.query(key); err == nil {
		return i.List()
	}
	return nil
}

func (s Store) LoadConfFile(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	err = s.parse(reader, filename)
	return err
}

func (s Store) LoadConfByte(b []byte, name string) (err error) {
	reader := bufio.NewReader(bytes.NewReader(b))
	err = s.parse(reader, name)
	return err
}

var StoreParseError = xrr.NewXrror("Store configuration parsing: syntax error at '%s:%d'.").Out

func (s Store) parse(reader *bufio.Reader, filename string) (err error) {
	lineno := 0
	section := ""
	for err == nil {
		l, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineno++
		if len(l) == 0 {
			continue
		}
		line := strings.TrimFunc(string(l), unicode.IsSpace)
		for line[len(line)-1] == '\\' {
			line = line[:len(line)-1]
			l, _, err := reader.ReadLine()
			if err != nil {
				return err
			}
			line += strings.TrimFunc(string(l), unicode.IsSpace)
		}
		section, err = s.parseLine(section, line)
		if err != nil {
			return StoreParseError(filename, lineno)
		}
	}
	return err
}

var (
	regDoubleQuote = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*\"([^\"]*)\"$")
	regSingleQuote = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*'([^']*)'$")
	regNoQuote     = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*([^#;]+)")
	regNoValue     = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*([#;].*)?")
)

func (s Store) parseLine(section, line string) (string, error) {
	if line[0] == '#' || line[0] == ';' {
		return section, nil
	}

	if line[0] == '[' && line[len(line)-1] == ']' {
		section := strings.TrimFunc(line[1:len(line)-1], unicode.IsSpace)
		section = strings.ToLower(section)
		return section, nil
	}

	if m := regDoubleQuote.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], m[0][2])
		return section, nil
	} else if m = regSingleQuote.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], m[0][2])
		return section, nil
	} else if m = regNoQuote.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], strings.TrimFunc(m[0][2], unicode.IsSpace))
		return section, nil
	} else if m = regNoValue.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], "")
		return section, nil
	}
	return section, errors.New("line parse error")
}

func (s Store) add(section, key, value string) {
	sec, seckey := strings.ToUpper(section), strings.ToUpper(key)
	if _, ok := s[sec]; !ok {
		s[sec] = make(map[string]*StoreItem)
	}
	s[sec][seckey] = newItem(value, false)
}

func (s Store) addDefault(section, key, value string) {
	sec, seckey := strings.ToUpper(section), strings.ToUpper(key)
	if _, ok := s[sec]; !ok {
		s[sec] = make(map[string]*StoreItem)
	}
	s[sec][seckey] = newItem(value, true)
}

type StoreItem struct {
	isDefault bool
	Value     string
}

func newItem(value string, isDefault bool) *StoreItem {
	return &StoreItem{isDefault, value}
}

var boolString = map[string]bool{
	"t":     true,
	"true":  true,
	"y":     true,
	"yes":   true,
	"on":    true,
	"1":     true,
	"f":     false,
	"false": false,
	"n":     false,
	"no":    false,
	"off":   false,
	"0":     false,
}

func (i *StoreItem) String() string {
	return i.Value
}

func (i *StoreItem) Bool() bool {
	if value, ok := boolString[strings.ToLower(i.Value)]; ok {
		return value
	}
	return false
}

func (i *StoreItem) Float() float64 {
	if value, err := strconv.ParseFloat(i.Value, 64); err == nil {
		return value
	}
	return 0.0
}

func (i *StoreItem) Int() int {
	if value, err := strconv.Atoi(i.Value); err == nil {
		return value
	}
	return 0
}

func (i *StoreItem) Int64() int64 {
	if value, err := strconv.ParseInt(i.Value, 10, 64); err == nil {
		return value
	}
	return -1
}

func (i *StoreItem) List(l ...string) []string {
	list := strings.Split(i.Value, ",")
	for _, item := range l {
		list = doAdd(item, list)
	}
	i.Value = strings.Join(list, ",")
	return strings.Split(i.Value, ",")
}
