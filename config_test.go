package config

import (
	"fmt"
	"strings"
	"testing"
)

func Test_ConfigReaderInt(t *testing.T) {
	var str = `
	[service]
	a = 1
	b = test 
	c = true 
	d = false 
	e[] = 1
	e [] = 2
	 e [] = 3
	 f = 9
	 e [] = 4
	`

	config := NewConfigReader()
	if err := config.ReadFromStream(strings.NewReader(str)); err != nil {
		t.Fatal(err)
	}

	a, err := config.Int("service", "a")
	if err != nil {
		t.Fatal(err)
	}

	b, err := config.String("service", "b")
	if err != nil {
		t.Fatal(err)
	}

	c, err := config.Bool("service", "c")
	if err != nil {
		t.Fatal(err)
	}

	e, err := config.ArrayInt("service", "e")
	if err != nil {
		t.Fatal(err)
	}

	bb := config.HasSection("pppp")
	if bb {
		t.Fatal("error")
	}

	bb = config.HasSection("service")
	if !bb {
		t.Fatal("error")
	}

	sections, err := config.SectionOptions("service")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range sections {
		if v == "e" {
			continue
		}
		options, e := config.String("service", v)
		if e != nil {
			t.Fatal(e)
		}

		fmt.Println(options)
	}

	fmt.Println(a, b, c, e)
}
