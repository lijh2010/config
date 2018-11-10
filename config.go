package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type ConfigReader struct {
	comment      string
	section_btag string
	section_etag string
	array_tag    string
	sep          string
	sections     map[string]interface{}
	path         string
}

func NewConfigReader() *ConfigReader {
	return &ConfigReader{
		comment:      "#",
		section_btag: "[",
		section_etag: "]",
		array_tag:    "[]",
		sep:          "=",
		sections:     make(map[string]interface{}),
		path:         "",
	}
}

func (this *ConfigReader) SetComment(comment string) {
	this.comment = comment
}

func (this *ConfigReader) SetSectionTag(btag, etag string) {
	this.section_etag = etag
	this.section_btag = btag
}

func (this *ConfigReader) SetArrayTag(tag string) {
	this.array_tag = tag
}

func (this *ConfigReader) SetSep(tag string) {
	this.sep = tag
}

func (this *ConfigReader) Read(path string) error {
	this.path = path
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return this.ReadFromStream(f)
}

func (this *ConfigReader) ReadFromStream(reader io.Reader) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var str = string(b)
	str = strings.Replace(str, "\r\n", "\n", -1)
	str = strings.Replace(str, "\r", "\n", -1)

	var lines = strings.Split(str, "\n")
	var m map[string]interface{}
	var section = ""
	for i := 0; i < len(lines); i++ {
		var line = strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, this.comment) || line == "" {
			continue
		}

		if strings.HasPrefix(line, this.section_btag) {
			var btag = strings.Index(line, this.section_btag)
			var etag = strings.LastIndex(line, this.section_etag)
			if btag == -1 || etag == -1 {
				return errors.New(fmt.Sprintf("config file '%s' error, line no: %d", this.path, i+1))
			}

			if section != "" {
				this.sections[section] = m
			}

			section = strings.TrimSpace(line[btag+len(this.section_btag) : etag])
			m = make(map[string]interface{})

		} else {
			var pairs = strings.SplitN(line, this.sep, 2)
			if len(pairs) != 2 {
				return errors.New(fmt.Sprintf("config file '%s' error, line no: %d", this.path, i+1))
			}

			var index = strings.Index(pairs[0], this.array_tag)
			if index == -1 {
				m[strings.TrimSpace(pairs[0])] = strings.TrimSpace(pairs[1])
			} else {
				var key = strings.TrimSpace(pairs[0][:index])
				if v, found := m[key]; found {
					vv, ok := v.([]string)
					if !ok {
						return errors.New(fmt.Sprintf("config file '%s' error, type assertion []string failed", this.path))
					}

					vv = append(vv, strings.TrimSpace(pairs[1]))
					m[key] = vv
				} else {
					vv := []string{strings.TrimSpace(pairs[1])}
					m[key] = vv
				}
			}
		}
	}

	if section != "" {
		this.sections[section] = m
	}
	return nil
}

func (this *ConfigReader) checkSectionKey(section, key string) (interface{}, bool) {
	var v, found = this.sections[section]
	if !found {
		return nil, false
	}

	var m, ok = v.(map[string]interface{})
	if !ok {
		return nil, false
	}

	var p, f = m[key]
	if !f {
		return nil, false
	}
	return p, true
}

func (this *ConfigReader) Int(section, key string) (int, error) {
	var p, found = this.checkSectionKey(section, key)
	if !found {
		return 0, errors.New(fmt.Sprintf("section '%s' key '%s' not exists", section, key))
	}

	var str, ok = p.(string)
	if !ok {
		return 0, errors.New(fmt.Sprintf("section '%s' key '%s' type assertion string failed", section, key))
	}
	var i, err = strconv.Atoi(str)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("section '%s' key '%s' convert error: '%s'", section, key, err))
	}

	return i, nil
}
func (this *ConfigReader) String(section, key string) (string, error) {
	var p, found = this.checkSectionKey(section, key)
	if !found {
		return "", errors.New(fmt.Sprintf("section '%s' key '%s' not exists", section, key))
	}

	var str, ok = p.(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("section '%s' key '%s' type assertion string failed", section, key))
	}

	return str, nil
}

func (this *ConfigReader) Bool(section, key string) (bool, error) {
	var p, found = this.checkSectionKey(section, key)
	if !found {
		return false, errors.New(fmt.Sprintf("section '%s' key '%s' not exists", section, key))
	}

	var str, ok = p.(string)
	if !ok {
		return false, errors.New(fmt.Sprintf("section '%s' key '%s' type assertion string failed", section, key))
	}

	if str != "true" && str != "false" {
		return false, errors.New(fmt.Sprintf("section '%s' key '%s' is not a bool value(true or false)", section, key))
	}

	return str == "true", nil
}

func (this *ConfigReader) ArrayInt(section, key string) ([]int, error) {
	var p, found = this.checkSectionKey(section, key)
	if !found {
		return nil, errors.New(fmt.Sprintf("section '%s' key '%s' not exists", section, key))
	}

	var a, ok = p.([]string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("section '%s' key '%s' type assertion []string failed", section, key))
	}

	var res = make([]int, 0)
	for _, v := range a {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("section '%s' key '%s' Convert failed", section, key))
		}
		res = append(res, i)
	}
	return res, nil
}

func (this *ConfigReader) ArrayString(section, key string) ([]string, error) {
	var p, found = this.checkSectionKey(section, key)
	if !found {
		return nil, errors.New(fmt.Sprintf("section '%s', key '%s' not exists", section, key))
	}

	var a, ok = p.([]string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("section '%s' key '%s' type assertion []string failed", section, key))
	}

	return a, nil
}

func (this *ConfigReader) HasSection(section string) bool {
	_, found := this.sections[section]
	return found
}

func (this *ConfigReader) SectionOptions(section string) ([]string, error) {
	var p, found = this.sections[section]
	if !found {
		return nil, errors.New(fmt.Sprintf("section '%s' not exists", section))
	}

	var m, ok = p.(map[string]interface{})
	if !ok {
		return nil, errors.New(fmt.Sprintf("section '%s' type assertion map[string]interface{} failed", section))
	}
	var res = make([]string, 0, len(m))
	for k, _ := range m {
		res = append(res, k)
	}

	return res, nil
}

func (this *ConfigReader) MustInt(section, key string, def ...int) int {
	var v, err = this.Int(section, key)
	if err != nil {
		if len(def) == 1 {
			return def[0]
		}

		panic(fmt.Sprintf("section '%s' key '%s' MustInt error: '%s'", section, key, err))
	}
	return v
}

func (this *ConfigReader) MustBool(section, key string, def ...bool) bool {
	var v, err = this.Bool(section, key)
	if err != nil {
		if len(def) == 1 {
			return def[0]
		}
		panic(fmt.Sprintf("section '%s' key '%s' MustBool error: '%s'", section, key, err))
	}
	return v
}

func (this *ConfigReader) MustString(section, key string, def ...string) string {
	var v, err = this.String(section, key)
	if err != nil {
		if len(def) == 1 {
			return def[0]
		}
		panic(fmt.Sprintf("section '%s' key '%s' MustString error: '%s'", section, key, err))
	}

	return v
}

func (this *ConfigReader) MustArrayInt(section, key string, def ...[]int) []int {
	var v, err = this.ArrayInt(section, key)
	if err != nil {
		if len(def) == 1 {
			return def[0]
		}

		panic(fmt.Sprintf("section '%s' key '%s' MustArrayInt error: '%s'", section, key, err))
	}

	return v
}

func (this *ConfigReader) MustArrayString(section, key string, def ...[]string) []string {
	var v, err = this.ArrayString(section, key)
	if err != nil {
		if len(def) == 1 {
			return def[0]
		}
		panic(fmt.Sprintf("section '%s' key '%s' MustArrayString error: '%s'", section, key, err))
	}

	return v
}
