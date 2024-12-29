package main

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/suzuki-shunsuke/gen-go-jsonschema/jsonschema"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	if err := jsonschema.Write(&domain.Config{}, "json-schema/cmdx.json"); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	return nil
}
