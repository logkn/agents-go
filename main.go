package main

import (
	"fmt"

	"github.com/logkn/agents-go/internal/utils"
)

type Foo struct {
	A int    `json:"a" description:"A description of A"`
	B string `json:"b" description:"A description of B"`
	C []int  `json:"c" description:"A description of C"`
}

type DummyStruct struct {
	A int    `json:"a" description:"A description of A"`
	B string `json:"b" description:"A description of B"`
	C []Foo  `json:"c" description:"A description of C"`
}

func main() {
	schema := utils.GenerateSchema(DummyStruct{})
	dumps := utils.JsonDumps(schema)
	fmt.Println(dumps)
}
